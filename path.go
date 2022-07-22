package obsync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func prefixPath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	if strings.HasPrefix(filepath.Base(absPath), ".") {
		return true
	}

	if strings.HasPrefix(filepath.Base(filepath.Dir(absPath)), ".") {
		return true
	}

	if absPath == "/" {
		return false
	}

	return prefixPath(filepath.Dir(absPath))
}

// TasksByPath get tasks by the specified directory, ignore "." prefix files
func (r *Runner) TasksByPath(rootDir string, client *BucketClient, bucketConfig *BucketConfig) (tasks []*Task, err error) {
	var (
		absPath string
	)

	absPath, err = filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		// @TODO: handle
		// skip directories and dot prefix files
		if prefixPath(path) {
			return nil
		}

		if !info.IsDir() {
			pathKey := strings.Replace(path, absPath, "", 1)

			key := pathKey[1:]
			if bucketConfig.SubDir != "" {
				key = fmt.Sprintf("%s/%s", bucketConfig.SubDir, key)
			}

			// append new task to the list within directly configuration
			tasks = append(tasks, &Task{
				FilePath:  path,
				Key:       key,
				Client:    client,
				Overrides: r.config.Overrides,
			})
		}

		return nil
	})

	return tasks, nil
}
