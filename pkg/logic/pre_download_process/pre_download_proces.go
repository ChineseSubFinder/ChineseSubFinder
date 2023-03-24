package pre_download_process

import (
	"errors"
	"fmt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/local_http_proxy_server"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/subtitle_best"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	subSupplier "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/a4k"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/assrt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/shooter"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/xunlei"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/notify_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/something_static"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/url_connectedness_helper"
	"github.com/sirupsen/logrus"
)

type PreDownloadProcess struct {
	stageName      string
	gError         error
	log            *logrus.Logger
	fileDownloader *file_downloader.FileDownloader
	SubSupplierHub *subSupplier.SubSupplierHub
}

func NewPreDownloadProcess(_fileDownloader *file_downloader.FileDownloader) *PreDownloadProcess {

	preDownloadProcess := PreDownloadProcess{
		fileDownloader: _fileDownloader,
		log:            _fileDownloader.Log,
	}
	return &preDownloadProcess
}

func (p *PreDownloadProcess) Init() *PreDownloadProcess {

	if p.gError != nil {
		p.log.Infoln("Skip PreDownloadProcess.Init()")
		return p
	}
	p.stageName = stageNameInit
	defer func() {
		p.log.Infoln("PreDownloadProcess.Init() End")
	}()
	p.log.Infoln("PreDownloadProcess.Init() Start...")

	// ------------------------------------------------------------------------
	// 初始化通知缓存模块
	notify_center.Notify = notify_center.NewNotifyCenter(p.log, settings.Get().DeveloperSettings.BarkServerAddress)
	// 清理通知中心
	notify_center.Notify.Clear()
	// ------------------------------------------------------------------------
	// 获取验证码
	if pkg.LiteMode() == false {

		nowTT := time.Now()
		nowTimeFileNamePrix := fmt.Sprintf("%d%d%d", nowTT.Year(), nowTT.Month(), nowTT.Day())
		updateTimeString, code, err := something_static.GetCodeFromWeb(p.log, nowTimeFileNamePrix, p.fileDownloader)
		if err != nil {
			notify_center.Notify.Add("GetSubhdCode", "GetCodeFromWeb,"+err.Error())
			p.log.Errorln("something_static.GetCodeFromWeb", err)
			p.log.Errorln("Skip Subhd download")
			// 没有则需要清空
			common2.SubhdCode = ""
		} else {

			// 获取到的更新时间不是当前的日期，那么本次也跳过本次
			codeTime, err := time.Parse("2006-01-02", updateTimeString)
			if err != nil {
				p.log.Errorln("something_static.GetCodeFromWeb.time.Parse", err)
				// 没有则需要清空
				common2.SubhdCode = ""
			} else {

				nowTime := time.Now()
				if codeTime.YearDay() != nowTime.YearDay() {
					// 没有则需要清空
					common2.SubhdCode = ""
					p.log.Warningln("something_static.GetCodeFromWeb, GetCodeTime:", updateTimeString, "NowTime:", time.Now().String(), "Skip")
				} else {
					p.log.Infoln("GetCode", updateTimeString, code)
					common2.SubhdCode = code
				}
			}
		}
	}
	// ------------------------------------------------------------------------
	// 构建每个字幕站点下载者的实例
	if settings.Get().SpeedDevMode == true {

		p.SubSupplierHub = subSupplier.NewSubSupplierHub(
			//csf.NewSupplier(p.fileDownloader),
			assrt.NewSupplier(p.fileDownloader),
		)
	} else {

		p.SubSupplierHub = subSupplier.NewSubSupplierHub(
			//zimuku.NewSupplier(p.fileDownloader),
			xunlei.NewSupplier(p.fileDownloader),
			shooter.NewSupplier(p.fileDownloader),
			a4k.NewSupplier(p.fileDownloader),
		)

		if settings.Get().SubtitleSources.AssrtSettings.Enabled == true &&
			settings.Get().SubtitleSources.AssrtSettings.Token != "" {
			// 如果开启了 ASSRt 字幕源，则需要新增
			p.SubSupplierHub.AddSubSupplier(assrt.NewSupplier(p.fileDownloader))
		}

		if settings.Get().SubtitleSources.SubtitleBestSettings.Enabled == true &&
			settings.Get().SubtitleSources.SubtitleBestSettings.ApiKey != "" {
			// 如果开启了 SubtitleBest 字幕源，则需要新增
			p.SubSupplierHub.AddSubSupplier(subtitle_best.NewSupplier(p.fileDownloader))
		}

		if pkg.LiteMode() == false {
			// 如果不是 Lite 模式，那么就可以开启这个功能
			if common2.SubhdCode != "" {
				// 如果找到 code 了，那么就可以继续用这个实例
				//p.SubSupplierHub.AddSubSupplier(subhd.NewSupplier(p.fileDownloader))
			}
		}
	}
	// ------------------------------------------------------------------------
	// 清理自定义的 rod 缓存目录
	err := pkg.ClearRodTmpRootFolder()
	if err != nil {
		p.gError = errors.New("ClearRodTmpRootFolder " + err.Error())
		return p
	}

	p.log.Infoln("ClearRodTmpRootFolder Done")

	return p
}

func (p *PreDownloadProcess) Check() *PreDownloadProcess {

	if p.gError != nil {
		p.log.Infoln("Skip PreDownloadProcess.Check()")
		return p
	}
	p.stageName = stageNameCheck
	defer func() {
		p.log.Infoln("PreDownloadProcess.Check() End")
	}()
	p.log.Infoln("PreDownloadProcess.Check() Start...")
	// ------------------------------------------------------------------------
	// 是否启用代理
	if settings.Get().AdvancedSettings.ProxySettings.UseProxy == false {

		p.log.Infoln("UseHttpProxy = false")
		// 如果不使用代理，那么默认需要检测 baidu 的连通性，不通过也继续
		proxyStatus, proxySpeed, err := url_connectedness_helper.UrlConnectednessTest(url_connectedness_helper.BaiduUrl, "")
		if err != nil {
			p.log.Errorln(errors.New("UrlConnectednessTest Target Site " + url_connectedness_helper.BaiduUrl + ", " + err.Error()))
		} else {
			p.log.Infoln("UrlConnectednessTest Target Site", url_connectedness_helper.BaiduUrl, "Speed:", proxySpeed, "ms,", "Status:", proxyStatus)
		}
	} else {

		p.log.Infoln("UseHttpProxy By:", settings.Get().AdvancedSettings.ProxySettings.UseWhichProxyProtocol)
		// 如果使用了代理，那么默认需要检测 google 的连通性，不通过也继续
		proxyStatus, proxySpeed, err := url_connectedness_helper.UrlConnectednessTest(url_connectedness_helper.GoogleUrl, local_http_proxy_server.GetProxyUrl())
		if err != nil {
			p.log.Errorln(errors.New("UrlConnectednessTest Target Site " + url_connectedness_helper.GoogleUrl + ", " + err.Error()))
		} else {
			p.log.Infoln("UrlConnectednessTest Target Site", url_connectedness_helper.GoogleUrl, "Speed:", proxySpeed, "ms,", "Status:", proxyStatus)
		}
	}
	// ------------------------------------------------------------------------
	// 测试提供字幕的网站是有效的，是否下载次数超限
	p.SubSupplierHub.CheckSubSiteStatus()
	// ------------------------------------------------------------------------
	// 判断文件夹是否存在
	if len(settings.Get().CommonSettings.MoviePaths) < 1 {
		p.log.Warningln("MoviePaths not set, len == 0")
	}
	if len(settings.Get().CommonSettings.SeriesPaths) < 1 {
		p.log.Warningln("SeriesPaths not set, len == 0")
	}
	for i, path := range settings.Get().CommonSettings.MoviePaths {
		if pkg.IsDir(path) == false {
			p.log.Errorln("MovieFolder not found Index", i, "--", path)
		} else {
			p.log.Infoln("MovieFolder Index", i, "--", path)
		}
	}
	for i, path := range settings.Get().CommonSettings.SeriesPaths {
		if pkg.IsDir(path) == false {
			p.log.Errorln("SeriesPaths not found Index", i, "--", path)
		} else {
			p.log.Infoln("SeriesPaths Index", i, "--", path)
		}
	}
	// ------------------------------------------------------------------------
	// 检查、输出 Emby 文件夹的映射关系

	return p
}

func (p *PreDownloadProcess) Wait() error {
	defer func() {
		p.log.Infoln("PreDownloadProcess.Wait() Done.")
	}()
	if p.gError != nil {
		outErrString := "PreDownloadProcess.Wait() Get Error, " + "stageName:" + p.stageName + " -- " + p.gError.Error()
		p.log.Errorln(outErrString)
		return errors.New(outErrString)
	} else {
		return nil
	}
}

const (
	stageNameInit  = "Init"
	stageNameCheck = "Check"
)
