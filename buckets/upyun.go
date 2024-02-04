/**
 * File: upyun.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Tuesday, July 9th 2019, 10:41:02 am
 * Last Modified: Friday, July 22nd 2022, 2:00:13 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"fmt"
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

// Exists to check if specified path exists
func (t UpyunBucket) Exists(_ context.Context, path string) bool {
	if info, err := t.Client.GetInfo(path); err != nil {
		return false
	} else {
		return info.Size > 0
	}
}

// Put to update local file to bucket with specified key
func (t UpyunBucket) Put(ctx context.Context, filePath, key string) (err error) {
	if t.Exists(ctx, key) {
		err = fmt.Errorf("%s is exists", key)
		return
	}

	err = t.Client.Put(&upyun.PutObjectConfig{
		Path:      key,
		LocalPath: filePath,
	})

	return
}

func init() {
	_ = obsync.AddBucketSyncFunc("upyun", func(config obsync.BucketConfig) (obsync.BucketSync, error) {
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
