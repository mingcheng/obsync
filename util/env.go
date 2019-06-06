package util

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

func HomeDir() string {
	if user, err := user.Current(); err != nil {
		if runtime.GOOS == "windows" {
			home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
			if home == "" {
				home = os.Getenv("USERPROFILE")
			}
			return home
		}
		return os.Getenv("HOME")
	} else {
		return user.HomeDir
	}
}

const configFileName = "obsync.json"

func DefaultConfig() string {
	return filepath.Join("/etc", configFileName)
}

func DebugConfig() string {
	if pwd, err := os.Getwd(); err != nil {
		return DefaultConfig()
	} else {
		return filepath.Join(pwd, configFileName)
	}
}
