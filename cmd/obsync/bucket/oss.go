/**
 * File: oss.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, July 6th 2019, 9:10:37 pm
 * Last Modified: Saturday, July 6th 2019, 11:19:52 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"fmt"
	"log"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/mingcheng/obsync.go"
)

type OSSBucket struct {
	Client *oss.Client
	Config *obsync.BucketConfig
}

func (o *OSSBucket) Put(task obsync.BucketTask) {
	if bucket, err := o.GetBucket(); err != nil {
		log.Println(err)
	} else {
		err = bucket.PutObjectFromFile(task.Key, task.Local)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("upload %s to oss is finished", task.Key)
		}
	}
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
	if bucket, err := o.GetBucket(); err != nil {
		return false
	} else {
		result, err := bucket.IsObjectExist(path)
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
	obsync.RegisterBucket("oss", func(config obsync.BucketConfig) (obsync.Bucket, error) {
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
