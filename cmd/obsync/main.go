/**
 * File: main.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Monday, June 17th 2019, 3:12:43 pm
 * Last Modified: Monday, June 17th 2019, 3:48:51 pm
 *
 * http://www.opensource.org/licenses/MIT
 */
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
	"time"

	"github.com/mingcheng/obsync.go"
	_ "github.com/mingcheng/obsync.go/cmd/obsync/bucket"
	"github.com/mingcheng/obsync.go/util"
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

// PrintVersion that print version and build time
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
	// pid, err := pidfile.New(*pidFilePath)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	//
	// defer pid.Remove()
	// if config.Debug {
	// 	log.Println(pid)
	// }

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
		log.Fatal(err)
	}

	if config.Debug {
		log.Println(config)
	}

	// @TODO
	obsync.AddBucketRunners(config.Buckets, config.Debug)

	if *printInfo {
		info, _ := obsync.GetBucketInfo()
		for k, i := range info {
			log.Println(k, i)
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

	go func() {
		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

		for s := range sig {
			switch s {
			default:
				log.Println("caught signal, stopping all tasks")
				os.Exit(0)
			}
		}
	}()

	// root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start observe
	go obsync.Observe(ctx)
	defer obsync.StopObserve()

	// start ticker to running tasks
	standbyDuration := time.Duration(config.Interval) * time.Hour
	ticker := time.NewTicker(standbyDuration)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		// get all obs tasks and put
		if tasks, err := obsync.TasksByPath(config.Root); err != nil || len(tasks) <= 0 {
			log.Printf("director %v is empty, caught %v", config.Root, err)
			return
		} else {
			// if anything is fine, add tasks to runners
			obsync.AddTasks(tasks)
		}

		// detect whether is standalone
		if config.Standalone {
			log.Printf("standalone mode, duration %v", standbyDuration)
		} else {
			log.Println("obsync is not configured in standalone mode, quiting")
			return
		}
	}
}
