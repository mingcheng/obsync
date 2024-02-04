/**
 * File: runner2.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Tuesday, June 18th 2019, 6:27:36 pm
 * Last Modified: Tuesday, June 18th 2019, 7:04:14 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package obsync

import (
	"context"
	"fmt"
	"sync"
)

type BucketConfig struct {
	Type     string `json:"type" yaml:"type"`
	Name     string `json:"name" yaml:"name"`
	Key      string `json:"key" yaml:"key"`
	Secret   string `json:"secret" yaml:"secret"`
	EndPoint string `json:"endpoint" yaml:"endpoint"`
	SubDir   string `json:"subdir" yaml:"subdir"`
	Region   string `json:"region" yaml:"region"`
}

// BucketSync is a client for obsync bucket client interface.
type BucketSync interface {
	Info(context.Context) (interface{}, error)
	Exists(context.Context, string) bool
	Put(cxt context.Context, filePath, key string) error
}

// to store the client callback functions and instance variables
type (
	BucketSyncFunc  func(config BucketConfig) (BucketSync, error)
	BucketSyncFuncs map[string]BucketSyncFunc
)

var (
	bucketSyncFuncsChan = make(BucketSyncFuncs)
	addClientFuncLock   sync.Mutex
)

// AllSupportedBucketTypes to get all the registered buckets types.
func AllSupportedBucketTypes() (types []string) {
	for k := range bucketSyncFuncsChan {
		types = append(types, k)
	}

	return types
}

// AddBucketSyncFunc to register new type of bucket client
func AddBucketSyncFunc(typeName string, newClientFunc BucketSyncFunc) (err error) {
	addClientFuncLock.Lock()
	defer addClientFuncLock.Unlock()

	if _, ok := bucketSyncFuncsChan[typeName]; ok {
		return fmt.Errorf("bucket type name is %s already exists", typeName)
	}

	bucketSyncFuncsChan[typeName] = newClientFunc
	return
}

// RemoveBucketSyncFunc to unregister bucketSyncFuncsChan from local maps
func RemoveBucketSyncFunc(typeName string) error {
	addClientFuncLock.Lock()
	defer addClientFuncLock.Unlock()

	if len(typeName) <= 0 || bucketSyncFuncsChan[typeName] == nil {
		return fmt.Errorf("bucket type name %s does not exists", typeName)
	}
	delete(bucketSyncFuncsChan, typeName)
	return nil
}

// GetBucketSyncFunc provide a callback function for creating new buckets function
func GetBucketSyncFunc(typeName string) (BucketSyncFunc, error) {
	callback, ok := bucketSyncFuncsChan[typeName]
	if !ok {
		return nil, fmt.Errorf("bucket type name %s does not exists", typeName)
	}

	return callback, nil
}
