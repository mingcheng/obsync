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
	RunTasks(tasks []BucketTask)
	Info() (interface{}, error)
	Exists(path string) bool
	Wait()
}

var buckets []Bucket

func RegisterBucket(bucket Bucket) {
	buckets = append(buckets, bucket)
}

func GetBucketInfo() ([]interface{}, error) {
	var result []interface{}
	for _, bucket := range buckets {
		if data, err := bucket.Info(); err != nil {
			return nil, err
		} else {
			result = append(result, data)
		}
	}

	return result, nil
}

func RunTasks(t []BucketTask) {
	for _, bucket := range buckets {
		bucket.RunTasks(t)
	}
}

func Wait() {
	for _, bucket := range buckets {
		bucket.Wait()
	}
}

// BucketTasksByPath get tasks by directory, ignore "." prefix files
func BucketTasksByPath(root string) ([]BucketTask, error) {
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
