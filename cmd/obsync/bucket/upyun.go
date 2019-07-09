package bucket

import (
	"log"
	"math/rand"
	"time"

	"github.com/mingcheng/obsync.go"
	"github.com/upyun/go-sdk/upyun"
)

type UpyunBucket struct {
	Config obsync.BucketConfig
	Client *upyun.UpYun
}

func (t UpyunBucket) Info() (interface{}, error) {
	return t.Client.Usage()
}

func (t UpyunBucket) Exists(path string) bool {
	if info, err := t.Client.GetInfo(path); err != nil {
		return false
	} else {
		return info.Size > 0
	}
}

func (t UpyunBucket) Put(task obsync.BucketTask) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	if t.Exists(task.Key) {
		log.Printf("%s is exists", task.Key)
		return
	}

	err := t.Client.Put(&upyun.PutObjectConfig{
		Path:      task.Key,
		LocalPath: task.Local,
	})

	if err != nil {
		log.Println(err)
	} else {
		log.Printf("%s is uploaded to UpYun", task.Key)
	}
}

func init() {
	obsync.RegisterBucket("upyun", func(config obsync.BucketConfig) (obsync.Bucket, error) {
		client := upyun.NewUpYun(&upyun.UpYunConfig{
			Bucket:   config.Name,
			Operator: config.Key,
			Password: config.Secret,
		})

		if _, err := client.Usage(); err != nil {
			return nil, err
		}

		return UpyunBucket{
			Client: client,
			Config: config,
		}, nil
	})
}
