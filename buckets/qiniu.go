/**
 * File: qiniu.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, June 24th 2022, 11:09:00 pm
 * Last Modified: Friday, July 22nd 2022, 2:07:24 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"fmt"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	log "github.com/sirupsen/logrus"

	"github.com/mingcheng/obsync"

	"github.com/qiniu/go-sdk/v7/storage"
)

type putRet struct {
	Key   string
	Hash  string
	Fsize int
}

// QiNiuBucket implements buckets.Bucket interface
type QiNiuBucket struct {
	Config *obsync.BucketConfig
	Mac    *qbox.Mac
}

// Info implements form buckets.Bucket interface
func (t QiNiuBucket) Info(ctx context.Context) (interface{}, error) {
	manager := storage.NewBucketManager(t.Mac, &storage.Config{
		UseHTTPS: true,
	})

	if buckets, err := manager.Buckets(true); err != nil {
		return nil, err
	} else {
		for _, name := range buckets {
			if name == t.Config.Name {
				return name, nil
			}
		}

		return nil, fmt.Errorf("buckets %s not found", t.Config.Name)
	}
}

// Exists to check if specified path is existed
func (t QiNiuBucket) Exists(ctx context.Context, path string) bool {
	manager := storage.NewBucketManager(t.Mac, &storage.Config{
		UseHTTPS: true,
	})

	if info, err := manager.Stat(t.Config.Name, path); err != nil {
		return false
	} else {
		return info.Fsize > 0
	}
}

// Put to upload local file into bucket within specified key
func (t QiNiuBucket) Put(ctx context.Context, localFile, key string) (err error) {
	formUploader := storage.NewFormUploader(&storage.Config{
		UseHTTPS: true,
	})

	var ret putRet
	err = formUploader.PutFile(ctx, &ret, t.UploadToken(key), key, localFile, &storage.PutExtra{})

	if err != nil {
		log.Error(err)
	}

	return
}

// UploadToken to update the key from specified scope
func (t QiNiuBucket) UploadToken(key string) string {
	putPolicy := storage.PutPolicy{
		Scope:      fmt.Sprintf("%s:%s", t.Config.Name, key),
		Expires:    3600,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize)}`,
	}

	return putPolicy.UploadToken(t.Mac)
}

func init() {
	log.Trace("register bucket client callback which type `qiniu`")
	obsync.AddBucketSyncFunc("qiniu", func(config obsync.BucketConfig) (obsync.BucketSync, error) {
		return QiNiuBucket{
			Config: &config,
			Mac:    qbox.NewMac(config.Key, config.Secret),
		}, nil
	})
}
