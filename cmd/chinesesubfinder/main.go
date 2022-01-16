package main

import (
	commonValue "github.com/allanpk716/ChineseSubFinder/internal/common"
	config2 "github.com/allanpk716/ChineseSubFinder/internal/pkg/config"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/hot_fix"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/proxy_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/prometheus/common/log"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func init() {

	log.Infoln("ChineseSubFinder Version:", AppVersion)

	if my_util.OSCheck() == false {
		panic(`You should search runtime.GOOS in the project, Implement unimplemented function`)
	}
}

func main() {
	if log == nil {
		panic("log init error")
	}
	if config == nil {
		panic("read config error")
	}
	httpProxy := config.HttpProxy
	if config.UseProxy == false {
		httpProxy = ""
	}
	if config.UseProxy == false {
		log.Infoln("UseProxy = false")
	} else {
		log.Infoln("UseProxy:", httpProxy)
		proxySpeed, proxyStatus, err := proxy_helper.ProxyTest(httpProxy)
		if err != nil {
			log.Errorln("ProxyTest Target Site http://google.com", err)
			return
		} else {
			log.Infoln("ProxyTest Target Site http://google.com", "Speed:", proxySpeed, "Status:", proxyStatus)
		}
	}

	// 判断文件夹是否存在
	if my_util.IsDir(config.MovieFolder) == false {
		log.Errorln("MovieFolder not found --", config.MovieFolder)
		return
	}
	if my_util.IsDir(config.SeriesFolder) == false {
		log.Errorln("SeriesFolder not found --", config.SeriesFolder)
		return
	}
	// 读取到的文件夹信息展示
	log.Infoln("MovieFolder:", config.MovieFolder)
	log.Infoln("SeriesFolder:", config.SeriesFolder)

	// ------ Hot Fix Start ------
	// 开始修复
	log.Infoln("HotFix Start, wait ...")
	log.Infoln(commonValue.NotifyStringTellUserWait)
	err := hot_fix.HotFixProcess(types.HotFixParam{
		MovieRootDir:  config.MovieFolder,
		SeriesRootDir: config.SeriesFolder,
	})
	if err != nil {
		log.Errorln("HotFixProcess()", err)
		log.Infoln("HotFix End")
		return
	}
	log.Infoln("HotFix End")
	// ------ Hot Fix End ------

	// ------ Change SubName Format Start ------
	/*
		字幕命名格式转换，需要数据库支持
		如果数据库没有记录经过转换，那么默认从 Emby 的格式作为检测的起点，转换到目标的格式
		然后需要在数据库中记录本次的转换结果
	*/
	log.Infoln("Change Sub Name Format Start...")
	log.Infoln(commonValue.NotifyStringTellUserWait)
	renameResults, err := sub_formatter.SubFormatChangerProcess(config.MovieFolder, config.SeriesFolder, common.FormatterName(config.SubNameFormatter))
	// 出错的文件有哪一些
	for s, i := range renameResults.ErrFiles {
		log_helper.GetLogger().Errorln("reformat ErrFile:"+s, i)
	}
	if err != nil {
		log.Errorln("SubFormatChangerProcess()", err)
		return
	}

	log.Infoln("Change Sub Name Format End")
	// ------ Change SubName Format End ------

	// 初始化通知缓存模块
	notify_center.Notify = notify_center.NewNotifyCenter(config.WhenSubSupplierInvalidWebHook)

	log.Infoln("ReloadBrowser Start...")
	// ReloadBrowser 提前把浏览器下载好
	rod_helper.ReloadBrowser()
	log.Infoln("ReloadBrowser End")

	// 任务还没执行完，下一次执行时间到来，下一次执行就跳过不执行
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	entryID, err := c.AddFunc("@every "+config.EveryTime, func() {

		DownLoadStart(httpProxy)
	})
	if err != nil {
		log.Errorln("cron entryID:", entryID, "Error:", err)
		return
	}

	if config.RunAtStartup == true {
		log.Infoln("First Time Download Start")

		DownLoadStart(httpProxy)

		log.Infoln("First Time Download End")
	} else {
		log.Infoln("config.yaml set RunAtStartup: false, so will not Run At Startup, wait", config.EveryTime, "to Download")
	}

	c.Start()
	// 阻塞
	select {}
}

func DownLoadStart(httpProxy string) {

}

/*
	没有很好的想法，因为喜欢使用 tag 进行版本的输出标记，但是 tag 的时候编译 docker 前确实可以修改源码替换关键词做到版本与 tag 同步变更
	但是， goreleaser 却不支持这样，会提示源码被改了，无法进行编译发布
	除非不发布、编译 Linux 和 Windows 程序，这样就能做到 tag 与 程序内部输出版本一致。
*/
var AppVersion = "unknow"
