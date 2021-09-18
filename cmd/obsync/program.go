package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/judwhite/go-svc"
	"github.com/mingcheng/obsync/runner"
	"github.com/mingcheng/obsync/util"
)

type program struct {
	Config
	runner.Runner
	Ticker *time.Ticker
}

func (p *program) Init(env svc.Environment) error {
	interval, _ := strconv.ParseUint(os.Getenv("INTERVAL"), 10, 64)
	if interval == 0 {
		interval = 1
	}

	dur := time.Duration(interval) * time.Hour
	log.Print(dur)

	p.Ticker = time.NewTicker(dur)
	return nil
}

func (p *program) Start() error {
	// start observe
	go p.Runner.Observe(context.Background())

	go func() {
		for ; true; <-p.Ticker.C {
			// get all obs tasks and send to server
			tasks, err := util.TasksByPath(p.Config.Root)
			if err != nil || len(tasks) <= 0 {
				log.Printf("director %v is empty, caught %v", p.Config.Root, err)
				return
			}

			// if anything is fine, add tasks to runners
			p.Runner.AddTasks(tasks)
		}
	}()

	return nil
}

func (p *program) Stop() error {
	p.Ticker.Stop()
	p.Runner.Stop()
	return nil
}
