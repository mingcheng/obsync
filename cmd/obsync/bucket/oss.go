/**
 * File: oss.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, July 6th 2019, 9:10:37 pm
 * Last Modified: Tuesday, July 9th 2019, 10:56:59 am
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"context"
	"fmt"
	"log"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/mingcheng/obsync"
	"github.com/mingcheng/obsync/bucket"
)

type OSSBucket struct {
	Client *oss.Client
	Config *bucket.Config
}

func (o *OSSBucket) OnStart(ctx context.Context) error {
	return nil
}

func (o *OSSBucket) OnStop(ctx context.Context) error {
	return nil
}

func (o *OSSBucket) Put(task obsync.Task) error {
	getBucket, err := o.GetBucket()
	if err != nil {
		log.Println(err)
		return err
	}

	if err = getBucket.PutObjectFromFile(task.Key, task.Local); err != nil {
		log.Println(err)
		return err
	}

	log.Printf("upload %s to oss is finished", task.Key)
	return nil
}

func (o *OSSBucket) Info() (interface{}, error) {
	if info, err := o.Client.GetBucketInfo(o.Config.Name); err != nil {
		return nil, err
	} else {
		if info.BucketInfo.Name != o.Config.Name {
			return nil, fmt.Errorf("oss bucket info does not match configured name")
		} else {
			return info, nil
		}
	}
}

func (o *OSSBucket) Exists(path string) bool {
	if getBucket, err := o.GetBucket(); err != nil {
		return false
	} else {
		result, err := getBucket.IsObjectExist(path)
		if err != nil {
			return false
		}
		return result
	}
}

func (o *OSSBucket) GetBucket() (*oss.Bucket, error) {
	return o.Client.Bucket(o.Config.Name)
}

func init() {
	bucket.Register("oss", func(config bucket.Config) (bucket.Bucket, error) {
		client, err := oss.New(config.EndPoint, config.Key, config.Secret)
		if err != nil {
			return nil, err
		}

		return &OSSBucket{
			Client: client,
			Config: &config,
		}, nil
	})
}
