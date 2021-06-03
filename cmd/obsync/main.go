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
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/judwhite/go-svc"
	_ "github.com/mingcheng/obsync/cmd/obsync/bucket"
	"github.com/mingcheng/obsync/runner"
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
	configFilePath = flag.String("f", "", "config file path")
	printVersion   = flag.Bool("v", false, "print version and exit")
	printInfo      = flag.Bool("i", false, "print bucket info and exit")
)

// PrintVersion that prints version and build time
func PrintVersion() {
	_, _ = fmt.Fprintf(os.Stderr, "Obsync v%v(%v), built at %v on %v/%v \n\n", version, commit, date, runtime.GOARCH, runtime.GOOS)
}

func Runner(config *Config) (runner.Runner, error) {
	runner, err := runner.Init(config.Buckets, os.Getenv("DEBUG") != "")
	if err != nil {
		return nil, err
	}

	return runner, nil
}

func init() {
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
		log.Panic(err)
	}

	// detect root directory
	info, err := os.Stat(config.Root)
	if os.IsNotExist(err) || !info.IsDir() {
		log.Panicf("config root %s, is not exits or not a directory\n", config.Root)
	}

	r, err := Runner(config)
	if err != nil {
		log.Panic(err)
	}

	if *printInfo {
		info, _ := r.AllStatus()
		for k, i := range info {
			log.Println(k, i)
		}

		return
	}

	if err := svc.Run(&program{
		Config: *config,
		Runner: r,
	}); err != nil {
		log.Fatal(err)
	}
}
