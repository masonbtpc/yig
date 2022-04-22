package types

import "github.com/journeymidnight/yig/helper"

const (
	TIME_LAYOUT_TIDB             = "2006-01-02 15:04:05"
	INITIALIZATION_VECTOR_LENGTH = 16 // 12 bytes is best performance for GCM, but for CTR
	ObjectNameEnding             = ":"
	ObjectNameSeparator          = "\n"
	ObjectNameSmallestStr        = " "
	ResponseNumberOfRows         = 1024
)

var (
	CREATE_TIME_LAYOUT = helper.CONFIG.TimeFormat
	XXTEA_KEY          = []byte("hehehehe")
	SSE_S3_MASTER_KEY  = []byte("hehehehehehehehehehehehehehehehe") // 32 bytes to select AES-256
)
