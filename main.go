package main

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	var err error
	log = model.GetLogger()
	configViper, err = model.InitConfigure()
	if err != nil {
		log.Errorln("InitConfigure", err)
		return 
	}
	config, err = model.ReadConfig(configViper)
	if err != nil {
		log.Errorln("ReadConfig", err)
		return 
	}
}

// TODO 考虑加入 TV 相关季开播的信息读取（每一集都有对应的 nfo 文件，可以考虑从这里面读取），这样可以更加容易判断跳过老的剧集，近期下载了，字幕下载一次即可，无需反反复复

func main() {
	if log == nil {
		panic("log init error")
	}
	if configViper == nil {
		panic("init viper error")
	}
	if config == nil {
		panic("read config error")
	}
	httpProxy := config.HttpProxy
	if config.UseProxy == false {
		httpProxy = ""
	}
	// 判断文件夹是否存在
	if model.IsDir(config.MovieFolder) == false {
		log.Errorln("MovieFolder not found")
		return
	}
	log.Infoln("MovieFolder:", config.MovieFolder)

	//任务还没执行完，下一次执行时间到来，下一次执行就跳过不执行
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	entryID, err := c.AddFunc("@every " + config.EveryTime, func() {

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
	}()

	// 下载实例
	downloader := NewDownloader(common.ReqParam{
		HttpProxy:       httpProxy,
		DebugMode:       config.DebugMode,
		SaveMultiSub:    config.SaveMultiSub,
		Threads:         config.Threads,
		SubTypePriority: config.SubTypePriority,
	})

	log.Infoln("Download One Started...")
	// 开始下载
	err := downloader.DownloadSub4Movie(config.MovieFolder)
	if err != nil {
		log.Errorln("DownloadSub4Movie", err)
		return
	}

	err = downloader.DownloadSub4Series(config.SeriesFolder)
	if err != nil {
		log.Errorln("DownloadSub4Series", err)
		return
	}
}

var(
	log         *logrus.Logger
	configViper *viper.Viper
	config      *common.Config
)
