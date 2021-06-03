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
	"math/rand"
	"time"

	"github.com/mingcheng/obsync"
	"github.com/mingcheng/obsync/bucket"
)

type TestBucket struct {
	Config bucket.Config
}

func (t TestBucket) Info() (interface{}, error) {
	return "This is a test bucket", nil
}

func (t TestBucket) Exists(path string) bool {
	return false
}

func (t TestBucket) Put(task obsync.BucketTask) error {
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	return nil
}

func init() {
	bucket.Register("test", func(config bucket.Config) (bucket.Bucket, error) {
		return TestBucket{
			Config: config,
		}, nil
	})
}
