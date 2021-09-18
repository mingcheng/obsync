package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mingcheng/obsync"
)

// TasksByPath get tasks by the specified directory, ignore "." prefix files
func TasksByPath(root string) ([]obsync.Task, error) {
	var tasks []obsync.Task

	if e := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// skip directories and dot prefix files
		if !info.IsDir() && strings.HasPrefix(path, root) && !strings.HasPrefix(info.Name(), ".") {
			key := path[len(root)+1:]
			if !strings.HasPrefix(key, ".") {
				tmp := obsync.Task{
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
