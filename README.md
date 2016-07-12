# batRun [![Build Status]][Build] [![GoDoc]][Documentation] [![Report Status]][Report]

[Build Status]: https://img.shields.io/travis/uynap/batRun.svg?style=flat-square
[Build]: https://travis-ci.org/uynap/batRun
[GoDoc]: https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square
[Documentation]: http://godoc.org/github.com/uynap/batRun
[Report Status]: https://img.shields.io/badge/go_report-A+-brightgreen.svg?style=flat-square
[Report]: https://goreportcard.com/report/github.com/uynap/batRun

[Context]: https://godoc.org/golang.org/x/net/context

batRun is a Multitask framework using Producer Consumer pattern to run multiple works in different goroutines. The detail documentation can be found at [GoDoc](http://godoc.org/github.com/uynap/batRun).

## Overview
```
+--------------+    +---------------------+    +---------------------+
|  Producer A  |    |        | Worker 1-1 |    |        | Worker 2-1 |
+--------------+    |        | Worker 1-2 |    |        | Worker 2-2 |                 +----------+
      ...        -> | Work 1 | Worker 1-3 | -> | Work 2 | Worker 2-3 | -> ...Work N -> | Reporter |
+--------------+    |        | ...        |    |        | ...        |                 +----------+
|  Producer Z  |    |        | Worker 1-n |    |        | Worker 2-n |
+--------------+    +---------------------+    +---------------------+
```


## Terms
* Producer: The producer's job is to generate tasks, put it into the queue.
* Task: A task consists at leat one work.
* Work: A work is a job that can be run by worker(s).
* Worker: A group of workers are working together for one work. Each worker is run in a goroutine.

## Features
**Support Multiple-producer and Multiple-work chain**: As the graph shows above, batRun supports multiple producers to generate task.

**Multiple workers for each work**: You can set the quantity of workers for each work.

**Support timeout for each task**

**Support cancellation fuction for each worker**

## Usage

Basically, you only need to create at least one "producer"(`func produce(task *batRun.Task)`) and one "worker"(`func worker(ctx *batRun.Context) error`).

A quick example is:

```go
package main

import (
    "github.com/uynap/batRun"
    "time"
）

func main() {
    bat := batRun.NewBat()
    bat.AddProducers(producer)
    bat.AddStep(worker, 2)  // 2 concurrent workers for step1
    bat.AddStep(worker2, 4) // 4 concurrent workers for step2
    report := batRun.Run()
}

func produce(task *batRun.Task) {
    for i := 1; i <= 10; i++ {
        ctx := task.NewContext(map[string]int{"id": i})
        task.Submit(ctx, 4*time.Second)  // Timeout in 4 seconds
    }
}

func worker(ctx *batRun.Context) error {
    data := ctx.GetContext().(map[string]int)
    data["age"] = 34
    ctx.SetContext(data)
    return nil
}

func worker2(ctx *batRun.Context) error {
    data := ctx.GetContext().(map[string]int)
    println(data["age"])
    return nil
}
```


Installation
------------
The only dependence is [golang.org/x/net/context][Context].
According to the Go1.7's roadmap, golang.org/x/net/context will be included into Golang's core. Then there's no dependence anymore.

`$ go get github.com/uynap/batRun`


Quantity of the workers for each work
------------
```go
bat.AddWork(worker, 2) // 2 workers for work 1
```


Timeout for tasks
------------
```go
func produce(task *batRun.Task) {
    ctx := task.NewContext(map[string]int{"id": i})
    task.Submit(ctx, 4*time.Second)  // Timeout in 4 seconds
}
```
Once the timeout is reached, all the following works will be cancelled.


Create cancel functions for workers
------------
```go
bat.AddWork(func(ctx *batRun.Context) error {
    ctx.Cancel = func() {
        println("Cancel is called from work 1")
    }
    data := ctx.GetContext().(map[string]int)
    data["age"] = 22
    ctx.SetContext(data)
    return nil
}, 5)
```
When a task is cancelled, a cancel function can be called to rollback.

For example, for an account creating task, there are 3 works. 

Work 1 is "create a system directory" with a cancel function of "delete the directory". 

Work 2 is "create a group for the user" with a cancel function of "delete the group".

Work 3 is "create the user account" with a cancel function of "delete the account". 

If the task is timeout or issue found during the Work 3, 
all the cancel functions will be called in FIFO sequence.

Please refer to the [test case](batRun_test.go) for a real example.

License
------------

This project is licensed under the MIT License.

License can be found [here](LICENSE).
