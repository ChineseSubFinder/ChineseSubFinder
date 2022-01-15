package downloader_helper

import (
	commonValue "github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/something_static"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/sirupsen/logrus"
)

type DownloaderHelper struct {
	downloader *downloader.Downloader
	logger     *logrus.Logger
}

func NewDownloaderHelper() *DownloaderHelper {
	return &DownloaderHelper{}
}

func (d DownloaderHelper) Start(config types.Config) error {
	var err error
	// 清理通知中心
	notify_center.Notify.Clear()

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

	// 下载实例
	d.downloader, err = downloader.NewDownloader(sub_formatter.GetSubFormatter(config.SubNameFormatter),
		types.ReqParam{
			HttpProxy:                     config.HttpProxy,
			DebugMode:                     config.DebugMode,
			SaveMultiSub:                  config.SaveMultiSub,
			Threads:                       config.Threads,
			SubTypePriority:               config.SubTypePriority,
			WhenSubSupplierInvalidWebHook: config.WhenSubSupplierInvalidWebHook,
			EmbyConfig:                    config.EmbyConfig,
			SaveOneSeasonSub:              config.SaveOneSeasonSub,

			SubTimelineFixerConfig: config.SubTimelineFixerConfig,
			FixTimeLine:            config.FixTimeLine,
		})

	if err != nil {
		d.logger.Errorln("NewDownloader", err)
	}

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
		err = d.downloader.RestoreFixTimelineBK(config.MovieFolder, config.SeriesFolder)
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
	err = d.downloader.GetUpdateVideoListFromEmby(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		d.logger.Errorln("GetUpdateVideoListFromEmby", err)
		return err
	}
	// 开始下载，电影
	err = d.downloader.DownloadSub4Movie(config.MovieFolder)
	if err != nil {
		d.logger.Errorln("DownloadSub4Movie", err)
		return err
	}
	// 开始下载，连续剧
	err = d.downloader.DownloadSub4Series(config.SeriesFolder)
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
