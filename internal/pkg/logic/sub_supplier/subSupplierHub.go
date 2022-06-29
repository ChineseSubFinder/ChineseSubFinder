package sub_supplier

import (
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	movieHelper "github.com/allanpk716/ChineseSubFinder/internal/pkg/logic/movie_helper"
	seriesHelper "github.com/allanpk716/ChineseSubFinder/internal/pkg/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/sirupsen/logrus"
	"gopkg.in/errgo.v2/fmt/errors"
	"path/filepath"
)

type SubSupplierHub struct {
	settings  *settings.Settings
	log       *logrus.Logger
	Suppliers []ifaces.ISupplier
}

func NewSubSupplierHub(one ifaces.ISupplier, _inSupplier ...ifaces.ISupplier) *SubSupplierHub {
	s := SubSupplierHub{}
	s.settings = one.GetSettings()
	s.log = one.GetLogger()
	s.Suppliers = make([]ifaces.ISupplier, 0)
	s.Suppliers = append(s.Suppliers, one)
	if len(_inSupplier) > 0 {
		for _, supplier := range _inSupplier {
			s.Suppliers = append(s.Suppliers, supplier)
		}
	}

	return &s
}

// AddSubSupplier 添加一个下载器，目前目标是给 SubHD 使用
func (d *SubSupplierHub) AddSubSupplier(one ifaces.ISupplier) {
	d.Suppliers = append(d.Suppliers, one)
}

// DelSubSupplier 移除一个下载器
func (d *SubSupplierHub) DelSubSupplier(one ifaces.ISupplier) {

	for i := 0; i < len(d.Suppliers); i++ {

		if one.GetSupplierName() == d.Suppliers[i].GetSupplierName() {
			d.Suppliers = append(d.Suppliers[:i], d.Suppliers[i+1:]...)
		}
	}
}

// MovieNeedDlSub 电影是否符合要求需要下载字幕，比如
func (d *SubSupplierHub) MovieNeedDlSub(videoFullPath string, forcedScanAndDownloadSub bool) bool {

	if forcedScanAndDownloadSub == true {
		return true
	}

	var err error
	if d.settings.AdvancedSettings.ScanLogic.SkipChineseMovie == true {
		var skip bool
		// 跳过中文的电影，不是一定要跳过的
		skip, err = movieHelper.SkipChineseMovie(d.log, videoFullPath, d.settings.AdvancedSettings.ProxySettings)
		if err != nil {
			d.log.Warnln("SkipChineseMovie", videoFullPath, err)
		}
		if skip == true {
			return false
		}
	}

	var needDlSub = false
	if forcedScanAndDownloadSub == true {
		// 强制下载字幕
		needDlSub = true
	} else {
		needDlSub, err = movieHelper.MovieNeedDlSub(d.log, videoFullPath, d.settings.AdvancedSettings.TaskQueue.ExpirationTime)
		if err != nil {
			d.log.Errorln(errors.Newf("MovieNeedDlSub %v %v", videoFullPath, err))
			return false
		}
	}

	return needDlSub
}

// SeriesNeedDlSub 连续剧是否符合要求需要下载字幕
func (d *SubSupplierHub) SeriesNeedDlSub(seriesRootPath string, forcedScanAndDownloadSub bool, need2AnalyzeSub bool) (bool, *series.SeriesInfo, error) {

	if forcedScanAndDownloadSub == false {
		if d.settings.AdvancedSettings.ScanLogic.SkipChineseSeries == true {
			var skip bool
			var err error
			// 跳过中文的电影，不是一定要跳过的
			skip, _, err = seriesHelper.SkipChineseSeries(d.log, seriesRootPath, d.settings.AdvancedSettings.ProxySettings)
			if err != nil {
				d.log.Warnln("SkipChineseMovie", seriesRootPath, err)
			}
			if skip == true {
				return false, nil, nil
			}
		}
	}

	// 读取本地的视频和字幕信息
	seriesInfo, err := seriesHelper.ReadSeriesInfoFromDir(d.log, seriesRootPath,
		d.settings.AdvancedSettings.TaskQueue.ExpirationTime,
		forcedScanAndDownloadSub,
		need2AnalyzeSub)
	if err != nil {
		return false, nil, errors.Newf("ReadSeriesInfoFromDir %v %v", seriesRootPath, err)
	}

	return true, seriesInfo, nil
}

// SeriesNeedDlSubFromEmby 连续剧是否符合要求需要下载字幕
func (d *SubSupplierHub) SeriesNeedDlSubFromEmby(seriesRootPath string, seriesVideoList []emby.EmbyMixInfo, ExpirationTime int, skipChineseMovie, forcedScanAndDownloadSub bool) (bool, *series.SeriesInfo, error) {

	if skipChineseMovie == true {
		var skip bool
		var err error
		// 跳过中文的电影，不是一定要跳过的
		skip, _, err = seriesHelper.SkipChineseSeries(d.log, seriesRootPath, d.settings.AdvancedSettings.ProxySettings)
		if err != nil {
			d.log.Warnln("SkipChineseMovie", seriesRootPath, err)
		}
		if skip == true {
			return false, nil, nil
		}
	}
	// 读取本地的视频和字幕信息
	seriesInfo, err := seriesHelper.ReadSeriesInfoFromEmby(d.log, seriesRootPath, seriesVideoList, ExpirationTime, forcedScanAndDownloadSub, false)
	if err != nil {
		return false, nil, errors.Newf("ReadSeriesInfoFromDir %v %v", seriesRootPath, err)
	}

	return true, seriesInfo, nil
}

// DownloadSub4Movie 某一个电影字幕下载，下载完毕后，返回下载缓存每个字幕的位置，这里将只关心下载字幕，判断是否在时间范围内要不要下载不在这里判断，包括是否是中文视频的问题
func (d *SubSupplierHub) DownloadSub4Movie(videoFullPath string, index int64) ([]string, error) {

	// 下载所有字幕
	subInfos := movieHelper.OneMovieDlSubInAllSite(d.log, d.Suppliers, videoFullPath, index)
	if subInfos == nil || len(subInfos) < 1 {
		d.log.Warningln("OneMovieDlSubInAllSite.subInfos == 0, No Sub Downloaded.")
		return nil, nil
	}
	// 整理字幕，比如解压什么的
	organizeSubFiles, err := sub_helper.OrganizeDlSubFiles(d.log, filepath.Base(videoFullPath), subInfos, true)
	if err != nil {
		return nil, errors.Newf("OrganizeDlSubFiles %v %v", videoFullPath, err)
	}
	// 因为是下载电影，需要合并返回
	var outSubFileFullPathList = make([]string, 0)
	for s := range organizeSubFiles {
		outSubFileFullPathList = append(outSubFileFullPathList, organizeSubFiles[s]...)
	}

	for i, subFile := range outSubFileFullPathList {
		d.log.Debugln("OneMovieDlSubInAllSite", videoFullPath, i, "SubFileFPath:", subFile)
	}

	return outSubFileFullPathList, nil
}

// DownloadSub4Series 某一部连续剧的字幕下载，下载完毕后，返回下载缓存每个字幕的位置（通用的下载逻辑，前面把常规（没有媒体服务器模式）和 Emby 这样的模式都转换到想到的下载接口上
func (d *SubSupplierHub) DownloadSub4Series(seriesDirPath string, seriesInfo *series.SeriesInfo, index int64) (map[string][]string, error) {

	organizeSubFiles, err := d.dlSubFromSeriesInfo(seriesDirPath, index, seriesInfo)
	if err != nil {
		return nil, err
	}
	return organizeSubFiles, nil
}

// CheckSubSiteStatus 检测多个字幕提供的网站是否是有效的，是否下载次数超限
func (d *SubSupplierHub) CheckSubSiteStatus() backend.ReplyCheckStatus {

	outStatus := backend.ReplyCheckStatus{
		SubSiteStatus: make([]backend.SiteStatus, 0),
	}

	// 测试提供字幕的网站是有效的
	d.log.Infoln("Check Sub Supplier Start...")
	for _, supplier := range d.Suppliers {
		bAlive, speed := supplier.CheckAlive()
		if bAlive == false {
			d.log.Warningln(supplier.GetSupplierName(), "Check Alive = false")
		} else {
			d.log.Infoln(supplier.GetSupplierName(), "Check Alive = true, Speed =", speed, "ms")
		}

		outStatus.SubSiteStatus = append(outStatus.SubSiteStatus, backend.SiteStatus{
			Name:  supplier.GetSupplierName(),
			Valid: bAlive,
			Speed: speed,
		})
	}

	suppliersLen := len(d.Suppliers)
	for i := 0; i < suppliersLen; {

		// 网络检测是否有效，以及每次的下载次数限制检测
		if d.Suppliers[i].IsAlive() == false || d.Suppliers[i].OverDailyDownloadLimit() == true {

			d.DelSubSupplier(d.Suppliers[i])
			// 删除后，从头再来
			suppliersLen = len(d.Suppliers)
			i = 0
			continue
		}
		i++
	}

	for _, supplier := range d.Suppliers {
		if supplier.IsAlive() == true {
			d.log.Infoln("Alive Supplier:", supplier.GetSupplierName())
		}
	}

	d.log.Infoln("Check Sub Supplier End")

	return outStatus
}

func (d *SubSupplierHub) dlSubFromSeriesInfo(seriesDirPath string, index int64, seriesInfo *series.SeriesInfo) (map[string][]string, error) {
	// 下载好的字幕
	subInfos := seriesHelper.DownloadSubtitleInAllSiteByOneSeries(d.log, d.Suppliers, seriesInfo, index)
	// 整理字幕，比如解压什么的
	// 每一集 SxEx - 对应解压整理后的字幕列表

	if len(subInfos) < 1 {
		d.log.Warningln("DownloadSubtitleInAllSiteByOneSeries.subInfos == 0, No Sub Downloaded.")
	}

	organizeSubFiles, err := sub_helper.OrganizeDlSubFiles(d.log, filepath.Base(seriesDirPath), subInfos, false)
	if err != nil {
		return nil, errors.Newf("OrganizeDlSubFiles %v %v", seriesDirPath, err)
	}
	return organizeSubFiles, nil
}
