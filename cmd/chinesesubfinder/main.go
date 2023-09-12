package main

import (
	"flag"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/cron_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/backend"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
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
	logger := log_helper.NewLogHelper(
		log_helper.LogNameChineseSubFinder,
		pkg.ConfigRootDirFPath(),
		level, time.Duration(7*24)*time.Hour, time.Duration(24)*time.Hour,
		settings.Get().ExperimentalFunction.ExtendLog)
	logger.AddHook(log_helper.NewLoggerHub())

	return logger
}

func init() {

	// 要先进行 flag 的读取，并且写入全局变量中，否则后续的逻辑由于顺序问题故障
	flag.Parse()
	pkg.SetLinuxConfigPathInSelfPath(*setLinuxConfigPathInSelfPathFlag)
	// 需要进行 settings 的初始化，指定初始化的目录
	settings.SetConfigRootPath(pkg.ConfigRootDirFPath())
	// 第一次运行需要清理、读取一次
	log_helper.CleanAndLoadOnceLogs()
	loggerBase = newLog()
	// --------------------------------------------------
	if strings.ToLower(LiteMode) == "true" || *setLiteModeFlag == true {
		loggerBase.Info("LiteMode is true")
		AppVersion += " Lite"
		pkg.SetLiteMode(true)
	} else {
		// 强制设置为 Lite 模式，取消 Chrome 的相关功能，交给外部的爬虫解决
		pkg.SetLiteMode(true)
	}

	loggerBase.Infoln("ChineseSubFinder Version:", AppVersion)
	pkg.SetAppVersion(AppVersion)
	pkg.SetExtEnCode(ExtEnCode)
	if pkg.ReadCustomAuthFile(loggerBase) == false {
		pkg.SetBaseKey(BaseKey)
		pkg.SetAESKey16(AESKey16)
		pkg.SetAESIv16(AESIv16)
	}
	// --------------------------------------------------
	if pkg.OSCheck() == false {
		loggerBase.Panicln(`You should search runtime.GOOS in the project, Implement unimplemented function`)
	}
	// --------------------------------------------------
	// 初始化设备的信息
	dao.UpdateInfo(AppVersion, settings.Get())

	// 砍掉，启动就进行扫描的逻辑
	settings.Get().CommonSettings.RunScanAtStartUp = false
	err := settings.Get().Save()
	if err != nil {
		loggerBase.Panicln("settings.Get().Save() err:", err)
	}
}

func main() {

	// ------------------------------------------------------------------------
	// 如果是 Debug 模式，那么就需要写入特殊文件
	{
		if settings.Get().AdvancedSettings.DebugMode == true {
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
		if pkg.LinuxConfigPathInSelfPath() != "" {

			loggerBase.Infoln("SetLinuxConfigPathInSelfPath:", pkg.LinuxConfigPathInSelfPath())

			if pkg.IsDir(pkg.LinuxConfigPathInSelfPath()) == false {
				// 如果设置了这个路径，但是不存在则会崩溃
				loggerBase.Panicln("LinuxConfigPathInSelfPath", pkg.LinuxConfigPathInSelfPath(), "is not dir")
			}
		}
	}
	// ------------------------------------------------------------------------
	// 设置接口的 API TOKEN
	{
		if settings.Get().ExperimentalFunction.ApiKeySettings.Enabled == true {
			common.SetApiToken(settings.Get().ExperimentalFunction.ApiKeySettings.Key)
		} else {
			common.SetApiToken("")
		}
		// 是否开启开发模式，跳过某些流程
		settings.Get().SpeedDevMode = false
		err := settings.Get().Save()
		if err != nil {
			loggerBase.Panicln("settings.Get().Save() err:", err)
		}
		if settings.Get().SpeedDevMode == true {
			loggerBase.Infoln("Speed Dev Mode is On")
			pkg.SetLiteMode(true)
		} else {
			loggerBase.Infoln("Speed Dev Mode is Off")
		}
	}
	// ------------------------------------------------------------------------
	// 改进为优先启动 http server，这样后面的初始化操作的进度，就不会跟之前一样，无法把进度呈现到 Web 前端给用户看
	fileDownloader := file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("local_task_queue", loggerBase),
		random_auth_key.AuthKey{
			BaseKey:  pkg.BaseKey(),
			AESKey16: pkg.AESKey16(),
			AESIv16:  pkg.AESIv16(),
		})
	// ----------------------------------------------
	// 定时任务实例
	cronHelper := cron_helper.NewCronHelper(fileDownloader)
	// 支持在外部配置特殊的端口号，以防止本地本占用了无法使用
	nowPort := pkg.ReadCustomPortFile(loggerBase)
	// 重启的信号
	restartSignal := make(chan interface{}, 1)
	defer close(restartSignal)
	bend := backend.NewBackEnd(loggerBase, cronHelper, nowPort, restartSignal)
	go bend.Restart()
	restartSignal <- 1
	// 阻塞
	select {}
}

/*
	使用 git tag 来做版本描述，然后在编译的时候传入版本号信息到这个变量上
*/
var AppVersion = "unknow"

// go build -ldflags="-X main.AppVersion=aabb -X main.ExtEnCode=ccdd" .
var ExtEnCode = "abcdefg1234567890"

// 针对制作群晖的 SPK 应用，无法写入默认的 /config 目录而给出的新的编译条件，直接指向这个目录到当前程序的目录
var setLinuxConfigPathInSelfPathFlag = flag.String("setconfigselfpath", "", "针对制作群晖的 SPK 应用，无法写入默认的 /config 目录而给出的新的编译条件，直接指向这个目录到当前程序的目录")

var setLiteModeFlag = flag.Bool("litemode", true, "设置为 Lite 模式，不启用 Chrome 相关操作")

var (
	BaseKey  = "0123456789123456789" // 基础的密钥，密钥会基于这个基础的密钥生成
	AESKey16 = "1234567890123456"    // AES密钥
	AESIv16  = "1234567890123456"    // 初始化向量
)

var LiteMode = "false" // 是否轻量级运行模式（不支持Chrome相关操作，也就是无法支持 subhd 和 zimuku 等类似需要复杂爬虫的字幕源）

var loggerBase *logrus.Logger
