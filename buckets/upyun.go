/**
 * File: upyun.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Tuesday, July 9th 2019, 10:41:02 am
 * Last Modified: Tuesday, April 19th 2022, 2:32:09 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"time"

	"github.com/mingcheng/obsync"
	"github.com/upyun/go-sdk/upyun"
)

// UpyunBucket is a buckets implementation for Upyun
type UpyunBucket struct {
	Config *obsync.BucketConfig
	Client *upyun.UpYun
}

// Info returns information about the buckets
func (t UpyunBucket) Info(_ context.Context) (interface{}, error) {
	return t.Client.Usage()
}

func (t UpyunBucket) Exists(_ context.Context, path string) bool {
	if info, err := t.Client.GetInfo(path); err != nil {
		return false
	} else {
		return info.Size > 0
	}
}

func (t UpyunBucket) Put(ctx context.Context, filePath, key string) (err error) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	if t.Exists(ctx, key) {
		err = fmt.Errorf("%s is exists", key)
		log.Error(err)
		return
	}

	err = t.Client.Put(&upyun.PutObjectConfig{
		Path:      key,
		LocalPath: filePath,
	})

	if err != nil {
		log.Error(err)
	}

	return
}

func init() {
	obsync.RegisterBucketClientFunc("upyun", func(config obsync.BucketConfig) (obsync.BucketClient, error) {
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
			Config: &config,
		}, nil
	})
}
