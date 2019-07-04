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
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/mingcheng/obsync.go"
	"github.com/mingcheng/obsync.go/obs"
)

// ObsBucket struct for obs client
type ObsBucket struct {
	ctx      context.Context
	wg       sync.WaitGroup
	taskChan chan obsync.BucketTask
	Client   *obs.ObsClient
	Config   obsync.BucketConfig
	Debug    bool
}

func (u *ObsBucket) Info() (interface{}, error) {
	return u.Client.GetBucketStorageInfo(u.Config.Name)
}

func (u *ObsBucket) RunTasks(tasks []obsync.BucketTask) {
	for _, task := range tasks {
		go u.Put(task)
	}
}

func (u *ObsBucket) Wait() {
	u.wg.Wait()
}

// NewObsBucket to create new obs client
func NewObsBucket(ctx context.Context, config obsync.BucketConfig, debug bool) (*ObsBucket, error) {
	client, err := obs.New(config.Key, config.Secret, config.EndPoint, obs.WithSocketTimeout(int(config.Timeout)))
	if err != nil {
		return nil, err
	}

	o := &ObsBucket{
		taskChan: make(chan obsync.BucketTask, config.Thread),
		ctx:      ctx,
		Client:   client,
		Config:   config,
		Debug:    debug,
	}

	if debug {
		info, _ := o.Info()
		log.Printf("Obs bucket status code is %v", info)
	}

	return o, nil
}

// Put a file to obs bucket
func (b *ObsBucket) Put(task obsync.BucketTask) {
	b.taskChan <- task
	b.wg.Add(1)
	defer b.wg.Done()

	select {
	case <-b.ctx.Done():
		if b.Debug {
			log.Printf("%s is canceled", task.Key)
		}
		return

	default:
		if b.Config.Force || !b.Exists(task.Key) {
			input := &obs.PutFileInput{}
			input.Bucket = b.Config.Name
			input.Key = task.Key
			input.SourceFile = task.Local

			log.Printf("start upload %s to obs", task.Key)
			if output, err := b.Client.PutFile(input); err != nil {
				log.Println(err)
			} else {
				log.Printf("put %s with out error, status code %d", task.Key, output.StatusCode)
			}
		} else {
			log.Printf("%s is exists, ignore", task.Key)
		}

		<-b.taskChan
	}
}

// Exists detect object whether exists
func (u *ObsBucket) Exists(path string) bool {
	meta, err := u.Client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: u.Config.Name,
		Key:    path,
	})

	if err != nil {
		return false
	}

	return meta.StatusCode == http.StatusOK
}
