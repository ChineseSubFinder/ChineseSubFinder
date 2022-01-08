package main

import (
	"github.com/allanpk716/ChineseSubFinder/cmd/GetCAPTCHA/backend"
	"github.com/allanpk716/ChineseSubFinder/cmd/GetCAPTCHA/backend/config"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/robfig/cron/v3"
)

func main() {

	notify_center.Notify = notify_center.NewNotifyCenter(config.GetConfig().WhenSubSupplierInvalidWebHook)

	// 任务还没执行完，下一次执行时间到来，下一次执行就跳过不执行
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	entryID, err := c.AddFunc("@every "+config.GetConfig().EveryTime, func() {

		err := Process()
		if err != nil {
			log_helper.GetLogger().Errorln(err.Error())
			return
		}
	})
	if err != nil {
		log_helper.GetLogger().Errorln("cron entryID:", entryID, "Error:", err)
		return
	}
	// 先执行一次
	err = Process()
	if err != nil {
		log_helper.GetLogger().Errorln(err.Error())
		return
	}

	c.Start()
	// 阻塞
	select {}
}

func Process() error {

	var err error
	notify_center.Notify.Clear()
	defer func() {

		if err != nil {
			notify_center.Notify.Add("GetSubhdCode", err.Error())
		}

		notify_center.Notify.Send()
	}()

	log_helper.GetLogger().Infoln("-----------------------------------------")

	codeB64, err := backend.GetCode(config.GetConfig().DesURL)
	if err != nil {
		return err
	}

	err = backend.GitProcess(*config.GetConfig(), codeB64)
	if err != nil {
		return err
	}

	return nil
}
