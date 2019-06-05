package util

import (
	"os"
	"os/user"
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

func DefaultConfig() string {
	return HomeDir() + "/.obsync.json"
}

func DebugConfig() string {
	dir, _ := os.Getwd()
	return dir + "/config.json"
}
