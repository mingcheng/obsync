/**
 * File: main.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Monday, June 17th 2019, 3:12:43 pm
 * Last Modified: Monday, June 8th 2020, 2:13:06 pm
 *
 * http://www.opensource.org/licenses/MIT
 */
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/mingcheng/obsync"
	_ "github.com/mingcheng/obsync/buckets"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
	"sync"
	//"github.com/judwhite/go-svc"
)

const logo = `
/~\|~)(~\/|\ ||~
\_/|_)_)/ | \||_
`

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	configFilePath = flag.String("f", "/etc/obsync.yaml",
		"specified configuration file path, in yaml format")
	printVersion = flag.Bool("v", false, "print version and exit")
)

// PrintVersion that prints version and build time
func PrintVersion() {
	_, _ = fmt.Fprintf(os.Stderr, "Obsync v%v(%v)\nBuilt on %v %v/%v \n",
		version, commit, date, runtime.GOARCH, runtime.GOOS)

	supportTypes := obsync.AllSupportedBucketTypes()
	_, _ = fmt.Fprintf(os.Stderr, "Support bucket types [ %s ]\n\n", strings.Join(supportTypes, ", "))
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	if _, found := os.LookupEnv("DEBUG"); found {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	log.SetOutput(os.Stdout)

	// show command line usage information
	flag.Usage = func() {
		fmt.Println(logo)
		PrintVersion()
		flag.PrintDefaults()
	}
}

func main() {
	// parse command line
	flag.Parse()

	// print version and exit
	if *printVersion {
		flag.Usage()
		return
	}

	config, err := NewConfig(*configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	if config.Log.Path != "" {
		if f, err := os.OpenFile(config.Log.Path, os.O_APPEND|os.O_CREATE, 0644); err != nil {
			log.Error(err)
		} else {
			log.Debugf("set log path: %s", config.Log.Path)
			log.SetOutput(f)
		}
	}

	if config.Log.Debug {
		log.SetLevel(log.TraceLevel)
	}

	log.Tracef("configure is %v", config)
	var wg sync.WaitGroup
	wg.Add(len(config.RunnerConfigs))

	for _, config := range config.RunnerConfigs {
		runner, err := obsync.NewRunner(config)
		if err != nil {
			log.Fatal(err) //
		}

		go func(f *obsync.RunnerConfig) {
			defer wg.Done()
			log.Debugf("start running %v", f.Description)
			if err := runner.Start(context.Background()); err != nil {
				log.Error(err)
				return
			}
		}(&config)
	}

	wg.Wait()
}
