/**
 * File: obs.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, June 21st 2019, 11:31:48 am
 * Last Modified: Friday, July 22nd 2022, 2:00:42 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package buckets

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/mingcheng/obsync"
)

// OBSBucket struct for obs client
type OBSBucket struct {
	Client *obs.ObsClient
	Config *obsync.BucketConfig
}

// Put a file to obs buckets
func (o *OBSBucket) Put(_ context.Context, localFile, key string) error {
	var input obs.PutFileInput

	input.Bucket = o.Config.Name
	input.Key = key
	input.SourceFile = localFile

	log.Debugf("start upload %s to obs", key)
	output, err := o.Client.PutFile(&input)

	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("put %s with out error, status code %d", key, output.StatusCode)
	return nil
}

// Exists detect object whether exists
func (o *OBSBucket) Exists(_ context.Context, path string) bool {
	if meta, err := o.Client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: o.Config.Name,
		Key:    path,
	}); err != nil {
		return false
	} else {
		return meta.StatusCode == http.StatusOK
	}
}

// Info get obs buckets info
func (o *OBSBucket) Info(_ context.Context) (interface{}, error) {
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
	obsync.AddBucketSyncFunc("obs", func(config obsync.BucketConfig) (obsync.BucketSync, error) {
		client, err := obs.New(config.Key, config.Secret, config.EndPoint)
		if err != nil {
			return nil, err
		}

		return &OBSBucket{
			Client: client,
			Config: &config,
		}, nil
	})
}
