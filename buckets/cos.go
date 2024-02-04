/**
 * File: cos.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Monday, July 8th 2019, 5:07:09 pm
 * Last Modified: Friday, July 22nd 2022, 2:03:33 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/mingcheng/obsync"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type COSBucket struct {
	Config *obsync.BucketConfig
	Client *cos.Client
}

// Info to get bucket information
func (t COSBucket) Info(ctx context.Context) (interface{}, error) {
	s, _, err := t.Client.Service.Get(context.Background())
	if err != nil {
		return nil, err
	}

	for _, b := range s.Buckets {
		if b.Name == t.Config.Name {
			return b, nil
		}
	}

	return nil, fmt.Errorf("buckets %s not found", t.Config.Name)
}

// Exists to check if the path exists in the bucket
func (t COSBucket) Exists(ctx context.Context, path string) bool {
	resp, err := t.Client.Object.Head(ctx, path, nil)
	if err != nil {
		return false
	}

	return resp.StatusCode == http.StatusOK
}

// Put to upload local file to bucket within specified key
func (t COSBucket) Put(ctx context.Context, localFile, key string) error {
	fd, err := os.Open(localFile)
	if err != nil {
		log.Printf("open file with error: %v", err)
		return err
	}

	resp, err := t.Client.Object.Put(ctx, key, fd, &cos.ObjectPutOptions{})
	if err != nil {
		log.Errorf("put file %s with error: %v", key, err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errs := fmt.Errorf("put file %s response status is not correct, status code is %d", key, resp.StatusCode)
		log.Error(errs)
		return errs
	}

	log.Debugf("put file %s is finished", key)
	return err
}

func init() {
	const Name = "cos"
	log.Tracef("start register bucket client which type is %s", Name)
	obsync.AddBucketSyncFunc(Name, func(config obsync.BucketConfig) (obsync.BucketSync, error) {
		u, _ := url.Parse(config.EndPoint)
		b := &cos.BaseURL{BucketURL: u}
		c := cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretKey: config.Key,
				SecretID:  config.Secret,
			},
		})

		return COSBucket{
			Config: &config,
			Client: c,
		}, nil
	})
}
