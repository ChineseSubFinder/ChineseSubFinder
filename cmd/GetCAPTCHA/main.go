package main

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/cmd/GetCAPTCHA/backend"
	"github.com/allanpk716/ChineseSubFinder/cmd/GetCAPTCHA/backend/config"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

func newLog() *logrus.Logger {
	var level logrus.Level
	// --------------------------------------------------
	// 之前是读取配置文件，现在改为，读取当前目录下，是否有一个特殊的文件，有则启动 Debug 日志级别
	// 那么怎么写入这个文件，就靠额外的逻辑控制了
	if my_util.IsFile(filepath.Join(global_value.ConfigRootDirFPath(), log_helper.DebugFileName)) == true {
		level = logrus.DebugLevel
	} else {
		level = logrus.InfoLevel
	}
	logger := log_helper.NewLogHelper(log_helper.LogNameGetCAPTCHA,
		global_value.ConfigRootDirFPath(),
		level, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)
	logger.AddHook(log_helper.NewLoggerHub())

	return logger
}

func init() {
	loggerBase = newLog()
}

func main() {

	notify_center.Notify = notify_center.NewNotifyCenter(loggerBase, config.GetConfig().WhenSubSupplierInvalidWebHook)

	proxySettings := settings.ProxySettings{
		UseProxy:                 config.GetConfig().UseProxy,
		UseWhichProxyProtocol:    config.GetConfig().UseWhichProxyProtocol,
		LocalHttpProxyServerPort: config.GetConfig().LocalHttpProxyServerPort,
		InputProxyAddress:        config.GetConfig().InputProxyAddress,
		InputProxyPort:           config.GetConfig().InputProxyPort,
		NeedPWD:                  config.GetConfig().NeedPWD,
		InputProxyUsername:       config.GetConfig().InputProxyUsername,
		InputProxyPassword:       config.GetConfig().InputProxyPassword,
	}

	// 任务还没执行完，下一次执行时间到来，下一次执行就跳过不执行
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	entryID, err := c.AddFunc("@every "+config.GetConfig().EveryTime, func() {

		err := Process(&proxySettings)
		if err != nil {
			loggerBase.Errorln(err.Error())
			return
		}
	})
	if err != nil {
		loggerBase.Errorln("cron entryID:", entryID, "Error:", err)
		return
	}

	// 先执行一次
	loggerBase.Infoln("-----------------------------------------")
	loggerBase.Infoln("First Time Start")
	err = Process(&proxySettings)
	if err != nil {
		loggerBase.Errorln(err.Error())
	}

	c.Start()
	// 阻塞
	select {}
}

func Process(proxySettings *settings.ProxySettings) error {

	var err error
	notify_center.Notify.Clear()
	defer func() {

		if err != nil {
			notify_center.Notify.Add("GetSubhdCode", err.Error())
		}

		notify_center.Notify.Send()
	}()

	loggerBase.Infoln("-----------------------------------------")

	codeB64, err := backend.GetCode(loggerBase, config.GetConfig().DesURL)
	if err != nil {
		return err
	}

	err = backend.GitProcess(loggerBase, *config.GetConfig(), codeB64)
	if err != nil {
		return err
	}

	nowTT := time.Now()
	nowTime := nowTT.Format("2006-01-02")
	nowTimeFileNamePrix := fmt.Sprintf("%d%d%d", nowTT.Year(), nowTT.Month(), nowTT.Day())
	httpClient, err := my_util.NewHttpClient(proxySettings)
	if err != nil {
		return err
	}
	var codeReply CodeReply
	_, err = httpClient.R().
		SetHeader("Authorization", "beer "+config.GetConfig().AuthToken).
		SetBody(CodeReq{
			EnCodeString:        codeB64,
			NowTime:             nowTime,
			NowTimeFileNamePrix: nowTimeFileNamePrix,
		}).
		SetResult(&codeReply).
		Post(config.GetConfig().PostUrl)
	if err != nil {
		return err
	}

	return nil
}

var loggerBase *logrus.Logger

type CodeReq struct {
	EnCodeString        string `json:"en_code_string"`
	NowTime             string `json:"now_time"`
	NowTimeFileNamePrix string `json:"now_time_file_name_prix"`
}

type CodeReply struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
