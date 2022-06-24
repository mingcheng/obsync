/**
 * File: test.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, July 6th 2019, 11:19:12 pm
 * Last Modified: Sunday, July 7th 2019, 5:40:21 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/mingcheng/obsync"
	"github.com/mingcheng/obsync/bucket"
)

// TestBucket is a test bucket
type TestBucket struct {
	Config bucket.Config
}

// OnStart called when bucket is started
func (r *TestBucket) OnStart(ctx context.Context) error {
	log.Println("on start")
	return nil
}

// OnStop called when the bucket uploader is stopped
func (r *TestBucket) OnStop(ctx context.Context) error {
	log.Println("on stop")
	return nil
}

// Info to get the bucket info
func (r *TestBucket) Info() (interface{}, error) {
	return "This is a test bucket", nil
}

// Exists to check if the file exists
func (r *TestBucket) Exists(path string) bool {
	return false
}

// Put to put the file to the bucket
func (r *TestBucket) Put(task obsync.Task) error {
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	return nil
}

// init function to initialize and register the bucket
func init() {
	bucket.Register("test", func(config bucket.Config) (bucket.Bucket, error) {
		return &TestBucket{
			Config: config,
		}, nil
	})
}
