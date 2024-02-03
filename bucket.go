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

// BucketClient is a client for obsync bucket client interface.
type BucketClient interface {
	Info(context.Context) (interface{}, error)
	Exists(context.Context, string) bool
	Put(cxt context.Context, filePath, key string) error
}

// to store the client callback functions and instance variables
type (
	NewBucketClient  func(config BucketConfig) (BucketClient, error)
	NewBucketClients map[string]NewBucketClient
)

var (
	newBucketClients  = make(NewBucketClients)
	addClientFuncLock sync.Mutex
)

// AllSupportedBucketTypes to get all the registered buckets types.
func AllSupportedBucketTypes() (types []string) {
	for k := range newBucketClients {
		types = append(types, k)
	}

	return types
}

// RegisterBucketClientFunc to register new type of bucket client
func RegisterBucketClientFunc(typeName string, newClientFunc NewBucketClient) (err error) {
	addClientFuncLock.Lock()
	defer addClientFuncLock.Unlock()

	if _, ok := newBucketClients[typeName]; ok {
		return fmt.Errorf("bucket type name is %s already exists", typeName)
	}

	newBucketClients[typeName] = newClientFunc
	return
}

// RemoveBucketClientFunc to unregister newBucketClients from local maps
func RemoveBucketClientFunc(typeName string) error {
	addClientFuncLock.Lock()
	defer addClientFuncLock.Unlock()

	if len(typeName) <= 0 || newBucketClients[typeName] == nil {
		return fmt.Errorf("bucket type name %s does not exists", typeName)
	}
	delete(newBucketClients, typeName)
	return nil
}

// NewBucketClientFunc provide a callback function for creating new buckets function
func NewBucketClientFunc(typeName string) (NewBucketClient, error) {
	callback, ok := newBucketClients[typeName]
	if !ok {
		return nil, fmt.Errorf("bucket type name %s does not exists", typeName)
	}

	return callback, nil
}
