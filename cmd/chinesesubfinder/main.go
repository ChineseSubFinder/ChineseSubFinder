package main

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/robfig/cron/v3"
)

func init() {

	log_helper.GetLogger().Infoln("ChineseSubFinder Version:", AppVersion)

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
	没有很好的想法，因为喜欢使用 tag 进行版本的输出标记，但是 tag 的时候编译 docker 前确实可以修改源码替换关键词做到版本与 tag 同步变更
	但是， goreleaser 却不支持这样，会提示源码被改了，无法进行编译发布
	除非不发布、编译 Linux 和 Windows 程序，这样就能做到 tag 与 程序内部输出版本一致。
*/
var AppVersion = "unknow"
