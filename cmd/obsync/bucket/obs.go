/**
 * File: obs.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, June 21st 2019, 11:31:48 am
 * Last Modified: Saturday, July 6th 2019, 11:27:19 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/mingcheng/obsync"
	"github.com/mingcheng/obsync/bucket"
)

// OBSBucket struct for obs client
type OBSBucket struct {
	Client *obs.ObsClient
	Config bucket.Config
}

func (o *OBSBucket) OnStart(ctx context.Context) error {
	return nil
}

func (o *OBSBucket) OnStop(ctx context.Context) error {
	return nil
}

// Put a file to obs bucket
func (o *OBSBucket) Put(task obsync.Task) error {
	input := &obs.PutFileInput{}
	input.Bucket = o.Config.Name
	input.Key = task.Key
	input.SourceFile = task.Local

	log.Printf("start upload %s to obs", task.Key)
	output, err := o.Client.PutFile(input)

	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("put %s with out error, status code %d", task.Key, output.StatusCode)
	return nil
}

// Exists detect object whether exists
func (o *OBSBucket) Exists(path string) bool {
	if meta, err := o.Client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: o.Config.Name,
		Key:    path,
	}); err != nil {
		return false
	} else {
		return meta.StatusCode == http.StatusOK
	}
}

// Info get obs bucket info
func (o *OBSBucket) Info() (interface{}, error) {
	info, err := o.Client.GetBucketStorageInfo(o.Config.Name)
	if err != nil {
		return nil, err
	}

	if info.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote obs server does not response ok status")
	}

	return info, nil
}

func init() {
	bucket.Register("obs", func(config bucket.Config) (bucket.Bucket, error) {
		client, err := obs.New(config.Key, config.Secret, config.EndPoint, obs.WithSocketTimeout(int(config.Timeout)))
		if err != nil {
			return nil, err
		}

		return &OBSBucket{
			Client: client,
			Config: config,
		}, nil
	})
}
