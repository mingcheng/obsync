/**
 * File: cos.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Monday, July 8th 2019, 5:07:09 pm
 * Last Modified: Tuesday, July 9th 2019, 10:50:12 am
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/mingcheng/obsync"
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
		return resp.StatusCode == http.StatusOK
	}
}

func (t COSBucket) Put(task obsync.BucketTask) error {
	fd, err := os.Open(task.Local)
	if err != nil {
		log.Printf("open file with error: %v", err)
		return err
	}

	resp, err := t.Client.Object.Put(context.Background(), task.Key, fd, &cos.ObjectPutOptions{})
	if err != nil {
		log.Printf("put file %s with error: %v", task.Key, err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errs := fmt.Errorf("put file %v with not normal status, code %v", task.Key, resp.StatusCode)
		log.Print(errs)
		return errs
	}

	log.Printf("put file %s is finished", task.Key)
	return err
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
