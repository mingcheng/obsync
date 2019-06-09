package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type tabby struct {
	Source string
	Size   string
	Remote string
	Result string
}

type Task struct {
	waitGroup *sync.WaitGroup
	execChan  chan bool
	ObsTasks  []*Obs
}

func NewTask(size uint, tasks []*Obs) *Task {
	if config.Debug {
		log.Printf("thread number is %v", size)
	}

	return &Task{
		waitGroup: &sync.WaitGroup{},
		execChan:  make(chan bool, size),
		ObsTasks:  tasks,
	}
}

func (t *Task) Run() {
	for _, j := range t.ObsTasks {
		if config.Debug {
			log.Printf("number of goroutine is %d", runtime.NumGoroutine())
		}
		t.execChan <- true
		go t.sync(j)
	}
}

func (t *Task) Done() {
	if config.Debug {
		log.Println("stop all running task")
	}
	close(t.execChan)
}

func (t *Task) sync(obs *Obs) {
	defer func() {
		<-t.execChan
	}()

	tab := tabby{}
	tab.Source = filepath.Base(obs.SourceFile)
	tab.Remote = obs.RemoteKey

	if fi, err := os.Stat(obs.SourceFile); os.IsNotExist(err) {
		tab.Result = "NOT EXISTS"
	} else {
		tab.Size = fmt.Sprintf("%.2d", fi.Size())

		if config.Force || !obs.Exists() {

			if config.Debug {
				time.Sleep(2 * time.Second)
				tab.Result = "OK"
			} else {
				if output, err := obs.Put(); err != nil {
					tab.Result = "ERROR"
				} else {
					if output.StatusCode == http.StatusOK {
						tab.Result = "OK"
					} else {
						tab.Result = string(output.StatusCode)
					}
				}
			}
		} else {
			tab.Result = "IGNORE"
		}
	}

	fmt.Println(tab.Source, tab.Size, tab.Remote, tab.Result)
}
