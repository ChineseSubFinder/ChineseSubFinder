package csf

import (
	"errors"
	"fmt"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_file_hash"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/mix_media_info"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/sirupsen/logrus"
)

type Supplier struct {
	settings       *settings.Settings
	log            *logrus.Logger
	fileDownloader *file_downloader.FileDownloader
	topic          int
	isAlive        bool
}

func NewSupplier(fileDownloader *file_downloader.FileDownloader) *Supplier {

	sup := Supplier{}
	sup.log = fileDownloader.Log
	sup.fileDownloader = fileDownloader
	sup.topic = common2.DownloadSubsPerSite
	sup.isAlive = true // 默认是可以使用的，如果 check 后，再调整状态

	sup.settings = fileDownloader.Settings
	if sup.settings.AdvancedSettings.Topic > 0 && sup.settings.AdvancedSettings.Topic != sup.topic {
		sup.topic = sup.settings.AdvancedSettings.Topic
	}

	return &sup
}

func (s *Supplier) CheckAlive() (bool, int64) {

	// 计算当前时间
	startT := time.Now()
	_, err := s.fileDownloader.SubtitleBestApi.GetMediaInfo("tt4236770", "imdb", "series", s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Error", err)
		s.isAlive = false
		return false, 0
	}
	s.isAlive = true
	return true, time.Since(startT).Milliseconds()
}

func (s *Supplier) IsAlive() bool {
	return s.isAlive
}

func (s *Supplier) OverDailyDownloadLimit() bool {
	// 对于这个接口暂时没有限制
	return false
}

func (s *Supplier) GetLogger() *logrus.Logger {
	return s.log
}

func (s *Supplier) GetSettings() *settings.Settings {
	return s.settings
}

func (s *Supplier) GetSupplierName() string {
	return common2.SubSiteChineseSubFinder
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if s.settings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled == false {
		return outSubInfos, nil
	}

	return s.getSubListFromFile(filePath, true)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if s.settings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled == false {
		return outSubInfos, nil
	}

	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if s.settings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled == false {
		return outSubInfos, nil
	}

	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) getSubListFromFile(videoFPath string, isMovie bool) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), videoFPath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), videoFPath, "Start...")

	outSubInfoList := make([]supplier.SubInfo, 0)

}

func (s *Supplier) findAllDownloadSub(videoFPath string, isMovie bool) ([]supplier.SubInfo, error) {

	outSubInfoList := make([]supplier.SubInfo, 0)

	fileHash, err := sub_file_hash.Calculate(videoFPath)
	if err != nil {
		s.log.Errorln("scanLowVideoSubInfo.ComputeFileHash", videoFPath, err)
		return outSubInfoList, errors.New("ComputeFileHash Error")
	}
	mediaInfo, err := mix_media_info.GetMixMediaInfo(s.log, s.fileDownloader.SubtitleBestApi, videoFPath, isMovie, s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), videoFPath, "GetMixMediaInfo", err)
		return nil, err
	}

	Season := ""
	Episode := ""
	randomAuthToken := my_util.RandStringBytesMaskImprSrcSB(10)
	askFindSubReply, err := s.fileDownloader.SubtitleBestApi.AskFindSub(fileHash, mediaInfo.ImdbId, mediaInfo.TmdbId, Season, Episode, randomAuthToken, "", s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("AskFindSub Error: %s", err.Error()))
	}

	if askFindSubReply.Status == 0 {
		return nil, errors.New(fmt.Sprintf("AskFindSub Error: %s", askFindSubReply.Message))
	} else if askFindSubReply.Status == 1 {
		// 成功，查询到了字幕列表（缓存有的）
	} else if askFindSubReply.Status == 2 {
		// 放入队列，或者已经在队列中了
	} else if askFindSubReply.Status == 3 {
		// 查询的队列满了
	} else {
		// 不支持的返回值
	}

	return outSubInfoList, nil
}
