/**
 * File: sleep.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, July 6th 2019, 11:19:12 pm
 * Last Modified: Sunday, July 7th 2019, 5:40:21 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"github.com/sirupsen/logrus"
	"log"
	"math/rand"
	"time"

	"github.com/mingcheng/obsync"
)

// TestBucket is a test buckets
type TestBucket struct {
	Config obsync.BucketConfig
}

// Start called when buckets is started
func (r *TestBucket) Start(ctx context.Context) error {
	log.Println("on start")
	return nil
}

// Stop OnStop when the buckets uploader is stopped
func (r *TestBucket) Stop(ctx context.Context) error {
	log.Println("on stop")
	return nil
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
	logrus.Trace(path, key)
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
