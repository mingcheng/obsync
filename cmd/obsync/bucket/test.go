package bucket

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/mingcheng/obsync.go"
)

type TestBucket struct {
	ctx      context.Context
	wg       sync.WaitGroup
	taskChan chan obsync.BucketTask
	Debug    bool
	Timeout  uint64
}

func (b *TestBucket) RunTasks(tasks []obsync.BucketTask) {
	for _, task := range tasks {
		go b.Put(task)
	}
}

func (b *TestBucket) Put(task obsync.BucketTask) {
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
		log.Println(task.Key)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		<-b.taskChan
	}
}

func (b *TestBucket) Wait() {
	b.wg.Wait()
}

func (*TestBucket) Info() (interface{}, error) {
	return "This is a TEST bucket", nil
}

func (*TestBucket) Exists(path string) bool {
	return false
}

func NewTestBucket(cxt context.Context, config obsync.BucketConfig, debug bool) (*TestBucket, error) {
	return &TestBucket{
		taskChan: make(chan obsync.BucketTask, config.Thread),
		ctx:      cxt,
		Timeout:  config.Timeout,
		Debug:    debug,
	}, nil
}
