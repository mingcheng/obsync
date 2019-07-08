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
	"sync"
	"time"
)

// BucketRunner provides runner struct
type BucketRunner struct {
	Type     string
	Client   Bucket
	Config   BucketConfig
	Context  *context.Context
	wg       *sync.WaitGroup
	taskChan chan BucketTask
	Debug    bool
}

// Info get bucket runnner info
func (b BucketRunner) Info() (interface{}, error) {
	return b.Client.Info()
}

// RunAll process all tasks
func (b BucketRunner) RunAll(tasks []BucketTask) error {
	if _, err := b.Info(); err != nil {
		if b.Debug {
			log.Printf("check status with error: %s", err)
		}

		return err
	}

	if len(tasks) <= 0 {
		err := fmt.Errorf("tasks are empty")
		if b.Debug {
			log.Println(err.Error())
		}
		return err
	}

	// process tasks without any error
	for _, task := range tasks {
		go b.Run(task)
	}

	return nil
}

// Run run single task
func (b BucketRunner) Run(task BucketTask) {
	b.taskChan <- task
	b.wg.Add(1)

	timeoutCtx, cancel := context.WithTimeout(*b.Context, time.Duration(b.Config.Timeout)*time.Second)
	done := make(chan bool, 0)

	defer func() {
		b.wg.Done()
		cancel()
	}()

	go func(d chan bool) {
		if b.Config.Force || !b.Client.Exists(task.Key) {
			b.Client.Put(task)
		} else if b.Debug {
			log.Printf("%s | %s | %s is exists, ignore", b.Config.Name, b.Type, task.Key)
		}

		<-b.taskChan
		d <- true
	}(done)

	select {
	case <-(*b.Context).Done():
		if b.Debug {
			log.Printf("%s | %s | %s was canceled", b.Config.Name, b.Type, task.Key)
		}

	case <-done:
		if b.Debug {
			log.Printf("%s | %s | %s was done", b.Config.Name, b.Type, task.Key)
		}

	case <-timeoutCtx.Done():
		if b.Debug {
			log.Printf("%s | %s | %s was timeout", b.Config.Name, b.Type, task.Key)
		}
	}
}

// Wait block when process task
func (b BucketRunner) Wait() {
	b.wg.Wait()
}

// NewBucketTask get new task instance
func NewBucketTask(ctx context.Context, typeName string, client Bucket, config BucketConfig, debug bool) (BucketRunner, error) {
	runner := BucketRunner{
		taskChan: make(chan BucketTask, config.Thread),
		wg:       &sync.WaitGroup{},
		Type:     typeName,
		Client:   client,
		Config:   config,
		Debug:    debug,
		Context:  &ctx,
	}

	return runner, nil
}
