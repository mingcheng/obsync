package obsync

import (
	"context"
	"log"
	"sync"
	"time"
)

type BucketRunner struct {
	Type     string
	Client   Bucket
	Config   BucketConfig
	Context  context.Context
	wg       sync.WaitGroup
	taskChan chan BucketTask
	Debug    bool
}

func (b *BucketRunner) Info() (interface{}, error) {
	return b.Client.Info()
}

func (b *BucketRunner) RunAll(tasks []BucketTask) {
	if _, err := b.Info(); err != nil {
		if b.Debug {
			log.Printf("check status with error: %v", err)
		}
		return
	}

	for _, task := range tasks {
		go b.Run(task)
	}
}

func (b *BucketRunner) Run(task BucketTask) {
	b.taskChan <- task
	b.wg.Add(1)
	defer b.wg.Done()

	timeoutCtx, _ := context.WithTimeout(b.Context, time.Duration(b.Config.Timeout)*time.Second)
	done := make(chan bool, 0)

	go func() {
		b.Client.Put(task)
		<-b.taskChan
		done <- true
	}()

	select {
	case <-b.Context.Done():
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

func (b *BucketRunner) Wait() {
	b.wg.Wait()
}

func NewBucketTask(ctx context.Context, typeName string, client Bucket, config BucketConfig, debug bool) (BucketRunner, error) {
	runner := BucketRunner{
		taskChan: make(chan BucketTask, config.Thread),
		Type:     typeName,
		Client:   client,
		Config:   config,
		Debug:    debug,
		Context:  ctx,
	}

	return runner, nil
}
