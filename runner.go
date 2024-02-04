package obsync

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	config *RunnerConfig
	pool   *ants.Pool
}

func (r *Runner) SyncDir(ctx context.Context, dir string) (err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Error(err)
		return
	}

	err = filepath.Walk(dir, func(localPath string, info os.FileInfo, err error) error {
		// @TODO: handle
		// skip directories and dot prefix files
		if prefixPath(localPath) {
			return nil
		}

		if !info.IsDir() {
			// exclude files by specified configuration
			for _, exclude := range r.config.Exclude {
				if found, _ := path.Match(exclude, filepath.Base(localPath)); found {
					log.Warnf("found exclude %s in %s", exclude, localPath)
					return nil
				}
			}

			pathKey := strings.Replace(localPath, dir, "", 1)
			key := pathKey[1:]

			log.Debugf("local path is %s, and key is %s", localPath, key)

			err = r.InvokeAll(ctx, key, localPath)
			if err != nil {
				log.Error(err)
				return err
			}
		}

		return err
	})

	return err
}

func (r *Runner) InvokeAll(ctx context.Context, key, local string) (err error) {
	var wg sync.WaitGroup
	for _, c := range r.config.BucketConfigs {
		wg.Add(1)
		err = r.pool.Submit(func() {
			err = r.Invoke(ctx, key, local, c)
			if err != nil {
				log.Error(err)
			}
			defer wg.Done()
		})
		if err != nil {
			log.Error(err)
		}
	}

	wg.Wait()
	return err
}

func (r *Runner) Invoke(ctx context.Context, key, local string, c BucketConfig) (err error) {
	if c.SubDir != "" {
		key = fmt.Sprintf("%s%c%s", c.SubDir, os.PathSeparator, key)
	}

	task, err := NewTask(key, local, r.config.Overrides, c)
	if err != nil {
		log.Error(err)
		return err
	}

	return task.Put(ctx)
}

func (r *Runner) Stop() error {
	if !r.pool.IsClosed() {
		r.pool.Release()
	}
	return nil
}

// NewRunner to instance a new runner with specified configuration
// notice: 	1. the bucket type must be registered
//  2. the local directory must readable
//  3. the threads must be greater than zero
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
		_, err := GetBucketSyncFunc(bucketConfig.Type)
		if err != nil {
			return nil, err
		}
	}

	pool, err := ants.NewPool(int(config.Threads))
	if err != nil {
		return nil, err
	}

	return &Runner{
		pool:   pool,
		config: &config,
	}, nil
}
