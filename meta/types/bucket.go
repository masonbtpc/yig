package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/journeymidnight/yig/api/datatype"
	"github.com/journeymidnight/yig/api/datatype/policy"
)

const (
	VersionEnabled   = "Enabled"
	VersionDisabled  = "Disabled"
	VersionSuspended = "Suspended"
)

type Bucket struct {
	Name string
	// Date and time when the bucket was created,
	// should be serialized into format "2006-01-02T15:04:05.000Z"
	CreateTime    time.Time
	OwnerId       string
	CORS          datatype.Cors
	ACL           datatype.Acl
	BucketLogging datatype.BucketLoggingStatus
	Lifecycle     datatype.Lifecycle
	Policy        policy.Policy
	Website       datatype.WebsiteConfiguration
	Encryption    datatype.EncryptionConfiguration
	Versioning    string // actually enum: Disabled/Enabled/Suspended
	Usage         int64
}

func (b *Bucket) String() (s string) {
	s += "Name: " + b.Name + "\t"
	s += "CreateTime: " + b.CreateTime.Format(CREATE_TIME_LAYOUT) + "\t"
	s += "OwnerId: " + b.OwnerId + "\t"
	s += "CORS: " + fmt.Sprintf("%+v", b.CORS) + "\t"
	s += "ACL: " + fmt.Sprintf("%+v", b.ACL) + "\t"
	s += "BucketLogging: " + fmt.Sprintf("%+v", b.BucketLogging) + "\t"
	s += "LifeCycle: " + fmt.Sprintf("%+v", b.Lifecycle) + "\t"
	s += "Policy: " + fmt.Sprintf("%+v", b.Policy) + "\t"
	s += "Website: " + fmt.Sprintf("%+v", b.Website) + "\t"
	s += "Encryption" + fmt.Sprintf("%+v", b.Encryption) + "\t"
	s += "Version: " + b.Versioning + "\t"
	s += "Usage: " + humanize.Bytes(uint64(b.Usage)) + "\t"
	return
}

//Tidb related function
func (b Bucket) GetUpdateSql() (string, []interface{}) {
	acl, _ := json.Marshal(b.ACL)
	cors, _ := json.Marshal(b.CORS)
	logging, _ := json.Marshal(b.BucketLogging)
	lc, _ := json.Marshal(b.Lifecycle)
	bucket_policy, _ := json.Marshal(b.Policy)
	website, _ := json.Marshal(b.Website)
	encryption, _ := json.Marshal(b.Encryption)
	sql := "update buckets set bucketname=$1,acl=$2,policy=$3,cors=$4,logging=$5,lc=$6,website=$7,encryption=$8,uid=$9,versioning=$10 where bucketname=$11"
	args := []interface{}{b.Name, acl, bucket_policy, cors, logging, lc, website, encryption, b.OwnerId, b.Versioning, b.Name}
	return sql, args
}

func (b Bucket) GetCreateSql() (string, []interface{}) {
	acl, _ := json.Marshal(b.ACL)
	cors, _ := json.Marshal(b.CORS)
	logging, _ := json.Marshal(b.BucketLogging)
	lc, _ := json.Marshal(b.Lifecycle)
	bucket_policy, _ := json.Marshal(b.Policy)
	website, _ := json.Marshal(b.Website)
	encryption, _ := json.Marshal(b.Encryption)
	createTime := b.CreateTime.Format(CREATE_TIME_LAYOUT)
	sql := "insert into buckets(bucketname,acl,cors,logging,lc,uid,policy,website,encryption,createtime,usages,versioning) " +
		"values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);"
	args := []interface{}{b.Name, acl, cors, logging, lc, b.OwnerId, bucket_policy, website, encryption, createTime, b.Usage, b.Versioning}
	return sql, args
}
