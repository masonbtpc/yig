package meta

import "github.com/journeymidnight/yig/meta/types"

func LifeCycleFromBucket(b types.Bucket) (lc types.LifeCycle) {
	lc.BucketName = b.Name
	lc.Status = "Pending"
	return
}

func (m *Meta) PutBucketToLifeCycle(bucket types.Bucket) error {
	lifeCycle := LifeCycleFromBucket(bucket)
	return m.Client.PutBucketToLifeCycle(lifeCycle)
}

func (m *Meta) RemoveBucketFromLifeCycle(bucket types.Bucket) error {
	return m.Client.RemoveBucketFromLifeCycle(bucket)
}

func (m *Meta) ScanLifeCycle(limit int, marker string) (result types.ScanLifeCycleResult, err error) {
	return m.Client.ScanLifeCycle(limit, marker)
}
