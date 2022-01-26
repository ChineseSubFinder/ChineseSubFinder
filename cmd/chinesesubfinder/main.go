package main

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"os"
	"strconv"
)

func init() {

	log_helper.GetLogger().Infoln("ChineseSubFinder Version:", AppVersion)

	global_value.AppVersion = AppVersion

	if my_util.OSCheck() == false {
		panic(`You should search runtime.GOOS in the project, Implement unimplemented function`)
	}
}

func main() {

	cronHelper, err := cron_helper.NewCronHelper()
	if err != nil {
		panic("NewCronHelper " + err.Error())
	}

	if settings.GetSettings().UserInfo.Username == "" || settings.GetSettings().UserInfo.Password == "" {
		// 如果没有完成，那么就不开启
		log_helper.GetLogger().Infoln("Need do Setup")
	} else {
		// 是否完成了 Setup，如果完成了，那么就开启第一次的扫描
		go func() {
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

const (
	defPort    = 19035
	customPort = "CustomPort"
)
