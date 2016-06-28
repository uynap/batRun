package batRun2_test

import (
	"testing"
    batRun "github.com/uynap/batRun2"
	"time"
	"errors"
)

func TestBatRunWithoutTimeout(t *testing.T) {
    bat := batRun.NewBat()
    batRun.AddProducers(func(task *batRun.Task) {
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
report, err := batRun.Run(2
    if err != nil {
        t.Error(err)
    }
	if report.Workers.TotalSuccess != 5 {
        t.Error("expected TotalSuccess = 5")
	}
	if report.Producer.Total != 5 {
        t.Error("expected Total = 5")
	}
	realTime := report.End.Sub(report.Start).Seconds()
	if realTime < 6 {
	    t.Error("Consume", realTime, "seconds. (should be > 6s)")
	}
}

