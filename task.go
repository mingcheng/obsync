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
	taskChan  chan BucketTask
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
		b.taskChan <- task
	}
}

func (b BucketRunner) Observe(ctx context.Context) {
	defer close(b.taskChan)
	for {
		select {
		case task := <-b.taskChan:
			if err := b.Run(ctx, task); err != nil && b.Debug {
				log.Println(err)
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

	done := make(chan bool)
	go func(d chan bool) {
		if b.Config.Force || !b.Client.Exists(task.Key) {
			b.Client.Put(task)
		} else if b.Debug {
			log.Printf("%s | %s | %s is exists, ignore", b.Config.Name, b.Type, task.Key)
		}
		d <- true
	}(done)

	select {
	case <-done:
		if b.Debug {
			log.Printf("%s | %s | %s was done", b.Config.Name, b.Type, task.Key)
		}
		return nil

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
		taskChan:  make(chan BucketTask, config.Thread),
		observing: make(chan bool),
		Type:      typeName,
		Client:    client,
		Config:    config,
		Debug:     debug,
	}

	return runner, nil
}
