package obsync

import (
	"context"
	"fmt"
	"github.com/Jeffail/tunny"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

// RunnerConfig represents a configuration for running
type RunnerConfig struct {
	LocalPath     string         `yaml:"path" json:"path"`
	Description   string         `yaml:"description" json:"description"`
	Overrides     bool           `yaml:"overrides" json:"overrides"`
	Exclude       []string       `yaml:"exclude" json:"exclude"`
	Timeout       time.Duration  `yaml:"timeout" json:"timeout"`
	Threads       uint           `yaml:"threads" json:"threads"`
	BucketConfigs []BucketConfig `yaml:"buckets" json:"buckets"`
}

// Runner supported synchronous local files to specified bucket
type Runner struct {
	taskPool map[int]*tunny.Pool
	config   *RunnerConfig
}

// Start the runner with specified configurationF
func (r *Runner) Start(ctx context.Context) (err error) {
	threads := int(r.config.Threads)
	if threads <= 0 {
		return fmt.Errorf("threads must be greater than 0")
	}

	var wg sync.WaitGroup

	for index, config := range r.config.BucketConfigs {
		clientFunc, err := NewBucketClientFunc(config.Type)
		if err != nil {
			log.Errorf("bucket which name is not supported: %v", err)
			continue
		}

		client, err := clientFunc(config)
		if err != nil {
			log.Errorf("new bucket client failed: %v", err)
			continue
		}

		tasks, err := r.TasksByPath(r.config.LocalPath, &client)
		if err != nil {
			log.Errorf("get task failed: %v", err)
			continue
		}
		wg.Add(len(tasks))

		log.Tracef("instance threads pool size: %d", threads)
		pool := tunny.NewCallback(threads)

		for _, t := range tasks {
			go func(task *Task) {
				defer wg.Done()

				log.Tracef("bucket [%s] local path: [%s], remote key: [%s]",
					config.Name, task.FilePath, task.Key)

				if _, err := pool.ProcessCtx(ctx, func() {
					// fork the new timeout context for running tasks
					taskCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
					defer cancel()

					// start running tasks within the specified context
					if err := task.Run(taskCtx); err != nil {
						log.Error(err)
					}
				}); err != nil {
					log.Error(err)
				}
			}(t)
		}

		r.taskPool[index] = pool
	}

	wg.Wait()
	return err
}

// Stop the running and closing threads pool
func (r *Runner) Stop() (err error) {
	for name, pool := range r.taskPool {
		if pool == nil {
			log.Debugf("closing task pool %v", name)
			pool.Close()
		}
	}
	return nil
}

// NewRunner to instance a new runner with specified configuration
// notice: 	1. the bucket type must be registered
//					2. the local directory must readable
// 					3. the threads must be greater than zero
func NewRunner(config RunnerConfig) (*Runner, error) {
	stat, err := os.Stat(config.LocalPath)
	if err != nil || !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", config.LocalPath)
	}

	if config.Threads <= 0 {
		return nil, fmt.Errorf("the number of threads must be greater than zero")
	}

	// check if the bucket type is supported
	for _, bucketConfig := range config.BucketConfigs {
		_, err = NewBucketClientFunc(bucketConfig.Type)
		if err != nil {
			return nil, err
		}
	}

	return &Runner{
		taskPool: map[int]*tunny.Pool{},
		config:   &config,
	}, nil
}
