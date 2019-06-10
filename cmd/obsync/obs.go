package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// get bucket info, usage and number of files
func BucketInfo() (info string, err error) {
	obs := &Obs{
		BucketName: config.Bucket,
	}

	if info, err := obs.Info(); err != nil {
		return "", err
	} else {
		return fmt.Sprintf("size %d Kb, %d files", info.Size/1024.0, info.ObjectNumber), nil
	}
}

// get obs tasks by directory, ignore "." prefix files
func ObsTasks(root string) (tasks []*Obs, err error) {
	var obs []*Obs

	if e := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// skip directories and dot prefix files
		if !info.IsDir() && strings.HasPrefix(path, root) && !strings.HasPrefix(info.Name(), ".") {
			key := path[len(root)+1:]
			if !strings.HasPrefix(key, ".") {
				tmp := &Obs{
					SourceFile: path,
					RemoteKey:  key,
					BucketName: config.Bucket,
				}

				obs = append(obs, tmp)
			}
		}

		return nil
	}); e != nil {
		return obs, e
	}

	if config.Debug {
		log.Printf("size of obs tasks is %d\n", len(obs))
	}
	return obs, nil
}
