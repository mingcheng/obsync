package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mingcheng/obsync.go/util"
	"github.com/mingcheng/pidfile"
)

const logo = `
/~\|~)(~\/|\ ||~
\_/|_)_)/ | \||_
`

var (
	version        = "dev"
	commit         = "none"
	date           = "unknown"
	config         = &util.Config{}
	configFilePath = flag.String("f", util.DefaultConfig(), "config file path")
	pidFilePath    = flag.String("pid", "/var/run/obsync.pid", "pid file path")
	printVersion   = flag.Bool("v", false, "print version and exit")
	printInfo      = flag.Bool("i", false, "print bucket info and exit")
)

// print version and build time, then exit
func PrintVersion() {
	_, _ = fmt.Fprintf(os.Stderr, "Obsync v%v, built at %v\n%v\n\n", version, date, commit)
}

func main() {
	// show command line usage information
	flag.Usage = func() {
		fmt.Println(logo)
		PrintVersion()
		flag.PrintDefaults()
	}

	// parse command line
	flag.Parse()

	// detect pid file exists, and generate pid file
	if pid, err := pidfile.New(*pidFilePath); err != nil {
		log.Println(err)
		return
	} else {
		defer pid.Remove()
		if config.Debug {
			log.Println(pid)
		}
	}

	// print version and exit
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
		log.Println(err)
		return
	} else {
		if len(config.Key) <= 0 {
			config.Key = os.Getenv("OBS_KEY")
		}

		if len(config.Secret) <= 0 {
			config.Secret = os.Getenv("OBS_SECRET")
		}

		if config.Debug {
			log.Println(config)
		}

		NewClient(config.Key, config.Secret, config.EndPoint, int(config.Timeout))
	}

	if *printInfo {
		if info, err := BucketInfo(); err != nil {
			log.Println(err)
		} else {
			if config.Debug {
				_, _ = fmt.Fprintln(os.Stderr, config)
			}
			_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", config.Bucket, info)
		}

		return
	}

	// detect root directory
	config.Root, _ = filepath.Abs(config.Root)
	if info, err := os.Stat(config.Root); os.IsNotExist(err) || !info.IsDir() {
		log.Printf("config root %s, is not exits or not a directory\n", config.Root)
		return
	} else if config.Debug {
		log.Printf("root path is %s\n", config.Root)
	}

	// get all obs tasks and put
	if obs, err := ObsTasks(config.Root); err != nil {
		log.Println(err)
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
						log.Println("caught signal, stopping all tasks")
						cancel()
					}
				}
			}()

			// running sync task
			syncTask.Run()
		} else {
			log.Println("obs list is empty")
		}
	}
}
