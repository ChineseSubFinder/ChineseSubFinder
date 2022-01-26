package main

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/robfig/cron/v3"
)

func init() {

	log_helper.GetLogger().Infoln("ChineseSubFinder Version:", AppVersion)

	global_value.AppVersion = AppVersion

	if my_util.OSCheck() == false {
		panic(`You should search runtime.GOOS in the project, Implement unimplemented function`)
	}
}

func main() {

	// 任务还没执行完，下一次执行时间到来，下一次执行就跳过不执行
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	entryID, err := c.AddFunc("@every "+settings.GetSettings().CommonSettings.ScanInterval, func() {
		// do something
	})
	if err != nil {
		log_helper.GetLogger().Errorln("cron entryID:", entryID, "Error:", err)
		return
	}

	if settings.GetSettings().CommonSettings.RunScanAtStartUp == true {
		log_helper.GetLogger().Infoln("First Time Download Start")
		// do something
		log_helper.GetLogger().Infoln("First Time Download End")
	} else {
		log_helper.GetLogger().Infoln("RunAtStartup: false, so will not Run At Startup, wait", settings.GetSettings().CommonSettings.ScanInterval, "to Download")
	}

	c.Start()
	// 阻塞
	select {}
}

/*
	使用 git tag 来做版本描述，然后在编译的时候传入版本号信息到这个变量上
*/
var AppVersion = "unknow"
