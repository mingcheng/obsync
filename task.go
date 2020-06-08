/**
 * File: task.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, July 6th 2019, 10:56:26 pm
 * Last Modified: Sunday, July 7th 2019, 7:05:57 am
 *
 * http://www.opensource.org/licenses/MIT
 */

package obsync

import (
	"context"
	"fmt"
	"log"
	"time"
)

// BucketRunner provides runner struct
type BucketRunner struct {
	Type      string
	Client    Bucket
	Config    BucketConfig
	taskPool  chan BucketTask
	observing chan bool
	Debug     bool
}

// Info get bucket runnner info
func (b BucketRunner) Info() (interface{}, error) {
	return b.Client.Info()
}

// AddTasks process all tasks
func (b BucketRunner) AddTasks(tasks []BucketTask) {
	if _, err := b.Info(); err != nil {
		if b.Debug {
			log.Printf("check status with error: %s", err)
		}

		return
	}

	if len(tasks) <= 0 {
		err := fmt.Errorf("tasks are empty")
		if b.Debug {
			log.Println(err.Error())
		}

		return
	}

	if b.Debug {
		log.Printf("total tasks are %d", len(tasks))
	}

	// process tasks without any error
	for _, task := range tasks {
		b.taskPool <- task
	}
}

func (b BucketRunner) Observe(ctx context.Context) {
	defer close(b.taskPool)
	for {
		select {
		case task := <-b.taskPool:
			if err := b.Run(ctx, task); err != nil && b.Debug {
				log.Println(err)
			} else {
				log.Printf("%v %v is finished, without any error", task.Key, task.Local)
			}

		case observing := <-b.observing:
			if !observing {
				log.Printf("%s | %s | %s", b.Config.Name, b.Type, "stop sbserving")
				return
			}
		}
	}
}

// Run run single task
func (b BucketRunner) Run(ctx context.Context, task BucketTask) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(b.Config.Timeout)*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		if b.Config.Force || !b.Client.Exists(task.Key) {
			if err := b.Client.Put(task); err != nil {
				log.Println(err)
				done <- err
				return
			}
		} else if b.Debug {
			log.Printf("%s | %s | %s is exists, ignore", b.Config.Name, b.Type, task.Key)
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if b.Debug {
			log.Printf("%s | %s | %s was done, error %v", b.Config.Name, b.Type, task.Key, err)
		}
		return err

	case <-timeoutCtx.Done():
		err := fmt.Errorf("%s | %s | %s was timeout", b.Config.Name, b.Type, task.Key)
		if b.Debug {
			log.Println(err)
		}
		return err
	}
}

// Stop to stopping observe
func (b BucketRunner) Stop() {
	if b.Debug {
		log.Printf("%v stop observing", b.Config.Name)
	}

	b.observing <- false
}

// NewBucketTask get new task instance
func NewBucketTask(typeName string, client Bucket, config BucketConfig, debug bool) (BucketRunner, error) {
	runner := BucketRunner{
		taskPool:  make(chan BucketTask, config.Thread),
		observing: make(chan bool),
		Type:      typeName,
		Client:    client,
		Config:    config,
		Debug:     debug,
	}

	return runner, nil
}
