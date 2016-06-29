package batRun_test

import (
	"testing"
    "github.com/uynap/batRun"
//	"time"
//	"errors"
	"fmt"
)

func TestBatRunWithoutTimeout(t *testing.T) {
    bat := batRun.NewBat()
    bat.AddProducers(func(task *batRun.Task) {
	    for i := 1; i <= 2; i++ {
			data := map[string]int{
				"id": i,
			}
			ctx := task.NewContext(data)
			task.Submit(ctx, 0)
		}
    }, func(task *batRun.Task) {
	    for i := 1; i <= 2; i++ {
			data := map[string]int{
				"id": i*10,
			}
			ctx := task.NewContext(data)
			task.Submit(ctx, 0)
		}
	})

	bat.AddWork(func(ctx *batRun.Context) error {
	    data := ctx.GetContext().(map[string]int)
//		fmt.Println("work 1 starts processing", data["id"])
        data["age"] = 22
		ctx.SetContext(data)
		return nil
	}, 2)

	bat.AddWork(func(ctx *batRun.Context) error {
	    data := ctx.GetContext().(map[string]int)
        data["money"] = 600
		ctx.SetContext(data)
		return nil
	}, 4)

    report := bat.Run()
	fmt.Printf("%#v\n", report)
//	    t.Error("Consume", realTime, "seconds. (should be > 6s)")
}

