package bucket

import (
	"log"
	"math/rand"
	"time"

	"github.com/mingcheng/obsync.go"
)

type TestBucket struct {
	Config obsync.BucketConfig
}

func (t TestBucket) Info() (interface{}, error) {
	return "This is a test bucket", nil
}

func (t TestBucket) Exists(path string) bool {
	return false
}

func (t TestBucket) Put(task obsync.BucketTask) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
}

func init() {
	obsync.RegisterBucket("test", func(config obsync.BucketConfig) (obsync.Bucket, error) {
		log.Printf("init function from TestBucket")
		return &TestBucket{
			Config: config,
		}, nil
	})
}
