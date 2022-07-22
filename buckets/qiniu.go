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

func (t QiNiuBucket) UploadToken(key string) string {
	putPolicy := storage.PutPolicy{
		Scope:      fmt.Sprintf("%s:%s", t.Config.Name, key),
		Expires:    3600,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize)}`,
	}

	return putPolicy.UploadToken(t.Mac)
}

func init() {
	obsync.RegisterBucketClientFunc("qiniu", func(config obsync.BucketConfig) (obsync.BucketClient, error) {
		return QiNiuBucket{
			Config: &config,
			Mac:    qbox.NewMac(config.Key, config.Secret),
		}, nil
	})
}
