package obsync

import (
	"os"
	"path/filepath"
	"strings"
)

// TasksByPath get tasks by the specified directory, ignore "." prefix files
func (r *Runner) TasksByPath(rootDir string, client *BucketClient) (tasks []*Task, err error) {
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

			tasks = append(tasks, &Task{
				FilePath:  path,
				Key:       pathKey[1:],
				Client:    client,
				Overrides: r.config.Overrides,
			})
		}

		return nil
	})

	return tasks, nil
}
