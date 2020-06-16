package obsync

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Runner interface {
	AddBucket(config BucketConfig) error
	AddBuckets(configs []BucketConfig) error

	AddTask(task BucketTask)
	AddTasks(tasks []BucketTask)

	AllStatus() ([]interface{}, error)
	Status(name string) interface{}

	Observe(ctx context.Context)
	Stop()
}

type BucketRunner struct {
	Debug     bool
	Timeout   time.Duration
	taskPool  []chan BucketTask
	buckets   []Bucket
	configs   []BucketConfig
	observing chan bool
}

func (b *BucketRunner) AllStatus() ([]interface{}, error) {
	for _, buckets := range b.buckets {
		fmt.Println(buckets.Info())
	}

	return nil, nil
}

func (b *BucketRunner) Status(name string) interface{} {
	panic("implement me")
}

func (b *BucketRunner) AddBucket(config BucketConfig) error {
	if len(b.buckets) != len(b.configs) || len(b.buckets) != len(b.taskPool) {
		return fmt.Errorf("the buckets and taskPool's size is not the same")
	}

	callback, err := BucketCallback(config.Type)
	if err != nil {
		return err
	}

	bucket, err := callback(config)
	if err != nil {
		return err
	}

	b.buckets = append(b.buckets, bucket)
	b.taskPool = append(b.taskPool, make(chan BucketTask, config.Thread))
	b.configs = append(b.configs, config)
	return nil
}

func (b *BucketRunner) AddBuckets(configs []BucketConfig) error {
	for _, config := range configs {
		if err := b.AddBucket(config); err != nil {
			return err
		}
	}

	return nil
}

func (b *BucketRunner) AddTask(task BucketTask) {
	for index := range b.taskPool {
		go func(i int) {
			config := b.configs[i]
			b.taskPool[i] <- BucketTask{
				Local:   task.Local,
				Key:     task.Key,
				Force:   config.Force,
				Timeout: time.Duration(config.Timeout) * time.Second,
			}
		}(index)
	}
}

func (b *BucketRunner) AddTasks(tasks []BucketTask) {
	for _, task := range tasks {
		b.AddTask(task)
	}
}

func (b *BucketRunner) Observe(ctx context.Context) {
	for index := range b.taskPool {
		go func(i int) {
			for {
				select {
				case task := <-b.taskPool[i]:
					if err := b.run(ctx, task, b.buckets[i], b.configs[i]); err != nil && b.Debug {
						log.Println(err)
					} else {
						log.Printf("%v %v is finished, without any error", task.Key, task.Local)
					}
				case observing := <-b.observing:
					if !observing {
						return
					}
				}
			}
		}(index)
	}
}

func (b *BucketRunner) Stop() {
	if b.Debug {
		log.Println("stop observing")
	}

	b.observing <- false
}

// Run single task with `BucketTask`
func (b *BucketRunner) run(ctx context.Context, task BucketTask, client Bucket, config BucketConfig) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	done := make(chan error)
	go func() {
		if task.Force || !client.Exists(task.Key) {
			if err := client.Put(task); err != nil {
				log.Println(err)
				done <- err
				return
			}
		} else if b.Debug {
			log.Printf(" %s is exists, ignore", task.Key)
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if b.Debug {
			log.Printf("%s was done, error %v", task.Key, err)
		}
		return err

	case <-timeoutCtx.Done():
		err := fmt.Errorf("%s was timeout", task.Key)
		if b.Debug {
			log.Println(err)
		}
		return err
	}
}

func InitRunnerWithBuckets(configs []BucketConfig, debug bool) (Runner, error) {
	var runner = &BucketRunner{
		observing: make(chan bool),
		Debug:     debug,
	}

	if err := runner.AddBuckets(configs); err != nil {
		return nil, err
	}

	return runner, nil
}
