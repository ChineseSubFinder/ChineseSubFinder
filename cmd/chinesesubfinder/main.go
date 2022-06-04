package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"

	"github.com/allanpk716/ChineseSubFinder/internal/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/pre_job"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/sirupsen/logrus"
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
	logger := log_helper.NewLogHelper(log_helper.LogNameChineseSubFinder,
		global_value.ConfigRootDirFPath(),
		level, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour)
	logger.AddHook(log_helper.NewLoggerHub())

	return logger
}

func init() {
	loggerBase = newLog()
	// --------------------------------------------------
	loggerBase.Infoln("ChineseSubFinder Version:", AppVersion)

	global_value.SetAppVersion(AppVersion)
	global_value.SetExtEnCode(ExtEnCode)
	if my_util.ReadCustomAuthFile(loggerBase) == false {
		global_value.SetBaseKey(BaseKey)
		global_value.SetAESKey16(AESKey16)
		global_value.SetAESIv16(AESIv16)
	}

	if my_util.OSCheck() == false {
		loggerBase.Panicln(`You should search runtime.GOOS in the project, Implement unimplemented function`)
	}
}

func main() {

	// ------------------------------------------------------------------------
	// 如果是 Debug 模式，那么就需要写入特殊文件
	if settings.GetSettings().AdvancedSettings.DebugMode == true {
		err := log_helper.WriteDebugFile()
		if err != nil {
			loggerBase.Errorln("log_helper.WriteDebugFile " + err.Error())
		}
		loggerBase = newLog()
		loggerBase.Infoln("Reload Log Settings, level = Debug")
	} else {
		err := log_helper.DeleteDebugFile()
		if err != nil {
			loggerBase.Errorln("log_helper.DeleteDebugFile " + err.Error())
		}
		loggerBase = newLog()
		loggerBase.Infoln("Reload Log Settings, level = Info")
	}

	// ------------------------------------------------------------------------
	// 设置接口的 API TOKEN
	if settings.GetSettings().ExperimentalFunction.ApiKeySettings.Enabled == true {
		common.SetApiToken(settings.GetSettings().ExperimentalFunction.ApiKeySettings.Key)
	} else {
		common.SetApiToken("")
	}
	// 是否开启开发模式，跳过某些流程
	//settings.GetSettings().SpeedDevMode = true
	if settings.GetSettings().SpeedDevMode == true {
		loggerBase.Infoln("Speed Dev Mode is On")
	} else {
		loggerBase.Infoln("Speed Dev Mode is Off")
	}
	if settings.GetSettings().AdvancedSettings.DebugMode == true {
		// 如果是 DebugMode 那么开启性能监控
		go func() {
			// 开启pprof，监听请求
			ip := "0.0.0.0:8080"
			if err := http.ListenAndServe(ip, nil); err != nil {
				fmt.Printf("start pprof failed on %s\n", ip)
			}
		}()
	}
	// ------------------------------------------------------------------------
	// 前置的任务，热修复、字幕修改文件名格式、提前下载好浏览器
	if settings.GetSettings().SpeedDevMode == false {
		pj := pre_job.NewPreJob(settings.GetSettings(), loggerBase)
		err := pj.HotFix().ChangeSubNameFormat().ReloadBrowser().Wait()
		if err != nil {
			loggerBase.Panicln("pre_job", err)
		}
	}
	// ----------------------------------------------
	fileDownloader := file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("local_task_queue", settings.GetSettings(), loggerBase),
		random_auth_key.AuthKey{
			BaseKey:  global_value.BaseKey(),
			AESKey16: global_value.AESKey16(),
			AESIv16:  global_value.AESIv16(),
		})
	// ----------------------------------------------
	cronHelper := cron_helper.NewCronHelper(fileDownloader)
	if settings.GetSettings().UserInfo.Username == "" || settings.GetSettings().UserInfo.Password == "" {
		// 如果没有完成，那么就不开启
		loggerBase.Infoln("Need do Setup")
	} else {
		// 是否完成了 Setup，如果完成了，那么就开启第一次的扫描
		go func() {
			loggerBase.Infoln("Setup is Done")
			cronHelper.Start(settings.GetSettings().CommonSettings.RunScanAtStartUp)
		}()
	}

	nowPort := my_util.ReadCustomPortFile(loggerBase)
	loggerBase.Infoln(fmt.Sprintf("WebUI will listen at 0.0.0.0:%d", nowPort))
	// 支持在外部配置特殊的端口号，以防止本地本占用了无法使用
	backend.StartBackEnd(fileDownloader, nowPort, cronHelper)
}

/*
	使用 git tag 来做版本描述，然后在编译的时候传入版本号信息到这个变量上
*/
var AppVersion = "unknow"

// go build -ldflags="-X main.AppVersion=aabb -X main.ExtEnCode=ccdd" .
var ExtEnCode = "abcdefg1234567890"

var (
	BaseKey  = "0123456789123456789" // 基础的密钥，密钥会基于这个基础的密钥生成
	AESKey16 = "1234567890123456"    // AES密钥
	AESIv16  = "1234567890123456"    // 初始化向量
)

var loggerBase *logrus.Logger
