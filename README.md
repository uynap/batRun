# batRun [![Build Status]][Build] [![GoDoc]][Documentation]

[Build Status]: https://img.shields.io/travis/uynap/batRun.svg
[Build]: http://www.example.com
[GoDoc]: https://img.shields.io/badge/documentation-reference-5272B4.svg
[Documentation]: https://www.py.com

batRun is a Multitask framework using Producer Consumer pattern to run multiple works in different goroutines. The detail documentation can be found at [GoDoc](https://www.godoc.com).

## Usage

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
    bat.AddStep(worker, 2) // 2 concurrent workers for step1
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
The only dependence is golang.org/x/net/context.

`$ go get github.com/uynap/batRun`


Overview
------------
+--------------+    +---------------------+    +---------------------+
|  Producer A  |    |        | Worker 1-1 |    |        | Worker 2-1 |
+--------------+    |        | Worker 1-2 |    |        | Worker 2-2 |                 +----------+
      ...        -> | Work 1 | Worker 1-3 | -> | Work 2 | Worker 2-3 | -> ...Work N -> | Reporter |
+--------------+    |        | ...        |    |        | ...        |                 +----------+
|  Producer Z  |    |        | Worker 1-n |    |        | Worker 2-n |
+--------------+    +---------------------+    +---------------------+



License
------------

This project is licensed under the MIT License.

License can be found [here](LICENSE).
