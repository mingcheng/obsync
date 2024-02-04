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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mingcheng/obsync"
	log "github.com/sirupsen/logrus"
	"os"
)

// MinioBucket is a test buckets
type S3Bucket struct {
	Config *obsync.BucketConfig
	Client *s3.S3
}

// Info to get the buckets info
func (r *S3Bucket) Info(ctx context.Context) (interface{}, error) {
	return r.Client.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(r.Config.Name),
	})
}

// Exists to check if the file exists
func (r *S3Bucket) Exists(ctx context.Context, path string) bool {
	obj, err := r.Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.Config.Name),
		Key:    aws.String(path),
	})

	log.Debugf("get object [%s] from buckets [%s] with error [%v]", path, r.Config.Name, err)
	log.Debugf("object [%v]", obj)
	if err != nil || obj == nil {
		return false
	}

	return true
}

// Put to put the file to the buckets
func (r *S3Bucket) Put(ctx context.Context, localPath, key string) error {
	f, err := os.Open(localPath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer f.Close()

	result, err := r.Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.Config.Name),
		Key:    aws.String(key),
		Body:   aws.ReadSeekCloser(f),
	})
	if err != nil {
		log.Error(err)
		return err
	}

	// check upload results
	if !r.Exists(ctx, key) {
		log.Errorf("put object [%s] to buckets [%s] failed", key, r.Config.Name)
		_, err = r.Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(r.Config.Name),
			Key:    aws.String(key),
		})

		if err != nil {
			log.Errorf("delete uploaded object [%s] with an error %s", key, err)
		}
	}

	log.Debugf("put object [%s] to buckets [%s] with result [%v]", key, r.Config.Name, result)
	return nil
}

// init function to initialize and register the buckets
func init() {
	log.Tracef("register buckets with type name is s3")
	_ = obsync.RegisterBucketClientFunc("s3", func(config obsync.BucketConfig) (obsync.BucketClient, error) {

		sess, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(config.Key, config.Secret, ""),
		})

		if err != nil {
			log.Error(err)
			return nil, err
		}

		c := aws.NewConfig().WithEndpoint(config.EndPoint)
		if config.Region != "" {
			c.Region = aws.String(config.Region)
		} else {
			c.Region = aws.String("auto")
		}
		log.Debugf("region is [%s]", *c.Region)

		svc := s3.New(sess, c)
		_, err = svc.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(config.Name),
		})

		if err != nil {
			log.Error(err)
			return nil, err
		}

		return &S3Bucket{
			Config: &config,
			Client: svc,
		}, nil
	})
}
