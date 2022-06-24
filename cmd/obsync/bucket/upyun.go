/**
 * File: upyun.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Tuesday, July 9th 2019, 10:41:02 am
 * Last Modified: Tuesday, April 19th 2022, 2:32:09 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/mingcheng/obsync"
	"github.com/mingcheng/obsync/bucket"
	"github.com/upyun/go-sdk/upyun"
)

// UpyunBucket is a bucket implementation for Upyun
type UpyunBucket struct {
	Config bucket.Config
	Client *upyun.UpYun
}

// OnStart to run when the bucket is started
func (t UpyunBucket) OnStart(ctx context.Context) error {
	return nil
}

// OnStop to stop the bucket
func (t UpyunBucket) OnStop(ctx context.Context) error {
	return nil
}

// Info returns information about the bucket
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

func (t UpyunBucket) Put(task obsync.Task) error {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	if t.Exists(task.Key) {
		errs := fmt.Errorf("%s is exists", task.Key)
		log.Println(errs)
		return errs
	}

	err := t.Client.Put(&upyun.PutObjectConfig{
		Path:      task.Key,
		LocalPath: task.Local,
	})

	if err != nil {
		log.Println(err)
		return err
	} else {
		log.Printf("%s is uploaded to UpYun", task.Key)
		return nil
	}
}

func init() {
	bucket.Register("upyun", func(config bucket.Config) (bucket.Bucket, error) {
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
