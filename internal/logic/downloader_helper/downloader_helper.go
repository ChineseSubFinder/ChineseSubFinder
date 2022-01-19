package downloader_helper

import (
	commonValue "github.com/allanpk716/ChineseSubFinder/internal/common"
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/zimuku"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/something_static"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/sirupsen/logrus"
)

type DownloaderHelper struct {
	downloader *downloader.Downloader
	settings   settings.Settings
	logger     *logrus.Logger
}

func NewDownloaderHelper(settings settings.Settings) *DownloaderHelper {
	return &DownloaderHelper{
		settings: settings,
	}
}

func (d DownloaderHelper) Start() error {
	var err error
	// 清理通知中心
	notify_center.Notify.Clear()
	// 获取验证码
	updateTimeString, code, err := something_static.GetCodeFromWeb()
	if err != nil {
		notify_center.Notify.Add("GetSubhdCode", "GetCodeFromWeb,"+err.Error())
		d.logger.Errorln("something_static.GetCodeFromWeb", err)
		d.logger.Errorln("Skip Subhd download")
		// 没有则需要清空
		commonValue.SubhdCode = ""
	} else {
		d.logger.Infoln("GetCode", updateTimeString, code)
		commonValue.SubhdCode = code
	}
	// 构建每个字幕站点下载者的实例
	var subSupplierHub = subSupplier.NewSubSupplierHub(
		//subhd.NewSupplier(d.settings),
		zimuku.NewSupplier(d.settings),
		xunlei.NewSupplier(d.settings),
		shooter.NewSupplier(d.settings),
	)
	if commonValue.SubhdCode != "" {
		// 如果找到 code 了，那么就可以继续用这个实例
		subSupplierHub.AddSubSupplier(subhd.NewSupplier(d.settings))
	}
	// 下载实例
	d.downloader, err = downloader.NewDownloader(subSupplierHub, sub_formatter.GetSubFormatter(d.settings.AdvancedSettings.SubNameFormatter), d.settings)
	if err != nil {
		d.logger.Errorln("NewDownloader", err)
	}
	// 最后的清理和通知统计
	defer func() {
		d.logger.Infoln("Download One End...")
		notify_center.Notify.Send()
		//my_util.CloseChrome()
		//rod_helper.Clear()
	}()

	d.logger.Infoln("Download One Started...")

	// 优先级最高。读取特殊文件，启用一些特殊的功能，比如 forced_scan_and_down_sub
	err = d.downloader.ReadSpeFile()
	if err != nil {
		d.logger.Errorln("ReadSpeFile", err)
	}
	// 从 csf-bk 文件还原时间轴修复前的字幕文件
	if d.downloader.NeedRestoreFixTimeLineBK == true {
		err = d.downloader.RestoreFixTimelineBK()
		if err != nil {
			d.logger.Errorln("RestoreFixTimelineBK", err)
		}
	}
	// 刷新 Emby 的字幕，如果下载了字幕倒是没有刷新，则先刷新一次，便于后续的 Emby api 统计逻辑
	err = d.downloader.RefreshEmbySubList()
	if err != nil {
		d.logger.Errorln("RefreshEmbySubList", err)
		return err
	}
	err = d.downloader.GetUpdateVideoListFromEmby()
	if err != nil {
		d.logger.Errorln("GetUpdateVideoListFromEmby", err)
		return err
	}
	// 开始下载，电影
	err = d.downloader.DownloadSub4Movie()
	if err != nil {
		d.logger.Errorln("DownloadSub4Movie", err)
		return err
	}
	// 开始下载，连续剧
	err = d.downloader.DownloadSub4Series()
	if err != nil {
		d.logger.Errorln("DownloadSub4Series", err)
		return err
	}
	// 刷新 Emby 的字幕，下载完毕字幕了，就统一刷新一下
	err = d.downloader.RefreshEmbySubList()
	if err != nil {
		d.logger.Errorln("RefreshEmbySubList", err)
		return err
	}

	return nil
}

func (d DownloaderHelper) Stop() {
	d.downloader.Cancel()
}
