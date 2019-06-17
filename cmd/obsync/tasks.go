/**
 * File: tasks.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Thursday, June 13th 2019, 4:08:07 pm
 * Last Modified: Monday, June 17th 2019, 3:53:00 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package main

import (
	"context"
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

// Task struct contrains task information
type Task struct {
	wg       sync.WaitGroup
	ctx      context.Context
	execChan chan bool
	ObsTasks []*Obs
}

// NewTask provides new task
func NewTask(ctx context.Context, size uint, tasks []*Obs) *Task {
	if config.Debug {
		log.Printf("thread number is %v", size)
	}

	return &Task{
		execChan: make(chan bool, size),
		ObsTasks: tasks,
		ctx:      ctx,
	}
}

// Run a task
func (t *Task) Run() {
	t.wg.Add(len(t.ObsTasks))

	for _, j := range t.ObsTasks {
		if config.Debug {
			log.Printf("number of goroutine is %d", runtime.NumGoroutine())
		}

		select {
		case <-t.ctx.Done():
			close(t.execChan)
			return

		case t.execChan <- true:
			go t.sync(j)
		}
	}

	t.wg.Wait()
}

func (t *Task) sync(obs *Obs) {
	defer func() {
		t.wg.Done()
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
				// NOTICE: only for test, do not actually upload
				time.Sleep(2 * time.Second)
				tab.Result = "OK"
			} else {
				if output, err := obs.Put(); err != nil {
					log.Println(err)
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
