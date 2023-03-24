package subtitle_best

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/mix_media_info"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/series"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type Supplier struct {
	log                *logrus.Logger
	fileDownloader     *file_downloader.FileDownloader
	topic              int
	isAlive            bool
	api                *Api
	dailyDownloadCount int
	dailyDownloadLimit int
}

func NewSupplier(fileDownloader *file_downloader.FileDownloader) *Supplier {

	sup := Supplier{}
	sup.api = NewApi(headerToken, settings.Get().SubtitleSources.SubtitleBestSettings.ApiKey)
	sup.log = fileDownloader.Log
	sup.fileDownloader = fileDownloader
	sup.isAlive = true // 默认是可以使用的，如果 check 后，再调整状态
	sup.dailyDownloadCount = 0
	sup.dailyDownloadLimit = 0

	if settings.Get().AdvancedSettings.Topic != common2.DownloadSubsPerSite {
		settings.Get().AdvancedSettings.Topic = common2.DownloadSubsPerSite
	}

	return &sup
}

func (s *Supplier) CheckAlive() (bool, int64) {

	// 如果没有设置这个 API 接口，那么就任务是不可用的
	if settings.Get().SubtitleSources.SubtitleBestSettings.ApiKey == "" {
		s.isAlive = false
		return false, 0
	}

	// 计算当前时间
	startT := time.Now()

	client, err := pkg.NewHttpClient()
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Error", err)
		return false, 0
	}
	_, limitInfo, err := s.api.QueryMovieSubtitle(client, "tt00000")
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive.QueryMovieSubtitle", "Error", err)
		return false, 0
	}
	s.updateLimitInfo(limitInfo)

	s.isAlive = true
	return true, time.Since(startT).Milliseconds()
}

func (s *Supplier) IsAlive() bool {
	return s.isAlive
}

func (s *Supplier) OverDailyDownloadLimit() bool {

	// 如果没有设置这个 API 接口，那么就任务是不可用的
	if settings.Get().SubtitleSources.SubtitleBestSettings.ApiKey == "" {
		return true
	}
	// 留 5 个下载次数的余量
	if s.dailyDownloadCount >= s.dailyDownloadLimit-5 {
		return true
	}

	// 没有超出限制
	return false
}

func (s *Supplier) GetLogger() *logrus.Logger {
	return s.log
}

func (s *Supplier) GetSupplierName() string {
	return common.SubSiteSubtitleBest
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if settings.Get().SubtitleSources.SubtitleBestSettings.Enabled == false {
		return outSubInfos, nil
	}

	if settings.Get().SubtitleSources.SubtitleBestSettings.ApiKey == "" {
		return nil, errors.New("Token is empty")
	}

	return s.getSubListFromFile(filePath, true, 0, 0)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if settings.Get().SubtitleSources.SubtitleBestSettings.Enabled == false {
		return outSubInfos, nil
	}

	if settings.Get().SubtitleSources.SubtitleBestSettings.ApiKey == "" {
		return nil, errors.New("Token is empty")
	}

	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if settings.Get().SubtitleSources.SubtitleBestSettings.Enabled == false {
		return outSubInfos, nil
	}

	if settings.Get().SubtitleSources.SubtitleBestSettings.ApiKey == "" {
		return nil, errors.New("Token is empty")
	}

	return s.downloadSub4Series(seriesInfo)
}

// 更新当前的下载次数
func (s *Supplier) updateLimitInfo(limitInfo *LimitInfo) {
	s.dailyDownloadCount = limitInfo.DailyCount()
	s.dailyDownloadLimit = limitInfo.DailyLimit()
}

func (s *Supplier) downloadSub4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	var allSupplierSubInfo = make([]supplier.SubInfo, 0)

	index := 0
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		index++
		one, err := s.getSubListFromFile(episodeInfo.FileFullPath, false, episodeInfo.Season, episodeInfo.Episode)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "getSubListFromFile", episodeInfo.FileFullPath, err)
			continue
		}
		if one == nil {
			// 没有搜索到字幕
			s.log.Infoln(s.GetSupplierName(), "Not Find Sub can be download",
				episodeInfo.Title, episodeInfo.Season, episodeInfo.Episode)
			continue
		}
		// 需要赋值给字幕结构
		for i := range one {
			one[i].Season = episodeInfo.Season
			one[i].Episode = episodeInfo.Episode
		}
		allSupplierSubInfo = append(allSupplierSubInfo, one...)
	}
	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return allSupplierSubInfo, nil
}

func (s *Supplier) getSubListFromFile(videoFPath string, isMovie bool, season, episode int) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), videoFPath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), videoFPath, "Start...")

	outSubInfoList := make([]supplier.SubInfo, 0)
	mediaInfo, err := mix_media_info.GetMixMediaInfo(s.fileDownloader.MediaInfoDealers, videoFPath, isMovie)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), videoFPath, "GetMixMediaInfo", err)
		return nil, err
	}

	client, err := pkg.NewHttpClient()
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "NewHttpClient", "Error", err)
		return nil, err
	}

	var subtitle *SubtitleResponse
	var limitInfo *LimitInfo
	if isMovie == true {
		subtitle, limitInfo, err = s.api.QueryMovieSubtitle(client, mediaInfo.ImdbId)
	} else {
		subtitle, limitInfo, err = s.api.QueryTVEpsSubtitle(client, mediaInfo.ImdbId, season, episode)
	}
	if err != nil {
		return nil, err
	}
	s.updateLimitInfo(limitInfo)

	if len(subtitle.Subtitles) <= 0 {
		return nil, nil
	}

	for index, subInfo := range subtitle.Subtitles {

		var found bool
		var dSubInfo *supplier.SubInfo
		// 获取具体的下载地址
		// 这里需要先从本地的缓存判断是否已经下载过了
		found, dSubInfo, err = s.fileDownloader.CacheCenter.DownloadFileGet(subInfo.SubSha256)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "DownloadFileGet", err)
			continue
		}
		if found == false {
			// 本地没有缓存，需要从网络下载
			var downloadUrl *GetUrlResponse
			downloadUrl, limitInfo, err = s.api.GetDownloadUrl(client, subInfo.SubSha256, mediaInfo.ImdbId,
				subInfo.IsMovie, subInfo.Season, subInfo.Episode,
				"", subInfo.Language, subInfo.Token)
			if err != nil {
				return nil, err
			}
			s.updateLimitInfo(limitInfo)

			// 下载地址为空
			if len(downloadUrl.DownloadLink) < 1 {
				continue
			}
			// 这里需要注意的是 SubtitleBest 的下载地址是时效性的，所以不能以下载地址进行唯一性存储
			dSubInfo, err = s.fileDownloader.GetSubtitleBest(s.GetSupplierName(), int64(index), subInfo.Season, subInfo.Episode,
				subInfo.Title, subInfo.Ext, subInfo.SubSha256, downloadUrl.DownloadLink)
			if err != nil {
				s.log.Error("FileDownloader.Get", err)
				continue
			}
		}

		outSubInfoList = append(outSubInfoList, *dSubInfo)
		// 如果够了那么多个字幕就返回
		if len(outSubInfoList) >= settings.Get().AdvancedSettings.Topic {
			return outSubInfoList, nil
		}
	}

	return outSubInfoList, nil
}

const (
	headerToken = "5akwmGAbuFWqgaZf9QwT"
)
