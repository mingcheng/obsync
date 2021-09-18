package runner

import (
	"context"

	"github.com/mingcheng/obsync/bucket"
	"github.com/mingcheng/obsync/internal"
)

type Runner interface {
	AddBucket(config bucket.Config) error
	AddBuckets(configs []bucket.Config) error

	AddTask(task internal.Task)
	AddTasks(tasks []internal.Task)

	AllStatus() ([]interface{}, error)
	Status(name string) interface{}

	Observe(ctx context.Context)
	Stop()
}
