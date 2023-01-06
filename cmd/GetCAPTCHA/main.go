package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"

	"github.com/ChineseSubFinder/ChineseSubFinder/cmd/GetCAPTCHA/backend"
	"github.com/ChineseSubFinder/ChineseSubFinder/cmd/GetCAPTCHA/backend/config"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/notify_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func newLog() *logrus.Logger {
	var level logrus.Level
	// --------------------------------------------------
	// 之前是读取配置文件，现在改为，读取当前目录下，是否有一个特殊的文件，有则启动 Debug 日志级别
	// 那么怎么写入这个文件，就靠额外的逻辑控制了
	if pkg.IsFile(filepath.Join(pkg.ConfigRootDirFPath(), log_helper.DebugFileName)) == true {
		level = logrus.DebugLevel
	} else {
		level = logrus.InfoLevel
	}
	logger := log_helper.NewLogHelper(log_helper.LogNameGetCAPTCHA,
		pkg.ConfigRootDirFPath(),
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

		err = pkg.ClearRodTmpRootFolder()
		if err != nil {
			loggerBase.Errorln(err.Error())
			return
		}
		loggerBase.Infoln("ClearRodTmpRootFolder OK")
	}()

	loggerBase.Infoln("-----------------------------------------")

	if pkg.ReadCustomAuthFile(loggerBase) == false {
		return fmt.Errorf("ReadCustomAuthFile failed")
	}
	AuthKey := random_auth_key.AuthKey{
		BaseKey:  pkg.BaseKey(),
		AESKey16: pkg.AESKey16(),
		AESIv16:  pkg.AESIv16(),
	}
	randomAuthKey := random_auth_key.NewRandomAuthKey(5, AuthKey)
	nowAuthKey, err := randomAuthKey.GetAuthKey()
	if err != nil {
		return err
	}

	codeB64, err := backend.GetCode(loggerBase, config.GetConfig().DesURL)
	if err != nil {
		return err
	}

	err = backend.GitProcess(loggerBase, *config.GetConfig(), codeB64)
	if err != nil {
		return err
	}

	loggerBase.Infoln("try to upload code to web api")
	nowTT := time.Now()
	nowTime := nowTT.Format("2006-01-02")
	nowTimeFileNamePrix := fmt.Sprintf("%d%d%d", nowTT.Year(), nowTT.Month(), nowTT.Day())
	httpClient, err := pkg.NewHttpClient("")
	if err != nil {
		return err
	}

	loggerBase.Infoln("PostUrl:", config.GetConfig().PostUrl)

	var codeReply CodeReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+nowAuthKey).
		SetBody(CodeReq{
			UploadToken:         config.GetConfig().AuthToken,
			EnCodeString:        codeB64,
			NowTime:             nowTime,
			NowTimeFileNamePrix: nowTimeFileNamePrix,
		}).
		SetResult(&codeReply).
		Post(config.GetConfig().PostUrl)
	if err != nil {
		return err
	}

	loggerBase.Infoln("PostUrl Resp StatusCode:", resp.StatusCode())

	if codeReply.Status == 0 {
		return fmt.Errorf("codeReply.Status == 0", "codeReply.Message:", codeReply.Message)
	}

	loggerBase.Infoln("upload code to web api done")

	return nil
}

var loggerBase *logrus.Logger

type CodeReq struct {
	UploadToken         string `json:"upload_token"`
	EnCodeString        string `json:"en_code_string"`
	NowTime             string `json:"now_time"`
	NowTimeFileNamePrix string `json:"now_time_file_name_prix"`
}

type CodeReply struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
