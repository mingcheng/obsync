package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mingcheng/obsync.go/util"
)

const logo = `
/~\|~)(~\/|\ ||~
\_/|_)_)/ | \||_
`

var (
	version        = "dev"
	commit         = "none"
	date           = "unkown"
	config         = &util.Config{}
	configFilePath = flag.String("f", util.DefaultConfig(), "config file path")
	printVersion   = flag.Bool("v", false, "print version and exit")
	printInfo      = flag.Bool("i", false, "print bucket info and exit")
)

// print version and build time, then exit
func PrintVersion() {
	_, _ = fmt.Fprintf(os.Stderr, "Obsync v%v, built at %v\n%v\n\n", version, date, commit)
}

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

func main() {
	flag.Usage = func() {
		fmt.Println(logo)
		PrintVersion()
		flag.PrintDefaults()
	}

	// parse command line
	flag.Parse()

	if *printVersion {
		flag.Usage()
		return
	}

	// detect config file path
	configFilePath, _ := filepath.Abs(*configFilePath)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.Fatalf("configure file %s is not exists\n", configFilePath)
	}

	// read config and initial obs client
	if err := config.Read(configFilePath); err != nil {
		log.Fatalln(err)
	} else {
		NewClient(config.Key, config.Secret, config.EndPoint, int(config.Timeout))
	}

	if *printInfo {
		if info, err := BucketInfo(); err != nil {
			log.Fatalln(err)
		} else {
			if config.Debug {
				_, _ = fmt.Fprintln(os.Stderr, config)
			}
			_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", config.Bucket, info)
			os.Exit(0)
		}
	}

	// detect root directory
	config.Root, _ = filepath.Abs(config.Root)
	if info, err := os.Stat(config.Root); os.IsNotExist(err) || !info.IsDir() {
		log.Fatalf("config root %s is not exits or not a directory\n", config.Root)
	} else if config.Debug {
		log.Printf("root path is %s\n", config.Root)
	}

	// get all obs tasks and put
	if obs, err := ObsTasks(config.Root); err != nil {
		log.Fatalln(err)
	} else {
		if len(obs) > 0 {
			ctx, cancel := context.WithCancel(context.TODO())
			syncTask := NewTask(ctx, config.MaxThread, obs)

			// waiting for system s or user interrupt
			go func() {
				// register system signal
				sig := make(chan os.Signal)
				signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

				for s := range sig {
					switch s {
					default:
						log.Println("caught signal stopping all tasks")
						cancel()
					}
				}
			}()

			// running sync task
			syncTask.Run()
		} else {
			log.Fatalln("obs list is empty")
		}
	}
}
