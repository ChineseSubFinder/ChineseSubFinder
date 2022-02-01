package downloader_helper

import (
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/sirupsen/logrus"
	"time"
)

type DownloaderHelper struct {
	subSupplierHub *subSupplier.SubSupplierHub
	downloader     *downloader.Downloader
	settings       settings.Settings
	logger         *logrus.Logger
}

func NewDownloaderHelper(settings settings.Settings, _subSupplierHub *subSupplier.SubSupplierHub) *DownloaderHelper {
	return &DownloaderHelper{
		subSupplierHub: _subSupplierHub,
		settings:       settings,
		logger:         log_helper.GetLogger(),
	}
}

// Start 开启任务
func (d DownloaderHelper) Start() error {
	var err error
	// 下载实例
	d.downloader, err = downloader.NewDownloader(d.subSupplierHub, sub_formatter.GetSubFormatter(d.settings.AdvancedSettings.SubNameFormatter), d.settings)
	if err != nil {
		d.logger.Errorln("NewDownloader", err)
	}
	// 最后的清理和通知统计
	defer func() {
		d.logger.Infoln("Download One End...")
		notify_center.Notify.Send()
		my_util.CloseChrome()
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

	d.logger.Infoln("Will Scan SubFixCache Folder, Clear files that are more than 7 * 24 hours old")
	// 清理多天没有使用的时间轴字幕校正缓存文件
	rootSubFixCache, err := my_util.GetRootSubFixCacheFolder()
	if err != nil {
		d.logger.Errorln("GetRootSubFixCacheFolder", err)
		return err
	}
	err = my_util.ClearIdleSubFixCacheFolder(rootSubFixCache, 7*24*time.Hour)
	if err != nil {
		d.logger.Errorln("ClearIdleSubFixCacheFolder", err)
		return err
	}

	return nil
}

// Cancel 提前取消任务的执行
func (d DownloaderHelper) Cancel() {
	d.downloader.Cancel()
}
