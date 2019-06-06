package main

import (
	"net/http"
	"sync"

	"github.com/mingcheng/obsync.go/obs"
)

var client *obs.ObsClient
var once sync.Once

type ObsTask interface {
	Exists() bool
	Meta() (output *obs.GetObjectMetadataOutput, err error)
	Put() (output *obs.PutObjectOutput, err error)
	Del() (output *obs.DeleteObjectOutput, err error)
	Client() *obs.ObsClient
}

type Obs struct {
	SourceFile string
	RemoteKey  string
	BucketName string
}

// @see http://marcio.io/2015/07/singleton-pattern-in-go/
func Client(ak, sk, endpoint string) *obs.ObsClient {
	once.Do(func() {
		var err error
		client, err = obs.New(ak, sk, endpoint)
		if err != nil {
			panic(err)
		}
	})

	return client
}

func (s *Obs) Put() (output *obs.PutObjectOutput, err error) {
	input := &obs.PutFileInput{}
	input.Bucket = s.BucketName
	input.Key = s.RemoteKey
	input.SourceFile = s.SourceFile

	return client.PutFile(input)
}

func (s *Obs) Del() (output *obs.DeleteObjectOutput, err error) {
	input := &obs.DeleteObjectInput{
		Bucket: s.BucketName,
		Key:    s.RemoteKey,
	}

	return client.DeleteObject(input)
}

func (s *Obs) Exists() bool {
	if output, err := s.Meta(); err != nil {
		return false
	} else {
		return output.StatusCode == http.StatusOK
	}
}

func (s *Obs) Meta() (output *obs.GetObjectMetadataOutput, err error) {
	return client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: s.BucketName,
		Key:    s.RemoteKey,
	})
}

func (s *Obs) Client() *obs.ObsClient {
	return client
}

func (s *Obs) Info() (output *obs.GetBucketStorageInfoOutput, err error) {
	return client.GetBucketStorageInfo(s.BucketName)
}
