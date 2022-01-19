package pre_download_process

import (
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/url_connectedness_helper"
)

type PreDownloadProcess struct {
	stageName string
	gError    error
}

func NewPreDownloadProcess() *PreDownloadProcess {
	return &PreDownloadProcess{}
}

func (p *PreDownloadProcess) Init() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = "Init"
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
	// 测试代理，同时
	if settings.GetSettings().AdvancedSettings.ProxySettings.UseHttpProxy == false {
		log_helper.GetLogger().Infoln("UseHttpProxy = false")
	} else {
		log_helper.GetLogger().Infoln("UseHttpProxy:", settings.GetSettings().AdvancedSettings.ProxySettings.HttpProxyAddress)
		proxySpeed, proxyStatus, err := url_connectedness_helper.UrlConnectednessTest(settings.GetSettings().AdvancedSettings.ProxySettings.HttpProxyAddress)
		if err != nil {
			p.gError = errors.New("UrlConnectednessTest Target Site http://google.com " + err.Error())
			return p
		} else {
			log_helper.GetLogger().Infoln("UrlConnectednessTest Target Site http://google.com", "Speed:", proxySpeed, "Status:", proxyStatus)
		}
	}
	// ------------------------------------------------------------------------
	// 判断文件夹是否存在
	if len(settings.GetSettings().CommonSettings.MoviePaths) < 1 {
		log_helper.GetLogger().Infoln("MoviePaths not set, len == 0")
	}
	if len(settings.GetSettings().CommonSettings.SeriesPaths) < 1 {
		log_helper.GetLogger().Infoln("SeriesPaths not set, len == 0")
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
	// Hot Fix Start

	// ------------------------------------------------------------------------

	return p
}

func (p *PreDownloadProcess) Start() *PreDownloadProcess {

	if p.gError != nil {
		return p
	}
	p.stageName = "Start"

	return p
}

func (p *PreDownloadProcess) Do() error {

	return errors.New(p.stageName + " " + p.gError.Error())
}
