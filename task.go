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
)

type Task struct {
	FilePath  string
	Key       string
	Overrides bool
	Client    *BucketSync
}

func (t *Task) Put(ctx context.Context) (err error) {
	if t.Client == nil {
		return fmt.Errorf("the client is nil")
	}

	if (*t.Client).Exists(ctx, t.Key) && !t.Overrides {
		return fmt.Errorf("%s is already exists", t.Key)
	}

	return (*t.Client).Put(ctx, t.FilePath, t.Key)
}

func NewTask(key, path string, overrides bool, config BucketConfig) (task *Task, err error) {
	fn, err := GetBucketSyncFunc(config.Type)
	if err != nil {
		return nil, err
	}

	client, err := fn(config)
	if err != nil {
		return nil, err
	}

	task = &Task{
		Key:       key,
		FilePath:  path,
		Overrides: overrides,
		Client:    &client,
	}

	return task, nil
}
