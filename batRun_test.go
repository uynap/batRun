package batRun_test

import (
	"github.com/uynap/batRun"
	"testing"
	"time"
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
				"id": i * 10,
			}
			ctx := task.NewContext(data)
			task.Submit(ctx, 0)
		}
	})

	bat.AddWork(func(ctx *batRun.Context) error {
		data := ctx.GetContext().(map[string]int)
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

	bat.AddWork(func(ctx *batRun.Context) error {
		data := ctx.GetContext().(map[string]int)
		if data["age"] != 22 || data["money"] != 600 {
			t.Error("context error")
		}
		ctx.SetContext(data)
		return nil
	}, 1)

	bat.Run()
}

func TestBatRunWithTimeout(t *testing.T) {
	bat := batRun.NewBat()
	bat.AddProducers(func(task *batRun.Task) {
		for i := 1; i <= 2; i++ {
			data := map[string]int{
				"id": i,
			}
			ctx := task.NewContext(data)
			task.Submit(ctx, 3*time.Second) // create a task with timeout in 3s
		}
	})

	bat.AddWork(func(ctx *batRun.Context) error {
		data := ctx.GetContext().(map[string]int)
		data["age"] = 22
		ctx.SetContext(data)
		time.Sleep(1 * time.Second)
		return nil
	}, 2)

	bat.AddWork(func(ctx *batRun.Context) error {
		data := ctx.GetContext().(map[string]int)
		data["money"] = 600
		ctx.SetContext(data)
		time.Sleep(3 * time.Second)
		return nil
	}, 4)

	bat.Run()
}
