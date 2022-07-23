package obsync

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
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
	var absPath string

	absPath, err = filepath.Abs(rootDir)
	if err != nil {
		log.Error(err)
		return
	}

	err = filepath.Walk(absPath, func(localPath string, info os.FileInfo, err error) error {
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

			pathKey := strings.Replace(localPath, absPath, "", 1)
			key := pathKey[1:]

			if bucketConfig.SubDir != "" {
				key = fmt.Sprintf("%s%c%s", bucketConfig.SubDir, os.PathSeparator, key)
			}

			// append new task to the list within directly configuration
			tasks = append(tasks, &Task{
				FilePath:  localPath,
				Key:       key,
				Client:    client,
				Overrides: r.config.Overrides,
			})
		}

		return nil
	})

	return tasks, nil
}
