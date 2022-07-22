/**
 * File: sleep.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, July 6th 2019, 11:19:12 pm
 * Last Modified: Friday, July 22nd 2022, 2:00:15 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/mingcheng/obsync"
)

// TestBucket is a test buckets
type TestBucket struct {
	Config obsync.BucketConfig
}

// Info to get the buckets info
func (r *TestBucket) Info(ctx context.Context) (interface{}, error) {
	return "This is a test buckets", nil
}

// Exists to check if the file exists
func (r *TestBucket) Exists(ctx context.Context, path string) bool {
	return false
}

// Put to put the file to the buckets
func (r *TestBucket) Put(ctx context.Context, path, key string) error {
	log.Debugf("received path [%s] and key [%s]", path, key)
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	return nil
}

// init function to initialize and register the buckets
func init() {
	_ = obsync.RegisterBucketClientFunc("sleep", func(config obsync.BucketConfig) (obsync.BucketClient, error) {
		return &TestBucket{
			Config: config,
		}, nil
	})
}
