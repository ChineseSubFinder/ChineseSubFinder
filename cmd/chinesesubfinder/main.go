package main

import (
	"github.com/allanpk716/ChineseSubFinder/internal"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/hot_fix"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func init() {

	if pkg.OSCheck() == false {
		panic("only support Linux and Windows, if you want support MacOS, you need implement getDbName() in file: internal/dao/init.go ")
	}

	log = log_helper.GetLogger()
	config = pkg.GetConfig()
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
	// 判断文件夹是否存在
	if pkg.IsDir(config.MovieFolder) == false {
		log.Errorln("MovieFolder not found")
		return
	}
	if pkg.IsDir(config.SeriesFolder) == false {
		log.Errorln("SeriesFolder not found")
		return
	}
	// 读取到的文件夹信息展示
	log.Infoln("MovieFolder:", config.MovieFolder)
	log.Infoln("SeriesFolder:", config.SeriesFolder)

	// ------ Hot Fix Start ------
	// 开始修复
	log.Infoln("HotFix Start...")
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

	log.Infoln("Change Sub Name Format End")
	// ------ Change SubName Format End ------

	// 初始化通知缓存模块
	notify_center.Notify = notify_center.NewNotifyCenter(config.WhenSubSupplierInvalidWebHook)

	// ReloadBrowser 提前把浏览器下载好
	rod_helper.ReloadBrowser()

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
	log.Infoln("First Time Download Start")

	DownLoadStart(httpProxy)

	log.Infoln("First Time Download End")

	c.Start()
	// 阻塞
	select {}
}

func DownLoadStart(httpProxy string) {
	defer func() {
		log.Infoln("Download One End...")
		notify_center.Notify.Send()
		pkg.CloseChrome()
	}()
	notify_center.Notify.Clear()

	// 下载实例
	downloader := internal.NewDownloader(sub_formatter.GetSubFormatter(config.SubNameFormatter),
		types.ReqParam{
			HttpProxy:                     httpProxy,
			DebugMode:                     config.DebugMode,
			SaveMultiSub:                  config.SaveMultiSub,
			Threads:                       config.Threads,
			SubTypePriority:               config.SubTypePriority,
			WhenSubSupplierInvalidWebHook: config.WhenSubSupplierInvalidWebHook,
			EmbyConfig:                    config.EmbyConfig,
			SaveOneSeasonSub:              config.SaveOneSeasonSub,
		})

	log.Infoln("Download One Started...")

	// 刷新 Emby 的字幕，如果下载了字幕倒是没有刷新，则先刷新一次，便于后续的 Emby api 统计逻辑
	err := downloader.RefreshEmbySubList()
	if err != nil {
		log.Errorln("RefreshEmbySubList", err)
		return
	}
	err = downloader.GetUpdateVideoListFromEmby(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		log.Errorln("GetUpdateVideoListFromEmby", err)
		return
	}
	// 开始下载，电影
	err = downloader.DownloadSub4Movie(config.MovieFolder)
	if err != nil {
		log.Errorln("DownloadSub4Movie", err)
		return
	}
	// 开始下载，连续剧
	err = downloader.DownloadSub4Series(config.SeriesFolder)
	if err != nil {
		log.Errorln("DownloadSub4Series", err)
		return
	}
	// 刷新 Emby 的字幕，下载完毕字幕了，就统一刷新一下
	err = downloader.RefreshEmbySubList()
	if err != nil {
		log.Errorln("RefreshEmbySubList", err)
		return
	}
}

var (
	log    *logrus.Logger
	config *types.Config
)
