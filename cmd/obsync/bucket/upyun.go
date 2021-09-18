/**
 * File: upyun.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Tuesday, July 9th 2019, 10:41:02 am
 * Last Modified: Tuesday, July 9th 2019, 10:50:29 am
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/mingcheng/obsync/bucket"
	"github.com/mingcheng/obsync/internal"
	"github.com/upyun/go-sdk/upyun"
)

type UpyunBucket struct {
	Config bucket.Config
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

func (t UpyunBucket) Put(task internal.Task) error {
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
