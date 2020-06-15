/**
 * File: bucket.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Tuesday, June 18th 2019, 6:27:36 pm
 * Last Modified: Tuesday, June 18th 2019, 7:04:14 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package obsync

import (
	"fmt"
)

// BucketConfig bucket config
type BucketConfig struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Key      string `json:"key"`
	Secret   string `json:"secret"`
	Force    bool   `json:"force"`
	EndPoint string `json:"endpoint"`
	Timeout  uint64 `json:"timeout"`
	Thread   uint64 `json:"thread"`
}

type Bucket interface {
	Info() (interface{}, error)
	Exists(path string) bool
	Put(task BucketTask) error
}

var (
	buckets = make(map[string]func(c BucketConfig) (Bucket, error))
)

// RegisterBucket for registering bucket to local maps
func RegisterBucket(typeName string, f func(c BucketConfig) (Bucket, error)) {
	buckets[typeName] = f
}

// UnRegisterBucket to unregister bucket from local maps
func UnRegisterBucket(typeName string) error {
	if len(typeName) <= 0 || buckets[typeName] == nil {
		return fmt.Errorf("the bucket with name %s is not exists", typeName)
	}
	delete(buckets, typeName)
	return nil
}

func BucketCallback(typeName string) (func(c BucketConfig) (Bucket, error), error) {
	if callback, ok := buckets[typeName]; !ok {
		return nil, fmt.Errorf("err: bucket callback which name %s does not exists", typeName)
	} else {
		return callback, nil
	}
}
