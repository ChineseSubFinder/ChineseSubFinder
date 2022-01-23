package pre_download_process

import (
	"errors"
	commonValue "github.com/allanpk716/ChineseSubFinder/internal/common"
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/zimuku"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/hot_fix"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/something_static"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/url_connectedness_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/prometheus/common/log"
)

type PreDownloadProcess struct {
	stageName string
	gError    error

	subSupplierHub *subSupplier.SubSupplierHub
}

func NewPreDownloadProcess() *PreDownloadProcess {
	return &PreDownloadProcess{}
}

func (p *PreDownloadProcess) Init() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = stageNameInit
	// ------------------------------------------------------------------------
	// 初始化通知缓存模块
	notify_center.Notify = notify_center.NewNotifyCenter(settings.GetSettings().DeveloperSettings.BarkServerAddress)
	// 清理通知中心
	notify_center.Notify.Clear()
	// ------------------------------------------------------------------------
	// 如果是 Debug 模式，那么就需要写入特殊文件
	if settings.GetSettings().AdvancedSettings.DebugMode == true {
		err := log_helper.WriteDebugFile()
		if err != nil {
			p.gError = errors.New("log_helper.WriteDebugFile " + err.Error())
			return p
		}
	} else {
		err := log_helper.DeleteDebugFile()
		if err != nil {
			p.gError = errors.New("log_helper.DeleteDebugFile " + err.Error())
			return p
		}
	}
	// ------------------------------------------------------------------------
	// 获取验证码
	updateTimeString, code, err := something_static.GetCodeFromWeb()
	if err != nil {
		notify_center.Notify.Add("GetSubhdCode", "GetCodeFromWeb,"+err.Error())
		log_helper.GetLogger().Errorln("something_static.GetCodeFromWeb", err)
		log_helper.GetLogger().Errorln("Skip Subhd download")
		// 没有则需要清空
		commonValue.SubhdCode = ""
	} else {
		log_helper.GetLogger().Infoln("GetCode", updateTimeString, code)
		commonValue.SubhdCode = code
	}
	// ------------------------------------------------------------------------
	// 构建每个字幕站点下载者的实例
	p.subSupplierHub = subSupplier.NewSubSupplierHub(
		*settings.GetSettings(),
		zimuku.NewSupplier(*settings.GetSettings()),
		xunlei.NewSupplier(*settings.GetSettings()),
		shooter.NewSupplier(*settings.GetSettings()),
	)
	if commonValue.SubhdCode != "" {
		// 如果找到 code 了，那么就可以继续用这个实例
		p.subSupplierHub.AddSubSupplier(subhd.NewSupplier(*settings.GetSettings()))
	}

	return p
}

func (p *PreDownloadProcess) Check() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = stageNameCheck
	// ------------------------------------------------------------------------
	// 是否启用代理
	if settings.GetSettings().AdvancedSettings.ProxySettings.UseHttpProxy == false {

		log_helper.GetLogger().Infoln("UseHttpProxy = false")
		// 如果不使用代理，那么默认需要检测 baidu 的连通性
		proxySpeed, proxyStatus, err := url_connectedness_helper.UrlConnectednessTest(url_connectedness_helper.BaiduUrl, "")
		if err != nil {
			p.gError = errors.New("UrlConnectednessTest Target Site " + url_connectedness_helper.BaiduUrl + ", " + err.Error())
			return p
		} else {
			log_helper.GetLogger().Infoln("UrlConnectednessTest Target Site", url_connectedness_helper.BaiduUrl, "Speed:", proxySpeed, "Status:", proxyStatus)
		}
	} else {

		log_helper.GetLogger().Infoln("UseHttpProxy:", settings.GetSettings().AdvancedSettings.ProxySettings.HttpProxyAddress)
		// 如果使用了代理，那么默认需要检测 google 的连通性
		proxySpeed, proxyStatus, err := url_connectedness_helper.UrlConnectednessTest(url_connectedness_helper.GoogleUrl, settings.GetSettings().AdvancedSettings.ProxySettings.HttpProxyAddress)
		if err != nil {
			p.gError = errors.New("UrlConnectednessTest Target Site " + url_connectedness_helper.GoogleUrl + ", " + err.Error())
			return p
		} else {
			log_helper.GetLogger().Infoln("UrlConnectednessTest Target Site", url_connectedness_helper.GoogleUrl, "Speed:", proxySpeed, "Status:", proxyStatus)
		}
	}
	// ------------------------------------------------------------------------
	// 测试提供字幕的网站是有效的
	p.subSupplierHub.CheckSubSiteStatus()
	// ------------------------------------------------------------------------
	// 判断文件夹是否存在
	if len(settings.GetSettings().CommonSettings.MoviePaths) < 1 {
		log_helper.GetLogger().Warningln("MoviePaths not set, len == 0")
	}
	if len(settings.GetSettings().CommonSettings.SeriesPaths) < 1 {
		log_helper.GetLogger().Warningln("SeriesPaths not set, len == 0")
	}
	for i, path := range settings.GetSettings().CommonSettings.MoviePaths {
		if my_util.IsDir(path) == false {
			log_helper.GetLogger().Errorln("MovieFolder not found Index", i, "--", path)
		} else {
			log_helper.GetLogger().Infoln("MovieFolder Index", i, "--", path)
		}
	}
	for i, path := range settings.GetSettings().CommonSettings.SeriesPaths {
		if my_util.IsDir(path) == false {
			log_helper.GetLogger().Errorln("SeriesPaths not found Index", i, "--", path)
		} else {
			log_helper.GetLogger().Infoln("SeriesPaths Index", i, "--", path)
		}
	}
	// ------------------------------------------------------------------------
	// 检查、输出 Emby 文件夹的映射关系

	return p
}

func (p *PreDownloadProcess) HotFix() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = stageNameCHotFix

	defer func() {
		log.Infoln("HotFix End")
	}()

	// ------------------------------------------------------------------------
	// 开始修复
	log.Infoln("HotFix Start, wait ...")
	log.Infoln(commonValue.NotifyStringTellUserWait)
	err := hot_fix.HotFixProcess(types.HotFixParam{
		MovieRootDirs:  settings.GetSettings().CommonSettings.MoviePaths,
		SeriesRootDirs: settings.GetSettings().CommonSettings.SeriesPaths,
	})
	if err != nil {
		log.Errorln("HotFixProcess()", err)
		p.gError = err
		return p
	}

	return p
}

func (p *PreDownloadProcess) ChangeSubNameFormat() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = stageNameChangeSubNameFormat

	defer func() {
		log.Infoln("Change Sub Name Format End")
	}()
	// ------------------------------------------------------------------------
	/*
		字幕命名格式转换，需要数据库支持
		如果数据库没有记录经过转换，那么默认从 Emby 的格式作为检测的起点，转换到目标的格式
		然后需要在数据库中记录本次的转换结果
	*/
	log.Infoln("Change Sub Name Format Start...")
	log.Infoln(commonValue.NotifyStringTellUserWait)
	renameResults, err := sub_formatter.SubFormatChangerProcess(
		settings.GetSettings().CommonSettings.MoviePaths,
		settings.GetSettings().CommonSettings.SeriesPaths,
		common.FormatterName(settings.GetSettings().AdvancedSettings.SubNameFormatter))
	// 出错的文件有哪一些
	for s, i := range renameResults.ErrFiles {
		log_helper.GetLogger().Errorln("reformat ErrFile:"+s, i)
	}
	if err != nil {
		log.Errorln("SubFormatChangerProcess() Error", err)
		p.gError = err
		return p
	}

	return p
}

func (p *PreDownloadProcess) ReloadBrowser() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = stageNameReloadBrowser
	// ------------------------------------------------------------------------
	log.Infoln("ReloadBrowser Start...")
	// ReloadBrowser 提前把浏览器下载好
	rod_helper.ReloadBrowser()
	log.Infoln("ReloadBrowser End")
	return p
}

func (p *PreDownloadProcess) Wait() error {

	return errors.New(p.stageName + " " + p.gError.Error())
}

const (
	stageNameInit                = "Init"
	stageNameCheck               = "Check"
	stageNameCHotFix             = "HotFix"
	stageNameChangeSubNameFormat = "ChangeSubNameFormat"
	stageNameReloadBrowser       = "ReloadBrowser"
)
