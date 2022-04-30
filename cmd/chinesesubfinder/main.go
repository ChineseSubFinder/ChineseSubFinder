package main

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"os"
	"strconv"
)

func init() {

	log_helper.GetLogger().Infoln("ChineseSubFinder Version:", AppVersion)

	global_value.SetAppVersion(AppVersion)

	global_value.SetExtEnCode(ExtEnCode)

	if my_util.OSCheck() == false {
		log_helper.GetLogger().Panicln(`You should search runtime.GOOS in the project, Implement unimplemented function`)
	}
}

func main() {

	//// ----------------------------------------------
	//// 前置的任务，热修复、字幕修改文件名格式、提前下载好浏览器
	//pj := pre_job.NewPreJob(settings.GetSettings(), log_helper.GetLogger())
	//err := pj.HotFix().ChangeSubNameFormat().ReloadBrowser().Wait()
	//if err != nil {
	//	log_helper.GetLogger().Panicln("pre_job", err)
	//}
	//// ----------------------------------------------
	//scan, err := scan_played_video_subinfo.NewScanPlayedVideoSubInfo(*settings.GetSettings())
	//if err != nil {
	//	log_helper.GetLogger().Panicln(err)
	//}
	//bok, err := scan.GetPlayedItemsSubtitle()
	//if err != nil {
	//	log_helper.GetLogger().Panicln(err)
	//}
	//if bok == true {
	//
	//	scan.Clear()
	//
	//	err = scan.Scan()
	//	if err != nil {
	//		log_helper.GetLogger().Panicln(err)
	//	}
	//}
	// ----------------------------------------------
	fileDownloader := file_downloader.NewFileDownloader(settings.GetSettings(), log_helper.GetLogger())
	cronHelper := cron_helper.NewCronHelper(fileDownloader)
	if settings.GetSettings().UserInfo.Username == "" || settings.GetSettings().UserInfo.Password == "" {
		// 如果没有完成，那么就不开启
		log_helper.GetLogger().Infoln("Need do Setup")
	} else {
		// 是否完成了 Setup，如果完成了，那么就开启第一次的扫描
		go func() {
			log_helper.GetLogger().Infoln("Setup is Done")
			cronHelper.Start(settings.GetSettings().CommonSettings.RunScanAtStartUp)
		}()
	}

	nowPort := readCustomPortFile()
	log_helper.GetLogger().Infoln(fmt.Sprintf("WebUI will listen at 0.0.0.0:%d", nowPort))
	// 支持在外部配置特殊的端口号，以防止本地本占用了无法使用
	backend.StartBackEnd(nowPort, cronHelper)
}

func readCustomPortFile() int {
	if my_util.IsFile(customPort) == false {
		return defPort
	} else {
		bytes, err := os.ReadFile(customPort)
		if err != nil {
			log_helper.GetLogger().Errorln("ReadFile CustomPort Error", err)
			log_helper.GetLogger().Infoln("Use DefPort", defPort)
			return defPort
		}

		atoi, err := strconv.Atoi(string(bytes))
		if err != nil {
			log_helper.GetLogger().Errorln("Atoi CustomPort Error", err)
			log_helper.GetLogger().Infoln("Use DefPort", defPort)
			return defPort
		}

		log_helper.GetLogger().Infoln("Use CustomPort", atoi)
		return atoi
	}
}

/*
	使用 git tag 来做版本描述，然后在编译的时候传入版本号信息到这个变量上
*/
var AppVersion = "unknow"

// go build -ldflags="-X main.AppVersion=aabb -X main.ExtEnCode=ccdd" .
var ExtEnCode = "abcdefg1234567890"

const (
	defPort    = 19035
	customPort = "CustomPort"
)
