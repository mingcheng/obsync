package buckets

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/mingcheng/obsync"
	log "github.com/sirupsen/logrus"
	qingConfig "github.com/yunify/qingstor-sdk-go/config"
	qingService "github.com/yunify/qingstor-sdk-go/service"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
)

const MaxChunkSize = 50 * (1 << 20) // 50 mb

// QingBucket is a bucket for qingcloud storage
type QingBucket struct {
	Config *obsync.BucketConfig
	bucket *qingService.Bucket
}

// Info to get the buckets info
func (r *QingBucket) Info(_ context.Context) (interface{}, error) {
	statics, err := r.bucket.GetStatistics()
	if err != nil || qingService.IntValue(statics.StatusCode) != http.StatusOK {
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

	return qingService.IntValue(obj.StatusCode) == http.StatusOK
}

func (r *QingBucket) InitUploadID(key string) (*string, error) {
	output, err := r.bucket.InitiateMultipartUpload(
		key,
		&qingService.InitiateMultipartUploadInput{},
	)

	if err != nil {
		return nil, err
	}

	log.Errorf("The status code expected: 200(actually: %d)", qingService.IntValue(output.StatusCode))
	return output.UploadID, err
}

// PartsCount return a file(from path) can be split to how many parts(depends on chunkSize) and file size.
func (r *QingBucket) PartsCount(path string, chunkSize int) (parts int, size int, err error) {
	file, err := os.Open(path)

	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, 0, err
	}

	sz := fileInfo.Size()
	return int(math.Ceil(float64(sz) / float64(chunkSize))), int(sz), nil
}

// PutDirectly to put the contents directly
func (r *QingBucket) PutDirectly(_ context.Context, path, key string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()

	// start to put the local file to the buckets
	output, err := r.bucket.PutObject(key, &qingService.PutObjectInput{Body: file})
	if err != nil {
		log.Error(err)
		return err
	}

	log.Tracef("the put result code: %d", qingService.IntValue(output.StatusCode))
	return nil
}

func (r *QingBucket) PutMultipart(_ context.Context, filePath, objectKey string) (err error) {
	uploadID, err := r.InitUploadID(objectKey)
	if err != nil {
		return err
	}

	// about to upload the multipart if something is not successful
	defer func() {
		if err != nil {
			log.Error(err)
			log.Warn("abort multipart upload")
			_ = r.AbortMultiUpload(objectKey, uploadID)
		} else {
			log.Debugf("upload %s to %s is successful", filePath, objectKey)
		}
	}()

	partsCount, size, err := r.PartsCount(filePath, MaxChunkSize)
	if err != nil {
		return err
	}
	log.Tracef("upload id %s parts count: %d", qingService.StringValue(uploadID), partsCount)

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	for i := 0; i < partsCount; i++ {
		partSize := int64(math.Min(float64(MaxChunkSize), float64(size-i*MaxChunkSize)))
		partBuffer := make([]byte, partSize)
		_, _ = f.Read(partBuffer)
		partNumber := i
		uploadOutput, err := r.bucket.UploadMultipart(
			objectKey,
			&qingService.UploadMultipartInput{
				UploadID:      uploadID,
				PartNumber:    &partNumber,
				ContentLength: &partSize,
				Body:          bytes.NewReader(partBuffer),
			},
		)

		if err != nil {
			log.Error(err)
		}
		log.Tracef("uploaded part %d of %d, status %d", i, partsCount, qingService.IntValue(uploadOutput.StatusCode))
	}
	_ = f.Close()

	// get all upload parts from bucket
	parts, err := r.ListMultiParts(objectKey, uploadID)
	if err != nil {
		return err
	}
	log.Tracef("list upload parts %d from upload id %s", len(parts), qingService.StringValue(uploadID))

	log.Tracef("start complete upload")
	err = r.CompleteMultiParts(filePath, objectKey, uploadID, parts)
	if err != nil {
		return err
	}

	return nil
}

// ListMultiParts to list multiple parts of a multipart upload request
// https://docs.qingcloud.com/qingstor/api/object/multipart/list_multipart.html
func (r *QingBucket) ListMultiParts(objectKey string, uploadID *string) ([]*qingService.ObjectPartType, error) {
	output, err := r.bucket.ListMultipart(
		objectKey,
		&qingService.ListMultipartInput{
			UploadID: uploadID,
		},
	)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output.ObjectParts, err
}

// CompleteMultiParts to complete multipart upload request with parts
func (r *QingBucket) CompleteMultiParts(filepath string, objectKey string, uploadID *string, parts []*qingService.ObjectPartType) (err error) {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	hash := md5.New()
	if _, err = io.Copy(hash, f); err != nil {
		return err
	}

	_ = f.Close()

	checksum := hex.EncodeToString(hash.Sum(nil))
	_, err = r.bucket.CompleteMultipartUpload(
		objectKey,
		&qingService.CompleteMultipartUploadInput{
			ETag:        &checksum,
			UploadID:    uploadID,
			ObjectParts: parts,
		},
	)

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (r *QingBucket) AbortMultiUpload(objectKey string, uploadID *string) error {
	_, err := r.bucket.AbortMultipartUpload(objectKey, &qingService.AbortMultipartUploadInput{UploadID: uploadID})
	if err != nil {
		return err
	}

	return nil
}

// Put to upload file to bucket
func (r *QingBucket) Put(ctx context.Context, path, key string) error {
	partsCount, _, err := r.PartsCount(path, MaxChunkSize)
	if err != nil {
		return err
	}

	// if partsCount larger than two(more than 5g) then start multipart upload
	if partsCount >= 2 {
		log.Debugf("start multipart upload, upload part %d", partsCount)
		return r.PutMultipart(ctx, path, key)
	}

	log.Debugf("start directory upload")
	return r.PutDirectly(ctx, path, key)
}

// Del to delete a object from the bucket
func (r *QingBucket) Del(_ context.Context, key string) error {
	result, err := r.bucket.DeleteObject(key)
	if err != nil {
		return err
	}

	log.Tracef("%v", *result)
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

	expression := regexp.MustCompile(`://([\w|-]+).([\w+|-]+).qingstor.com`)
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

	client := &QingBucket{
		bucket: bucket,
		Config: &config,
	}

	info, err := client.Info(context.Background())
	if err != nil {
		return nil, err
	}

	// fetch the information from the bucket
	result, ok := info.(*qingService.GetBucketStatisticsOutput)
	if !ok {
		return nil, errors.New("invalid information from bucket")
	}
	log.Tracef("the request bucket %s size is %d", qingService.StringValue(result.Name), result.Size)

	return client, nil
}

// init function to initialize and register the buckets
func init() {
	_ = obsync.AddBucketSyncFunc("qingcloud", func(config obsync.BucketConfig) (obsync.BucketSync, error) {
		return NewQingClient(config)
	})
}
