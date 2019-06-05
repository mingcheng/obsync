package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/InVisionApp/tabular"
	"github.com/mingcheng/obs-sync.go/util"
)

var (
	version        = ""
	buildTime      = ""
	config         = &util.Config{}
	configFilePath = flag.String("f", util.DefaultConfig(), "config file path")
	printVersion   = flag.Bool("v", false, "print version and exit")
	printInfo      = flag.Bool("i", false, "print bucket info and exit")
)

// print version and build time, then exit
func PrintVersion() {
	_, _ = fmt.Fprintf(os.Stderr, "version: %s, %s\n", version, buildTime)
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
		log.Printf("size of obs taks is %d\n", len(obs))
	}
	return obs, nil
}

func main() {
	flag.Usage = func() {
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
		panic("configure file not exists")
	}

	// read config and initial obs client
	if err := config.Read(configFilePath); err != nil {
		panic(err)
	} else {
		Client(config.Key, config.Secret, config.EndPoint)
	}

	if *printInfo {
		if info, err := BucketInfo(); err != nil {
			panic(err)
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
		panic("config root it not exits or not a directory")
	} else if config.Debug {
		log.Printf("root path is %s\n", config.Root)
	}

	// tasks for put files to bucket
	tasks := make(chan *Obs, config.MaxThread)
	defer close(tasks)

	// register system signal
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer close(sig)

	go func() {
		var tab tabular.Table
		type tabby struct {
			Source string
			Size   string
			Remote string
			Result string
		}

		init := func() {
			tab = tabular.New()
			tab.Col("source", "Source File", 80)
			tab.ColRJ("size", "Size", 10)
			tab.Col("remote", "Remote Key", 50)
			tab.ColRJ("result", "Result", 10)
		}

		init()

		format := tab.Print("*")

		for {
			select {
			case obs := <-tasks:
				tab := tabby{}
				tab.Source = obs.SourceFile
				tab.Remote = obs.RemoteKey

				if fi, err := os.Stat(obs.SourceFile); os.IsNotExist(err) {
					tab.Result = "NOT EXISTS"
					fmt.Printf(format, tab.Source, tab.Size, tab.Remote, tab.Result)
					return
				} else {
					tab.Size = fmt.Sprintf("%.2d", fi.Size())
				}

				if config.Force || !obs.Exists() {
					if output, err := obs.Put(); err != nil {
						tab.Result = "ERROR"
					} else {
						if output.StatusCode == http.StatusOK {
							tab.Result = "OK"
						} else {
							tab.Result = string(output.StatusCode)
						}
					}
				} else {
					tab.Result = "IGNORE"
				}

				fmt.Printf(format, tab.Source, tab.Size, tab.Remote, tab.Result)
			}
		}
	}()

	go func() {
		if obs, err := ObsTasks(config.Root); err != nil {
			panic(err)
		} else {
			for _, j := range obs {
				tasks <- j
			}

			// all is done
			sig <- syscall.SIGQUIT
		}
	}()

	done := false
	for !done {
		select {
		case <-sig:
			done = true
			if config.Debug {
				log.Println("All is Done")
			}
		}
	}
}
