package obsync

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var i int

type sleepClient struct {
	SleepDuration time.Duration
}

func (s sleepClient) Info(ctx context.Context) (interface{}, error) {
	return nil, nil
}

func (s sleepClient) Exists(ctx context.Context, s2 string) bool {
	return false
}

func (s sleepClient) Put(ctx context.Context, filePath, key string) error {
	i = i + 1

	done := make(chan bool)
	go func() {
		time.Sleep(s.SleepDuration)
		done <- true
	}()

	select {
	case <-done:
		fmt.Printf("%v\n", filePath)
		return nil
	case <-ctx.Done():
		fmt.Printf("%v\n", ctx)
		return fmt.Errorf("down")
	}
}

func init() {
	_ = AddBucketSyncFunc("sleep", func(config BucketConfig) (BucketSync, error) {
		return &sleepClient{
			time.Millisecond,
		}, nil
	})
}

func TestRunner_Start(t *testing.T) {
	runner, err := NewRunner(RunnerConfig{
		LocalPath: ".",
		Threads:   10,
		Timeout:   time.Second * 2,
		BucketConfigs: []BucketConfig{
			{
				Type: "sleep",
				Name: "sleep0",
			},
			{
				Type: "sleep",
				Name: "sleep1",
			},
		},
	})

	assert.NotNil(t, runner)

	err = runner.SyncDir(context.Background(), ".")
	assert.NoError(t, err)
}

func TestRunner_Watch(t *testing.T) {
	runner, err := NewRunner(RunnerConfig{
		LocalPath: ".",
		Threads:   10,
		Timeout:   time.Second * 2,
		BucketConfigs: []BucketConfig{
			{
				Type: "sleep",
				Name: "sleep0",
			},
			{
				Type: "sleep",
				Name: "sleep1",
			},
		},
	})

	assert.NotNil(t, runner)
	assert.NoError(t, err)
}

func TestNewRunner(t *testing.T) {

	_, err := NewRunner(RunnerConfig{
		LocalPath: ".",
		Threads:   10,
		BucketConfigs: []BucketConfig{
			{Type: "sleep", Name: "sleep0"},
		},
	})

	assert.NoError(t, err)

	_, err = NewRunner(RunnerConfig{
		LocalPath: "/dev/null",
		Threads:   10,
		BucketConfigs: []BucketConfig{
			{Type: "sleep", Name: "sleep0"},
		},
	})
	assert.Error(t, err)

	_, err = NewRunner(RunnerConfig{
		LocalPath: "/etc",
		Threads:   0,
		BucketConfigs: []BucketConfig{
			{Type: "sleep", Name: "sleep0"},
		},
	})
	assert.Error(t, err)

	_, err = NewRunner(RunnerConfig{
		LocalPath: "/etc",
		Threads:   10,
		BucketConfigs: []BucketConfig{
			{Type: "sleepxxx", Name: "sleep0"},
		},
	})
	assert.Error(t, err)
}
