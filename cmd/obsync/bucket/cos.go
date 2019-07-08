package bucket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/mingcheng/obsync.go"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type COSBucket struct {
	Config obsync.BucketConfig
	Client *cos.Client
}

func (t COSBucket) Info() (interface{}, error) {
	s, _, err := t.Client.Service.Get(context.Background())
	if err != nil {
		return nil, err
	}

	for _, b := range s.Buckets {
		if b.Name == t.Config.Name {
			return b, nil
		}
	}

	return nil, fmt.Errorf("bucket %s not found", t.Config.Name)
}

func (t COSBucket) Exists(path string) bool {
	if resp, err := t.Client.Object.Head(context.Background(), path, nil); err != nil {
		return false
	} else {
		return resp.StatusCode != http.StatusOK
	}
}

func (t COSBucket) Put(task obsync.BucketTask) {
	fd, err := os.Open(task.Local)
	if err != nil {
		log.Printf("open file with error: %v", err)
	}

	resp, err := t.Client.Object.Put(context.Background(), task.Key, fd, &cos.ObjectPutOptions{})

	if err != nil {
		log.Printf("put file %s with error: %v", task.Key, err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("put file %s with not normal status, code ", resp.StatusCode)
	} else {
		log.Printf("put file %s is finished", task.Key)
	}
}

func init() {
	obsync.RegisterBucket("cos", func(config obsync.BucketConfig) (obsync.Bucket, error) {
		u, _ := url.Parse(config.EndPoint)
		b := &cos.BaseURL{BucketURL: u}
		c := cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretKey: config.Key,
				SecretID:  config.Secret,
			},
		})

		return COSBucket{
			Config: config,
			Client: c,
		}, nil
	})
}
