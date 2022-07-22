/**
 * File: oss.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, July 6th 2019, 9:10:37 pm
 * Last Modified: Tuesday, July 9th 2019, 10:56:59 am
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/mingcheng/obsync"
)

type OSSBucket struct {
	Client *oss.Client
	Config *obsync.BucketConfig
}

func (o *OSSBucket) Put(_ context.Context, localFile, key string) (err error) {
	var getBucket *oss.Bucket

	getBucket, err = o.GetBucket()
	if err != nil {
		return
	}

	if err = getBucket.PutObjectFromFile(key, localFile); err != nil {
		return
	}

	return
}

func (o *OSSBucket) Info(_ context.Context) (result interface{}, err error) {
	var info oss.GetBucketInfoResult

	info, err = o.Client.GetBucketInfo(o.Config.Name)

	if err != nil {
		return nil, err
	}

	if info.BucketInfo.Name != o.Config.Name {
		return nil, fmt.Errorf("oss buckets info does not match configured name")
	} else {
		return info, nil
	}
}

func (o *OSSBucket) Exists(_ context.Context, path string) bool {
	getBucket, err := o.GetBucket()

	if err != nil {
		return false
	}

	result, err := getBucket.IsObjectExist(path)
	if err != nil {
		return false
	}

	return result
}

func (o *OSSBucket) GetBucket() (*oss.Bucket, error) {
	return o.Client.Bucket(o.Config.Name)
}

func init() {
	obsync.RegisterBucketClientFunc("oss", func(config obsync.BucketConfig) (obsync.BucketClient, error) {
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
