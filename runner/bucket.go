package runner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mingcheng/obsync"
	"github.com/mingcheng/obsync/bucket"
)

type Bucket struct {
	Debug     bool
	Timeout   time.Duration
	taskPool  []chan obsync.Task
	buckets   []bucket.Bucket
	configs   []bucket.Config
	observing chan bool
}

func (b *Bucket) AllStatus() ([]interface{}, error) {
	for _, buckets := range b.buckets {
		fmt.Println(buckets.Info())
	}

	return nil, nil
}

func (b *Bucket) Status(name string) interface{} {
	panic("implement me")
}

func (b *Bucket) AddBucket(config bucket.Config) error {
	if len(b.buckets) != len(b.configs) || len(b.buckets) != len(b.taskPool) {
		return fmt.Errorf("the buckets and taskPool's size is not the same")
	}

	callback, err := bucket.Func(config.Type)
	if err != nil {
		return err
	}

	bucketHandler, err := callback(config)
	if err != nil {
		return err
	}

	err = bucketHandler.OnStart(context.Background())
	if err != nil {
		return err
	}

	b.buckets = append(b.buckets, bucketHandler)
	b.taskPool = append(b.taskPool, make(chan obsync.Task, config.Thread))
	b.configs = append(b.configs, config)

	return nil
}

func (b *Bucket) AddBuckets(configs []bucket.Config) error {
	for _, config := range configs {
		if err := b.AddBucket(config); err != nil {
			return err
		}
	}

	return nil
}

func (b *Bucket) AddTask(task obsync.Task) {
	for index := range b.taskPool {
		go func(i int) {
			config := b.configs[i]
			b.taskPool[i] <- obsync.Task{
				Local:   task.Local,
				Key:     task.Key,
				Force:   config.Force,
				Timeout: time.Duration(config.Timeout) * time.Second,
			}
		}(index)
	}
}

func (b *Bucket) AddTasks(tasks []obsync.Task) {
	for _, task := range tasks {
		b.AddTask(task)
	}
}

func (b *Bucket) Observe(ctx context.Context) {
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

func (b *Bucket) Stop() {
	if b.Debug {
		log.Println("stop observing")
	}

	for _, v := range b.buckets {
		err := v.OnStop(context.Background())
		if err != nil {
			log.Println(err.Error())
		}
	}

	b.observing <- false
}

// Run single task with `Task`
func (b *Bucket) run(ctx context.Context, task obsync.Task, client bucket.Bucket, config bucket.Config) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	done := make(chan error)
	go func() {
		if task.Force || !client.Exists(task.Key) {
			if config.SubDir != "" {
				task.SubDir = config.SubDir
			}

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

func Init(configs []bucket.Config, debug bool) (Runner, error) {
	var runner = &Bucket{
		observing: make(chan bool),
		Debug:     debug,
	}

	if err := runner.AddBuckets(configs); err != nil {
		return nil, err
	}

	return runner, nil
}
