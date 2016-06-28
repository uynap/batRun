// Copyright 2016 Pan Yu. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

// Package batRun is a Multitask framework using Producer Consumer pattern to run multiple works in different goroutines.
//
// A quick example is:
//
//  package main
//  import (
//      "github.com/uynap/batRun"
//      "time"
//  ）
//
//  func main() {
//      bat := batRun.NewBat()
//      bat.AddProducers(producer)
//      bat.AddStep(worker, 2) // 2 concurrent workers for step1
//      bat.AddStep(worker2, 4) // 4 concurrent workers for step2
//      report := batRun.Run()
//  }
//
//  func produce(task *batRun.Task) {
//      for i := 1; i <= 10; i++ {
//  	    ctx := task.NewContext(map[string]int{"id": i})
//  		task.Submit(ctx, 4*time.Second)  // Timeout in 4 seconds
//  	}
//  }
//
//  func worker(ctx *batRun.Context) error {
//      data := ctx.GetContext().(map[string]int)
//  	data["age"] = 34
//      ctx.SetContext(data)
//  	return nil
//  }
//
//  func worker2(ctx *batRun.Context) error {
//      data := ctx.GetContext().(map[string]int)
//  	println(data["age"])
//  	return nil
//  }
// A more complex example is presented in the test case.
// Basically, you need to provide at least 2 functions.
// One is the "producer" which is used for creating tasks.
// Another one is the "worker" which will run for a work.

// Multiple producers and workers are supported. The diagram below shows the working process.
// +--------------+    +---------------------+    +---------------------+
// |  Producer A  |    |        | Worker 1-1 |    |        | Worker 2-1 |
// +--------------+    |        | Worker 1-2 |    |        | Worker 2-2 |                 +----------+
//    ...           -> | Work 1 | Worker 1-3 | -> | Work 2 | Worker 2-3 | -> ...Work N -> | Reporter |
// +--------------+    |        | ...        |    |        | ...        |                 +----------+
// |  Producer Z  |    |        | Worker 1-n |    |        | Worker 2-n |
// +--------------+    +---------------------+    +---------------------+

// batRun supports multiple producers generating tasks.
// A task may contains multiple works. For each of the work, you can specify a number of workers, which are goroutines.
// A common use case is you got a group of tasks. Each of the tasks contains a number of blocking assignments. Then you can put these blocking assignments in multiple works.

package batRun2

import (
	"golang.org/x/net/context"
	"sync"
	"time"
)

const (
	contextData = iota
	contextTimeout
    contextError
)

// Task is for "producer" to create contexts
type Task struct {
	TaskQueue chan context.Context
	Total     int
}

// Context is a wraper of context.Context and is used for handling
// data through out the task.
type Context struct {
	ctx    context.Context
	Cancel func()
}

// Report is the report for the whole task life circle.
type Report struct {
	Start time.Time
	End   time.Time
	// The result for all task.
    Result map[int]int
}

// Bat is a single assignment, consisting producer(s) and work(s).
type Bat struct {
	Producers []Producer
	Works     []Workers
}

// Producer is a function that can be used for generating tasks for workers.
type Producer func(*Task)

// Worker is a function that can consume tasks from producer(s).
type Worker func(*Context) error

// Workers is a certain number of workers, consisting a worker and a number. 
type Workers struct {
	Worker Worker
	Num    int
}

// Error contains 2 parts information. One is the error message if there is an error during in all workers. Another one is indicating in which stage it happened.
type Error struct {
    Error error
    Count int
}

// NewBat returns an empty Bat struct.
func NewBat() *Bat {
	return &Bat{}
}

// AddProducers is a function for users to add producers. Multiple producers are supported.
func (bat *Bat) AddProducers(producers ...Producer) {
	bat.Producers = producers
}

// AddWork is a function for users to add a work by providing a "worker" function and the quantity of these workers.
func (bat *Bat) AddWork(worker Worker, num int) {
	workers := Workers{
		Worker: worker,
		Num:    num,
	}
	bat.Works = append(bat.Works, workers)
}

// Run is last function you need to call for actually running the task.
func (bat *Bat) Run() map[int]int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskQueue := bat.runProducers(ctx)
	resultQueue := bat.runWorks(ctx, taskQueue)
    report := bat.runReporter(ctx, resultQueue)

	return report
}

func (bat *Bat) runReporter(ctx context.Context, in <-chan context.Context) map[int]int {
    report := make(map[int]int)
	for c := range in {
        report[-1] += 1
        if err, ok := c.Value(contextError).(*Error); ok && err != nil {
            stage := len(bat.Works) - err.Count
            println("job failed at stage", stage, "with error:", err.Error)
            report[stage] += 1
        }else{
            data := c.Value(contextData).(map[string]int)
            println("age:", data["age"])
            println("money:", data["money"])
        }
	}
    return report
}

func (bat *Bat) runWorks(ctx context.Context, in <-chan context.Context) <-chan context.Context {
	upstream := in
	for _, workers := range bat.Works {
		downstream := bat.runWorkers(ctx, workers, upstream)
		upstream = downstream
	}
	return upstream
}

func (bat *Bat) runWorkers(ctx context.Context, workers Workers, upstream <-chan context.Context) <-chan context.Context {
	workerQueue := make([]<-chan context.Context, workers.Num)
	for i := 0; i < workers.Num; i++ {
		workerQueue[i] = bat.runWorker(ctx, workers.Worker, upstream)
	}
	return merge(ctx, workerQueue)
}

func (bat *Bat) runWorker(ctx context.Context, worker Worker, upstream <-chan context.Context) <-chan context.Context {
	out := make(chan context.Context, 5)
	go func() {
        defer close(out)
        for c := range upstream {
            if err, ok := c.Value(contextError).(*Error); ok && err != nil {
                err.Count += 1
                c = context.WithValue(c, contextError, err)
            }else{
                _ctx := &Context{c, nil}
                err := worker(_ctx)
                if err != nil {
                    _err := &Error{err, 0}
                    c = context.WithValue(_ctx.ctx, contextError, _err)
                }
            }
            select {
            case out <- c:
            case <-ctx.Done():
                return
            }
        }
	}()
	return out
}

func (bat *Bat) runProducers(ctx context.Context) <-chan context.Context {
	out := make(chan context.Context, len(bat.Producers))
	go func() {
		defer close(out)
		for _, producer := range bat.Producers {
			task := &Task{TaskQueue: make(chan context.Context, 5)}

			go func() {
				defer close(task.TaskQueue)
				producer(task)
			}()

			for n := range task.TaskQueue {
				select {
				case out <- n:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// Submit is a function used within a producer to send a task to workers
func (task *Task) Submit(ctx context.Context, timeout time.Duration) {
	if timeout != 0 {
		ctx = context.WithValue(ctx, contextTimeout, timeout)
	}
	task.TaskQueue <- ctx
	task.Total += 1
}

// NewContext is a function used within a producer to create a new context.
func (task *Task) NewContext(data interface{}) context.Context {
	return context.WithValue(context.Background(), contextData, data)
}

// GetContext is a function used in both producer and worker with type assertion.
// For example, if you previous create a context like: task.NewContext(map[string]int{"name": 22}), then you have to use GetContext in this way: data := ctx.GetContext().(map[string]int)
func (ctx *Context) GetContext() interface{} {
	return ctx.ctx.Value(contextData)
}

// SetContext is a function used in both producer and worker
func (ctx *Context) SetContext(data interface{}) {
	ctx.ctx = context.WithValue(ctx.ctx, contextData, data)
}

func merge(ctx context.Context, cs []<-chan context.Context) <-chan context.Context {
	var wg sync.WaitGroup
	out := make(chan context.Context, 10)

	// Start an output goroutine for each input channel in cs.
	output := func(c <-chan context.Context) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

