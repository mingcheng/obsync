/**
 * File: sleep.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, July 6th 2019, 11:19:12 pm
 * Last Modified: Friday, July 22nd 2022, 2:00:15 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"github.com/mingcheng/obsync"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

// S3Bucket is a test buckets
type S3Bucket struct {
	Config *obsync.BucketConfig
	Client *minio.Client
}

// Info to get the buckets info
func (r *S3Bucket) Info(ctx context.Context) (interface{}, error) {
	return r.Client.GetBucketPolicy(context.Background(), r.Config.Name)
}

// Exists to check if the file exists
func (r *S3Bucket) Exists(ctx context.Context, path string) bool {
	_, err := r.Client.GetObject(ctx, r.Config.Name, path, minio.GetObjectOptions{})
	return err != nil
}

// Put to put the file to the buckets
func (r *S3Bucket) Put(ctx context.Context, localPath, key string) error {
	info, err := r.Client.FPutObject(ctx, r.Config.Name, key, localPath, minio.PutObjectOptions{})
	if err != nil {
		log.Error(err)
		return err
	}

	log.Tracef("put to buckets [%s] within key [%s] is finished", r.Config.Name, info.Key)
	return nil
}

// init function to initialize and register the buckets
func init() {
	log.Tracef("register buckets with type name is s3")
	_ = obsync.RegisterBucketClientFunc("s3", func(config obsync.BucketConfig) (obsync.BucketClient, error) {

		minioClient, err := minio.New(config.EndPoint, &minio.Options{
			Creds:  credentials.NewStaticV4(config.Key, config.Secret, ""),
			Secure: false,
		})

		if err != nil {
			log.Error(err)
			return nil, err
		}

		found, err := minioClient.BucketExists(context.Background(), config.Name)
		if err != nil || !found {
			log.Error(err)
			return nil, err
		}

		return &S3Bucket{
			Config: &config,
			Client: minioClient,
		}, nil
	})
}