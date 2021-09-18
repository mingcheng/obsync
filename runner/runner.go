package runner

import (
	"context"

	"github.com/mingcheng/obsync"
	"github.com/mingcheng/obsync/bucket"
)

type Runner interface {
	AddBucket(config bucket.Config) error
	AddBuckets(configs []bucket.Config) error

	AddTask(task obsync.Task)
	AddTasks(tasks []obsync.Task)

	AllStatus() ([]interface{}, error)
	Status(name string) interface{}

	Observe(ctx context.Context)
	Stop()
}
