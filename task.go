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
func (b BucketRunner) RunAll(ctx context.Context, tasks []BucketTask) {
	if _, err := b.Info(); err != nil {
		if b.Debug {
			log.Printf("check status with error: %s", err)
		}
	}

	if len(tasks) <= 0 {
		err := fmt.Errorf("tasks are empty")
		if b.Debug {
			log.Println(err.Error())
		}

		return
	} else if b.Debug {
		log.Printf("total tasks are %d", len(tasks))
	}

	go func() {
		var i = 0
		for i < len(tasks) {
			task := <-b.taskChan
			if err := b.Run(ctx, task); err != nil && b.Debug {
				log.Println(err)
			}
			b.wg.Done()
			i++
		}
	}()

	// process tasks without any error
	b.wg.Add(len(tasks))
	for _, task := range tasks {
		b.taskChan <- task
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

// Wait block when process task
func (b BucketRunner) Wait() {
	b.wg.Wait()
}

// NewBucketTask get new task instance
func NewBucketTask(typeName string, client Bucket, config BucketConfig, debug bool) (BucketRunner, error) {
	runner := BucketRunner{
		taskChan: make(chan BucketTask, config.Thread),
		wg:       &sync.WaitGroup{},
		Type:     typeName,
		Client:   client,
		Config:   config,
		Debug:    debug,
	}

	return runner, nil
}
