/**
 * File: obs.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, June 21st 2019, 11:31:48 am
 * Last Modified: Friday, June 21st 2019, 11:34:38 am
 *
 * http://www.opensource.org/licenses/MIT
 */

package bucket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mingcheng/obsync.go"
	"github.com/mingcheng/obsync.go/obs"
)

// ObsBucket struct for obs client
type OBSBucket struct {
	Client *obs.ObsClient
	Config obsync.BucketConfig
}

// Run a file to obs bucket
func (o *OBSBucket) Put(task obsync.BucketTask) {
	if o.Config.Force || !o.Exists(task.Key) {
		input := &obs.PutFileInput{}
		input.Bucket = o.Config.Name
		input.Key = task.Key
		input.SourceFile = task.Local

		log.Printf("start upload %s to obs", task.Key)
		if output, err := o.Client.PutFile(input); err != nil {
			log.Println(err)
		} else {
			log.Printf("put %s with out error, status code %d", task.Key, output.StatusCode)
		}
	} else {
		log.Printf("%s is exists, ignore", task.Key)
	}
}

// Exists detect object whether exists
func (o *OBSBucket) Exists(path string) bool {
	meta, err := o.Client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: o.Config.Name,
		Key:    path,
	})

	if err != nil {
		return false
	}

	return meta.StatusCode == http.StatusOK
}

func (o *OBSBucket) Info() (interface{}, error) {
	if info, err := o.Client.GetBucketStorageInfo(o.Config.Name); err != nil {
		return nil, err
	} else {
		if info.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("remote obs server does not response ok status")
		} else {
			return info, nil
		}
	}
}

func init() {
	obsync.RegisterBucket("obs", func(config obsync.BucketConfig) (obsync.Bucket, error) {
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
