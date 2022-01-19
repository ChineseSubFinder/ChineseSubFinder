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
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/something_static"
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
		zimuku.NewSupplier(*settings.GetSettings()),
		xunlei.NewSupplier(*settings.GetSettings()),
		shooter.NewSupplier(*settings.GetSettings()),
	)
	if commonValue.SubhdCode != "" {
		// 如果找到 code 了，那么就可以继续用这个实例
		p.subSupplierHub.AddSubSupplier(subhd.NewSupplier(*settings.GetSettings()))
	}
	// ------------------------------------------------------------------------
	// Hot Fix Start

	// ------------------------------------------------------------------------

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
		// 如果使用了代理，那么默认需要检测 baidu 的连通性
		proxySpeed, proxyStatus, err := url_connectedness_helper.UrlConnectednessTest(url_connectedness_helper.BaiduUrl, "")
		if err != nil {
			p.gError = errors.New("UrlConnectednessTest Target Site " + url_connectedness_helper.GoogleUrl + ", " + err.Error())
			return p
		} else {
			log_helper.GetLogger().Infoln("UrlConnectednessTest Target Site", url_connectedness_helper.GoogleUrl, "Speed:", proxySpeed, "Status:", proxyStatus)
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
	log_helper.GetLogger().Infoln("Check Sub Supplier Start...")
	for _, supplier := range p.subSupplierHub.Suppliers {
		bAlive, speed := supplier.CheckAlive()
		if bAlive == false {
			log_helper.GetLogger().Warningln(supplier.GetSupplierName(), "Check Alive = false")
		} else {
			log_helper.GetLogger().Infoln(supplier.GetSupplierName(), "Check Alive = true, Speed =", speed, "ms")
		}
	}
	log_helper.GetLogger().Infoln("Check Sub Supplier End")
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
	// 输出 Emby 文件夹的映射关系

	return p
}

func (p *PreDownloadProcess) HotFix() *PreDownloadProcess {

	defer func() {
		log.Infoln("HotFix End")
	}()

	if p.gError != nil {
		return p
	}
	p.stageName = stageNameStart
	// ------------------------------------------------------------------------

	// 开始修复
	log.Infoln("HotFix Start, wait ...")
	log.Infoln(commonValue.NotifyStringTellUserWait)
	err := hot_fix.HotFixProcess(types.HotFixParam{
		MovieRootDir:  config.MovieFolder,
		SeriesRootDir: config.SeriesFolder,
	})
	if err != nil {
		log.Errorln("HotFixProcess()", err)
		p.gError = err
		return p
	}

	return p
}

func (p *PreDownloadProcess) Start() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = stageNameStart
	// ------------------------------------------------------------------------

	return p
}

func (p *PreDownloadProcess) GetResult() error {

	return errors.New(p.stageName + " " + p.gError.Error())
}

const (
	stageNameInit  = "Init"
	stageNameCheck = "Check"
	stageNameStart = "Start"
)
