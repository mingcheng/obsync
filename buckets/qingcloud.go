package buckets

import (
	"context"
	"errors"
	"github.com/mingcheng/obsync"
	log "github.com/sirupsen/logrus"
	qingConfig "github.com/yunify/qingstor-sdk-go/config"
	qingService "github.com/yunify/qingstor-sdk-go/service"
	"net/http"
	"os"
	"regexp"
)

// TestBucket is a test buckets
type QingBucket struct {
	Config *obsync.BucketConfig
	bucket *qingService.Bucket
}

// Info to get the buckets info
func (r *QingBucket) Info(_ context.Context) (interface{}, error) {
	statics, err := r.bucket.GetStatistics()
	if err != nil || *statics.StatusCode != http.StatusOK {
		return nil, err
	}

	return statics, nil
}

// Exists to check if the file exists
func (r *QingBucket) Exists(_ context.Context, path string) bool {
	obj, err := r.bucket.GetObject(path, nil)
	if err != nil {
		log.Error(err)
		return false
	}

	log.Tracef("object %v exists", path)
	defer obj.Close()

	return *obj.StatusCode == http.StatusOK
}

// Put to put the file to the buckets
func (r *QingBucket) Put(_ context.Context, path, key string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// start to put the local file to the buckets
	output, err := r.bucket.PutObject(key, &qingService.PutObjectInput{Body: file})
	if err != nil {
		log.Error(err)
		return err
	}

	log.Tracef("%v", output)
	return nil
}

// NewQingClient to instance a new client for qingcloud object-storage service
func NewQingClient(config obsync.BucketConfig) (*QingBucket, error) {
	configuration, err := qingConfig.New(config.Key, config.Secret)
	if err != nil {
		return nil, err
	}

	service, err := qingService.Init(configuration)
	if err != nil {
		return nil, err
	}
	log.Tracef("qingcloud service: %v", service)

	expression := regexp.MustCompile(`://(\w+).(\w+).qingstor.com`)
	if !expression.MatchString(config.EndPoint) {
		return nil, errors.New("invalid end point")
	}

	matches := expression.FindStringSubmatch(config.EndPoint)
	if len(matches) != 3 {
		return nil, errors.New("invalid end point")
	}

	log.Tracef("qingcloud bucket %v region %v", matches[1], matches[2])
	bucket, err := service.Bucket(matches[1], matches[2])
	if err != nil {
		return nil, err
	}

	return &QingBucket{
		bucket: bucket,
		Config: &config,
	}, nil
}

// init function to initialize and register the buckets
func init() {
	_ = obsync.RegisterBucketClientFunc("qingcloud", func(config obsync.BucketConfig) (obsync.BucketClient, error) {
		return NewQingClient(config)
	})
}
