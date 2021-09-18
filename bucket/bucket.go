/**
 * File: bucket.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Tuesday, June 18th 2019, 6:27:36 pm
 * Last Modified: Tuesday, June 18th 2019, 7:04:14 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"fmt"

	"github.com/mingcheng/obsync"
)

type (
	// Config bucket config
	Config struct {
		Type     string `json:"type"`
		Name     string `json:"name"`
		Key      string `json:"key"`
		Secret   string `json:"secret"`
		Force    bool   `json:"force"`
		EndPoint string `json:"endpoint"`
		Timeout  uint64 `json:"timeout"`
		Thread   uint64 `json:"thread"`
	}

	Bucket interface {
		Info() (interface{}, error)
		Exists(path string) bool
		Put(task obsync.Task) error
	}

	BucketFunc func(c Config) (Bucket, error)
	Buckets    map[string]BucketFunc
)

var (
	buckets = make(Buckets)
)

// Register for registering bucket to local maps
func Register(typeName string, f func(c Config) (Bucket, error)) {
	buckets[typeName] = f
}

// Remove to unregister bucket from local maps
func Remove(typeName string) error {
	if len(typeName) <= 0 || buckets[typeName] == nil {
		return fmt.Errorf("the bucket with name %s is not exists", typeName)
	}
	delete(buckets, typeName)
	return nil
}

// Func provide a bucket callback func
func Func(typeName string) (BucketFunc, error) {
	callback, ok := buckets[typeName]
	if !ok {
		return nil, fmt.Errorf("err: bucket callback which name %s does not exists", typeName)
	}
	return callback, nil
}
