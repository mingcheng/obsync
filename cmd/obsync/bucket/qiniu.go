package bucket

import (
	"context"
	"fmt"
	"log"

	"github.com/mingcheng/obsync/bucket"
	"github.com/mingcheng/obsync/internal"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

type putRet struct {
	Key   string
	Hash  string
	Fsize int
}

type QiNiuBucket struct {
	Config bucket.Config
	Mac    *qbox.Mac
}

func (t QiNiuBucket) Info() (interface{}, error) {
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

		return nil, fmt.Errorf("bucket %s not found", t.Config.Name)
	}
}

func (t QiNiuBucket) Exists(path string) bool {
	manager := storage.NewBucketManager(t.Mac, &storage.Config{
		UseHTTPS: true,
	})

	if info, err := manager.Stat(t.Config.Name, path); err != nil {
		return false
	} else {
		return info.Fsize > 0
	}
}

func (t QiNiuBucket) Put(task internal.Task) error {
	formUploader := storage.NewFormUploader(&storage.Config{
		UseHTTPS: true,
	})

	ret := putRet{}
	err := formUploader.PutFile(context.TODO(), &ret, t.UploadToken(task), task.Key, task.Local, &storage.PutExtra{})

	if err != nil {
		log.Printf("put %s with error: %v", task.Key, err)
		return err
	}

	log.Printf("put %s finished, with hash %s", ret.Key, ret.Hash)
	return nil
}

func (t QiNiuBucket) UploadToken(task internal.Task) string {
	putPolicy := storage.PutPolicy{
		Scope:      fmt.Sprintf("%s:%s", t.Config.Name, task.Key),
		Expires:    uint32(t.Config.Timeout),
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize)}`,
	}

	return putPolicy.UploadToken(t.Mac)
}

func init() {
	bucket.Register("qiniu", func(config bucket.Config) (bucket.Bucket, error) {
		return QiNiuBucket{
			Config: config,
			Mac:    qbox.NewMac(config.Key, config.Secret),
		}, nil
	})
}
