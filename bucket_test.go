package obsync

import (
	"testing"
)

func TestBucketCallback(t *testing.T) {
	var name = "test"
	RegisterBucket(name, func(_ BucketConfig) (Bucket, error) {
		return nil, nil
	})

	if _, err := BucketCallback(name); err != nil {
		t.Error(err)
	}

	_ = UnRegisterBucket(name)

	if callback, err := BucketCallback(name); callback != nil {
		t.Errorf("not cleaned, %v", err)
	}
}

func TestAllRegisteredBucket(t *testing.T) {
	// for _, bucket := range []string{"cos", "obs", "oss", "qiniu", "upyun", "test"} {
	// 	if _, err := BucketCallback(strings.ToLower(bucket)); err != nil {
	// 		t.Error(err)
	// 	}
	// }
}
