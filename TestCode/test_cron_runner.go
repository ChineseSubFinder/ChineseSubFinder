package TestCode

import (
	"github.com/robfig/cron/v3"
	"time"
)

func CronRunner() {
	cronInstance := cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

	cronInstance.AddFunc("@every 2s", func() {
		println(time.Now().Format("2006-01-02 15:04:05"), "cron runner A Start")
		time.Sleep(5 * time.Second)
		println(time.Now().Format("2006-01-02 15:04:05"), "cron runner A End")

	})

	cronInstance.AddFunc("@every 5s", func() {
		println(time.Now().Format("2006-01-02 15:04:05"), "cron runner B Start")

		println(time.Now().Format("2006-01-02 15:04:05"), "cron runner B End")
	})

	cronInstance.Start()

	select {}
}
