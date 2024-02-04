package obsync

import (
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
