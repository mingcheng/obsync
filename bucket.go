/**
 * File: bucket.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Tuesday, June 18th 2019, 6:27:36 pm
 * Last Modified: Tuesday, June 18th 2019, 7:04:14 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package obsync

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// BucketConfig bucket config
type BucketConfig struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Key      string `json:"key"`
	Secret   string `json:"secret"`
	Force    bool   `json:"force"`
	EndPoint string `json:"endpoint"`
	Timeout  uint64 `json:"timeout"`
	Thread   uint64 `json:"thread"`
}

type BucketTask struct {
	Local string
	Key   string
}

type Bucket interface {
	Info() (interface{}, error)
	Exists(path string) bool
	Put(task BucketTask)
}

var (
	buckets = make(map[string]func(c BucketConfig) (Bucket, error))
	runners []BucketRunner
)

func RegisterBucket(typeName string, f func(c BucketConfig) (Bucket, error)) {
	buckets[typeName] = f
}

func AddBucketRunners(ctx context.Context, debug bool, configs []BucketConfig) {
	for _, config := range configs {
		if err := addSingleRunner(ctx, config.Type, debug, config); err != nil {
			log.Println(err.Error())
		}
	}
}

func addSingleRunner(ctx context.Context, typeName string, debug bool, config BucketConfig) error {
	if callback, ok := buckets[typeName]; !ok {
		return fmt.Errorf("err: bucket callback which name %s does not exists", typeName)
	} else {
		client, err := callback(config)
		if err != nil {
			return err
		}

		if runner, err := NewBucketTask(&ctx, typeName, client, config, debug); err != nil {
			return err
		} else {
			runners = append(runners, runner)
		}
		return nil
	}
}

func GetBucketInfo() ([]interface{}, error) {
	var result []interface{}
	for _, bucket := range runners {
		data, _ := bucket.Info()
		result = append(result, data)
	}

	return result, nil
}

func RunTasks(t []BucketTask) {
	for _, runner := range runners {
		runner.RunAll(t)
	}
}

func Wait() {
	for _, runner := range runners {
		runner.Wait()
	}
}

// TasksByPath get tasks by directory, ignore "." prefix files
func TasksByPath(root string) ([]BucketTask, error) {
	var tasks []BucketTask

	if e := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// skip directories and dot prefix files
		if !info.IsDir() && strings.HasPrefix(path, root) && !strings.HasPrefix(info.Name(), ".") {
			key := path[len(root)+1:]
			if !strings.HasPrefix(key, ".") {
				tmp := BucketTask{
					Local: path,
					Key:   key,
				}

				tasks = append(tasks, tmp)
			}
		}

		return nil
	}); e != nil {
		return nil, e
	}

	return tasks, nil
}
