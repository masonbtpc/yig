/*
 * Minio Cloud Storage, (C) 2015 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package minio

import (
	"encoding/hex"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	. "git.letv.cn/yig/yig/minio/datatype"
	"git.letv.cn/yig/yig/signature"
	mux "github.com/gorilla/mux"
)

// supportedGetReqParams - supported request parameters for GET presigned request.
var supportedGetReqParams = map[string]string{
	"response-expires":             "Expires",
	"response-content-type":        "Content-Type",
	"response-cache-control":       "Cache-Control",
	"response-content-disposition": "Content-Disposition",
}

// setGetRespHeaders - set any requested parameters as response headers.
func setGetRespHeaders(w http.ResponseWriter, reqParams url.Values) {
	for k, v := range reqParams {
		if header, ok := supportedGetReqParams[k]; ok {
			w.Header()[header] = v
		}
	}
}

// errAllowableNotFound - For an anon user, return 404 if have ListBucket, 403 otherwise
// this is in keeping with the permissions sections of the docs of both:
//   HEAD Object: http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectHEAD.html
//   GET Object: http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectGET.html
func errAllowableObjectNotFound(bucket string, r *http.Request) APIErrorCode {
	if signature.GetRequestAuthType(r) == signature.AuthTypeAnonymous {
		//we care about the bucket as a whole, not a particular resource
		url := *r.URL
		url.Path = "/" + bucket

		if s3Error := enforceBucketPolicy("s3:ListBucket", bucket, &url); s3Error != ErrNone {
			return ErrAccessDenied
		}
	}
	return ErrNoSuchKey
}

// Simple way to convert a func to io.Writer type.
type funcToWriter func([]byte) (int, error)

func (f funcToWriter) Write(p []byte) (int, error) {
	return f(p)
}

// GetObjectHandler - GET Object
// ----------
// This implementation of the GET operation retrieves object. To use GET,
// you must have READ access to the object.
func (api objectAPIHandlers) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	var object, bucket string
	vars := mux.Vars(r)
	bucket = vars["bucket"]
	object = vars["object"]

	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/using-with-s3-actions.html
		if s3Error := enforceBucketPolicy("s3:GetObject", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4:
		if _, s3Error := signature.IsReqAuthenticated(r); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	}
	// Fetch object stat info.
	objInfo, err := api.ObjectAPI.GetObjectInfo(bucket, object)
	if err != nil {
		errorIf(err, "Unable to fetch object info.")
		apiErr := ToAPIErrorCode(err)
		if apiErr == ErrNoSuchKey {
			apiErr = errAllowableObjectNotFound(bucket, r)
		}
		WriteErrorResponse(w, r, apiErr, r.URL.Path)
		return
	}

	// Get request range.
	var hrange *HttpRange
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		if hrange, err = ParseRequestRange(rangeHeader, objInfo.Size); err != nil {
			// Handle only ErrorInvalidRange
			// Ignore other parse error and treat it as regular Get request like Amazon S3.
			if err == ErrorInvalidRange {
				WriteErrorResponse(w, r, ErrInvalidRange, r.URL.Path)
				return
			}

			// log the error.
			errorIf(err, "Invalid request range")
		}

	}

	// Validate pre-conditions if any.
	if checkPreconditions(w, r, objInfo) {
		return
	}

	// Get the object.
	startOffset := int64(0)
	length := objInfo.Size
	if hrange != nil {
		startOffset = hrange.OffsetBegin
		length = hrange.GetLength()
	}
	// Indicates if any data was written to the http.ResponseWriter
	dataWritten := false
	// io.Writer type which keeps track if any data was written.
	writer := funcToWriter(func(p []byte) (int, error) {
		if !dataWritten {
			// Set headers on the first write.
			// Set standard object headers.
			SetObjectHeaders(w, objInfo, hrange)

			// Set any additional requested response headers.
			setGetRespHeaders(w, r.URL.Query())

			dataWritten = true
		}
		return w.Write(p)
	})
	// Reads the object at startOffset and writes to mw.
	if err := api.ObjectAPI.GetObject(objInfo, startOffset, length, writer); err != nil {
		errorIf(err, "Unable to write to client.")
		if !dataWritten {
			// Error response only if no data has been written to client yet. i.e if
			// partial data has already been written before an error
			// occurred then no point in setting StatusCode and
			// sending error XML.
			apiErr := ToAPIErrorCode(err)
			WriteErrorResponse(w, r, apiErr, r.URL.Path)
		}
		return
	}
	if !dataWritten {
		// If ObjectAPI.GetObject did not return error and no data has
		// been written it would mean that it is a 0-byte object.
		// call wrter.Write(nil) to set appropriate headers.
		writer.Write(nil)
	}
}

// HeadObjectHandler - HEAD Object
// -----------
// The HEAD operation retrieves metadata from an object without returning the object itself.
func (api objectAPIHandlers) HeadObjectHandler(w http.ResponseWriter, r *http.Request) {
	var object, bucket string
	vars := mux.Vars(r)
	bucket = vars["bucket"]
	object = vars["object"]

	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/using-with-s3-actions.html
		if s3Error := enforceBucketPolicy("s3:GetObject", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4:
		if _, s3Error := signature.IsReqAuthenticated(r); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	}

	objInfo, err := api.ObjectAPI.GetObjectInfo(bucket, object)
	if err != nil {
		errorIf(err, "Unable to fetch object info.")
		apiErr := ToAPIErrorCode(err)
		if apiErr == ErrNoSuchKey {
			apiErr = errAllowableObjectNotFound(bucket, r)
		}
		WriteErrorResponse(w, r, apiErr, r.URL.Path)
		return
	}

	// Validate pre-conditions if any.
	if checkPreconditions(w, r, objInfo) {
		return
	}

	// Set standard object headers.
	SetObjectHeaders(w, objInfo, nil)

	// Successfull response.
	w.WriteHeader(http.StatusOK)
}

// CopyObjectHandler - Copy Object
// ----------
// This implementation of the PUT operation adds an object to a bucket
// while reading the object from another source.
func (api objectAPIHandlers) CopyObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/using-with-s3-actions.html
		if s3Error := enforceBucketPolicy("s3:PutObject", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4:
		if _, s3Error := signature.IsReqAuthenticated(r); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	}

	// TODO: Reject requests where body/payload is present, for now we don't even read it.

	// objectSource
	objectSource, err := url.QueryUnescape(r.Header.Get("X-Amz-Copy-Source"))
	if err != nil {
		// Save unescaped string as is.
		objectSource = r.Header.Get("X-Amz-Copy-Source")
	}

	// Skip the first element if it is '/', split the rest.
	if strings.HasPrefix(objectSource, "/") {
		objectSource = objectSource[1:]
	}
	splits := strings.SplitN(objectSource, "/", 2)

	// Save sourceBucket and sourceObject extracted from url Path.
	var sourceBucket, sourceObject string
	if len(splits) == 2 {
		sourceBucket = splits[0]
		sourceObject = splits[1]
	}
	// If source object is empty, reply back error.
	if sourceObject == "" {
		WriteErrorResponse(w, r, ErrInvalidCopySource, r.URL.Path)
		return
	}

	// Source and destination objects cannot be same, reply back error.
	if sourceObject == object && sourceBucket == bucket {
		WriteErrorResponse(w, r, ErrInvalidCopyDest, r.URL.Path)
		return
	}

	objInfo, err := api.ObjectAPI.GetObjectInfo(sourceBucket, sourceObject)
	if err != nil {
		errorIf(err, "Unable to fetch object info.")
		WriteErrorResponse(w, r, ToAPIErrorCode(err), objectSource)
		return
	}

	// Verify before x-amz-copy-source preconditions before continuing with CopyObject.
	if checkCopyObjectPreconditions(w, r, objInfo) {
		return
	}

	/// maximum Upload size for object in a single CopyObject operation.
	if isMaxObjectSize(objInfo.Size) {
		WriteErrorResponse(w, r, ErrEntityTooLarge, objectSource)
		return
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		startOffset := int64(0) // Read the whole file.
		// Get the object.
		gErr := api.ObjectAPI.GetObject(objInfo, startOffset, objInfo.Size, pipeWriter)
		if gErr != nil {
			errorIf(gErr, "Unable to read an object.")
			pipeWriter.CloseWithError(gErr)
			return
		}
		pipeWriter.Close() // Close.
	}()

	// Size of object.
	size := objInfo.Size

	// Save metadata.
	metadata := make(map[string]string)
	// Save other metadata if available.
	metadata["content-type"] = objInfo.ContentType
	metadata["content-encoding"] = objInfo.ContentEncoding
	// Do not set `md5sum` as CopyObject will not keep the
	// same md5sum as the source.

	// Create the object.
	md5Sum, err := api.ObjectAPI.PutObject(bucket, object, size, pipeReader, metadata)
	if err != nil {
		errorIf(err, "Unable to create an object.")
		WriteErrorResponse(w, r, ToAPIErrorCode(err), r.URL.Path)
		return
	}

	objInfo, err = api.ObjectAPI.GetObjectInfo(bucket, object)
	if err != nil {
		errorIf(err, "Unable to fetch object info.")
		WriteErrorResponse(w, r, ToAPIErrorCode(err), r.URL.Path)
		return
	}

	response := GenerateCopyObjectResponse(md5Sum, objInfo.ModTime)
	encodedSuccessResponse := EncodeResponse(response)
	// write headers
	SetCommonHeaders(w)
	// write success response.
	WriteSuccessResponse(w, encodedSuccessResponse)
	// Explicitly close the reader, to avoid fd leaks.
	pipeReader.Close()
}

// PutObjectHandler - PUT Object
// ----------
// This implementation of the PUT operation adds an object to a bucket.
func (api objectAPIHandlers) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	// If the matching failed, it means that the X-Amz-Copy-Source was
	// wrong, fail right here.
	if _, ok := r.Header["X-Amz-Copy-Source"]; ok {
		WriteErrorResponse(w, r, ErrInvalidCopySource, r.URL.Path)
		return
	}
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	// Get Content-Md5 sent by client and verify if valid
	md5Bytes, err := checkValidMD5(r.Header.Get("Content-Md5"))
	if err != nil {
		errorIf(err, "Unable to validate content-md5 format.")
		WriteErrorResponse(w, r, ErrInvalidDigest, r.URL.Path)
		return
	}
	/// if Content-Length is unknown/missing, deny the request
	size := r.ContentLength
	if size == -1 && !contains(r.TransferEncoding, "chunked") {
		WriteErrorResponse(w, r, ErrMissingContentLength, r.URL.Path)
		return
	}
	/// maximum Upload size for objects in a single operation
	if isMaxObjectSize(size) {
		WriteErrorResponse(w, r, ErrEntityTooLarge, r.URL.Path)
		return
	}

	// Save metadata.
	metadata := extractMetadataFromHeader(r.Header)
	metadata["md5Sum"] = hex.EncodeToString(md5Bytes)

	var md5Sum string
	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/using-with-s3-actions.html
		if s3Error := enforceBucketPolicy("s3:PutObject", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
		// Create anonymous object.
		md5Sum, err = api.ObjectAPI.PutObject(bucket, object, size, r.Body, metadata)
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4,
		signature.AuthTypePresignedV2, signature.AuthTypeSignedV2:
		// Initialize signature verifier.
		reader := signature.NewSignVerify(r)
		// Create object.
		md5Sum, err = api.ObjectAPI.PutObject(bucket, object, size, reader, metadata)
	}
	if err != nil {
		errorIf(err, "Unable to create an object.")
		WriteErrorResponse(w, r, ToAPIErrorCode(err), r.URL.Path)
		return
	}
	if md5Sum != "" {
		w.Header().Set("ETag", "\""+md5Sum+"\"")
	}
	WriteSuccessResponse(w, nil)
}

/// Multipart objectAPIHandlers

// NewMultipartUploadHandler - New multipart upload
func (api objectAPIHandlers) NewMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	var object, bucket string
	vars := mux.Vars(r)
	bucket = vars["bucket"]
	object = vars["object"]

	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/mpuAndPermissions.html
		if s3Error := enforceBucketPolicy("s3:PutObject", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4:
		if _, s3Error := signature.IsReqAuthenticated(r); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	}

	// Save metadata.
	metadata := make(map[string]string)
	// Save other metadata if available.
	metadata["content-type"] = r.Header.Get("Content-Type")
	metadata["content-encoding"] = r.Header.Get("Content-Encoding")
	for key := range r.Header {
		cKey := http.CanonicalHeaderKey(key)
		if strings.HasPrefix(cKey, "x-amz-meta-") {
			metadata[cKey] = r.Header.Get(cKey)
		} else if strings.HasPrefix(key, "x-minio-meta-") {
			metadata[cKey] = r.Header.Get(cKey)
		}
	}

	uploadID, err := api.ObjectAPI.NewMultipartUpload(bucket, object, metadata)
	if err != nil {
		errorIf(err, "Unable to initiate new multipart upload id.")
		WriteErrorResponse(w, r, ToAPIErrorCode(err), r.URL.Path)
		return
	}

	response := GenerateInitiateMultipartUploadResponse(bucket, object, uploadID)
	encodedSuccessResponse := EncodeResponse(response)
	// write headers
	SetCommonHeaders(w)
	// write success response.
	WriteSuccessResponse(w, encodedSuccessResponse)
}

// PutObjectPartHandler - Upload part
func (api objectAPIHandlers) PutObjectPartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	// get Content-Md5 sent by client and verify if valid
	md5Bytes, err := checkValidMD5(r.Header.Get("Content-Md5"))
	if err != nil {
		WriteErrorResponse(w, r, ErrInvalidDigest, r.URL.Path)
		return
	}

	/// if Content-Length is unknown/missing, throw away
	size := r.ContentLength
	if size == -1 {
		WriteErrorResponse(w, r, ErrMissingContentLength, r.URL.Path)
		return
	}

	/// maximum Upload size for multipart objects in a single operation
	if isMaxObjectSize(size) {
		WriteErrorResponse(w, r, ErrEntityTooLarge, r.URL.Path)
		return
	}

	uploadID := r.URL.Query().Get("uploadId")
	partIDString := r.URL.Query().Get("partNumber")

	partID, err := strconv.Atoi(partIDString)
	if err != nil {
		WriteErrorResponse(w, r, ErrInvalidPart, r.URL.Path)
		return
	}

	// check partID with maximum part ID for multipart objects
	if isMaxPartID(partID) {
		WriteErrorResponse(w, r, ErrInvalidMaxParts, r.URL.Path)
		return
	}

	var partMD5 string
	incomingMD5 := hex.EncodeToString(md5Bytes)
	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/mpuAndPermissions.html
		if s3Error := enforceBucketPolicy("s3:PutObject", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
		// No need to verify signature, anonymous request access is already allowed.
		partMD5, err = api.ObjectAPI.PutObjectPart(bucket, object, uploadID, partID, size, r.Body, incomingMD5)
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4:
		// Initialize signature verifier.
		reader := signature.NewSignVerify(r)
		partMD5, err = api.ObjectAPI.PutObjectPart(bucket, object, uploadID, partID, size, reader, incomingMD5)
	}
	if err != nil {
		errorIf(err, "Unable to create object part.")
		// Verify if the underlying error is signature mismatch.
		WriteErrorResponse(w, r, ToAPIErrorCode(err), r.URL.Path)
		return
	}
	if partMD5 != "" {
		w.Header().Set("ETag", "\""+partMD5+"\"")
	}
	WriteSuccessResponse(w, nil)
}

// AbortMultipartUploadHandler - Abort multipart upload
func (api objectAPIHandlers) AbortMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/mpuAndPermissions.html
		if s3Error := enforceBucketPolicy("s3:AbortMultipartUpload", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4:
		if _, s3Error := signature.IsReqAuthenticated(r); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	}

	uploadID, _, _, _ := getObjectResources(r.URL.Query())
	if err := api.ObjectAPI.AbortMultipartUpload(bucket, object, uploadID); err != nil {
		errorIf(err, "Unable to abort multipart upload.")
		WriteErrorResponse(w, r, ToAPIErrorCode(err), r.URL.Path)
		return
	}
	WriteSuccessNoContent(w)
}

// Send whitespace character, once every 5secs, until CompleteMultipartUpload is done.
// CompleteMultipartUpload method of the object layer indicates that it's done via doneCh
func sendWhiteSpaceChars(w http.ResponseWriter, doneCh <-chan struct{}) {
	for {
		select {
		case <-time.After(5 * time.Second):
			w.Write([]byte(" "))
		case <-doneCh:
			return
		}
	}
}

// ListObjectPartsHandler - List object parts
func (api objectAPIHandlers) ListObjectPartsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/mpuAndPermissions.html
		if s3Error := enforceBucketPolicy("s3:ListMultipartUploadParts", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4:
		if _, s3Error := signature.IsReqAuthenticated(r); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	}

	uploadID, partNumberMarker, maxParts, _ := getObjectResources(r.URL.Query())
	if partNumberMarker < 0 {
		WriteErrorResponse(w, r, ErrInvalidPartNumberMarker, r.URL.Path)
		return
	}
	if maxParts < 0 {
		WriteErrorResponse(w, r, ErrInvalidMaxParts, r.URL.Path)
		return
	}
	listPartsInfo, err := api.ObjectAPI.ListObjectParts(bucket, object, uploadID, partNumberMarker, maxParts)
	if err != nil {
		errorIf(err, "Unable to list uploaded parts.")
		WriteErrorResponse(w, r, ToAPIErrorCode(err), r.URL.Path)
		return
	}
	response := GenerateListPartsResponse(listPartsInfo)
	encodedSuccessResponse := EncodeResponse(response)
	// Write headers.
	SetCommonHeaders(w)
	// Write success response.
	WriteSuccessResponse(w, encodedSuccessResponse)
}

// CompleteMultipartUploadHandler - Complete multipart upload
func (api objectAPIHandlers) CompleteMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	// Get upload id.
	uploadID, _, _, _ := getObjectResources(r.URL.Query())

	var md5Sum string
	var err error
	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/mpuAndPermissions.html
		if s3Error := enforceBucketPolicy("s3:PutObject", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	case signature.AuthTypePresignedV4, signature.AuthTypeSignedV4:
		if _, s3Error := signature.IsReqAuthenticated(r); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	}
	completeMultipartBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorIf(err, "Unable to complete multipart upload.")
		WriteErrorResponse(w, r, ErrInternalError, r.URL.Path)
		return
	}
	complMultipartUpload := &CompleteMultipartUpload{}
	if err = xml.Unmarshal(completeMultipartBytes, complMultipartUpload); err != nil {
		errorIf(err, "Unable to parse complete multipart upload XML.")
		WriteErrorResponse(w, r, ErrMalformedXML, r.URL.Path)
		return
	}
	if len(complMultipartUpload.Parts) == 0 {
		WriteErrorResponse(w, r, ErrMalformedXML, r.URL.Path)
		return
	}
	if !sort.IsSorted(CompletedParts(complMultipartUpload.Parts)) {
		WriteErrorResponse(w, r, ErrInvalidPartOrder, r.URL.Path)
		return
	}
	// Complete parts.
	var completeParts []CompletePart
	for _, part := range complMultipartUpload.Parts {
		part.ETag = strings.TrimPrefix(part.ETag, "\"")
		part.ETag = strings.TrimSuffix(part.ETag, "\"")
		completeParts = append(completeParts, part)
	}
	// Complete multipart upload.
	// Send 200 OK
	SetCommonHeaders(w)
	w.WriteHeader(http.StatusOK)
	// Xml headers need to be sent before we possibly send whitespace characters
	// to the client.
	_, err = w.Write([]byte(xml.Header))
	if err != nil {
		errorIf(err, "Unable to write XML header for complete multipart upload")
		WriteErrorResponseNoHeader(w, r, ErrInternalError, r.URL.Path)
		return
	}

	doneCh := make(chan struct{})
	// Signal that completeMultipartUpload is over via doneCh
	go func(doneCh chan<- struct{}) {
		md5Sum, err = api.ObjectAPI.CompleteMultipartUpload(bucket, object, uploadID, completeParts)
		doneCh <- struct{}{}
	}(doneCh)

	sendWhiteSpaceChars(w, doneCh)

	if err != nil {
		errorIf(err, "Unable to complete multipart upload.")
		switch oErr := err.(type) {
		case PartTooSmall:
			// Write part too small error.
			writePartSmallErrorResponse(w, r, oErr)
		default:
			// Handle all other generic issues.
			WriteErrorResponseNoHeader(w, r, ToAPIErrorCode(err), r.URL.Path)
		}
		return
	}

	// Get object location.
	location := GetLocation(r)
	// Generate complete multipart response.
	response := GenerateCompleteMultpartUploadResponse(bucket, object, location, md5Sum)
	encodedSuccessResponse, err := xml.Marshal(response)
	if err != nil {
		errorIf(err, "Unable to parse CompleteMultipartUpload response")
		WriteErrorResponseNoHeader(w, r, ErrInternalError, r.URL.Path)
		return
	}
	// write success response.
	w.Write(encodedSuccessResponse)
	w.(http.Flusher).Flush()
}

/// Delete objectAPIHandlers

// DeleteObjectHandler - delete an object
func (api objectAPIHandlers) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	switch signature.GetRequestAuthType(r) {
	default:
		// For all unknown auth types return error.
		WriteErrorResponse(w, r, ErrAccessDenied, r.URL.Path)
		return
	case signature.AuthTypeAnonymous:
		// http://docs.aws.amazon.com/AmazonS3/latest/dev/using-with-s3-actions.html
		if s3Error := enforceBucketPolicy("s3:DeleteObject", bucket, r.URL); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	case signature.AuthTypeSignedV4, signature.AuthTypePresignedV4:
		if _, s3Error := signature.IsReqAuthenticated(r); s3Error != ErrNone {
			WriteErrorResponse(w, r, s3Error, r.URL.Path)
			return
		}
	}
	/// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectDELETE.html
	/// Ignore delete object errors, since we are suppposed to reply
	/// only 204.
	api.ObjectAPI.DeleteObject(bucket, object)
	WriteSuccessNoContent(w)
}
