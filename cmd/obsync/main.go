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
	_, _ = fmt.Fprintf(os.Stderr, "Obsync v%v(%v)\nbuilt on %v %v/%v \n",
		version, commit, date, runtime.GOARCH, runtime.GOOS)

	supportTypes := obsync.AllSupportedBucketTypes()
	_, _ = fmt.Fprintf(os.Stderr, "supports bucket types [ %s ]\n\n", strings.Join(supportTypes, ", "))
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
		log.Println(logo)
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

	for _, c := range config.RunnerConfigs {
		runner, err := obsync.NewRunner(c)
		if err != nil {
			log.Fatal(err)
		}

		log.Debugf("start running %v", c.Description)
		if err := runner.SyncDir(context.Background(), c.LocalPath); err != nil {
			log.Error(err)
			return
		}
	}
}
