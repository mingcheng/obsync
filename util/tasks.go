package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mingcheng/obsync"
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
func TasksByPath(root string) ([]obsync.Task, error) {
	absPath, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var tasks []obsync.Task

	e := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		// skip directories and dot prefix files
		if prefixPath(path) {
			return nil
		}

		if !info.IsDir() {
			pathKey := strings.Replace(path, absPath, "", 1)
			tmp := obsync.Task{
				Local: path,
				Key:   pathKey[1:],
			}

			tasks = append(tasks, tmp)
		}

		return nil
	})

	if e != nil {
		return nil, e
	}

	return tasks, nil
}
