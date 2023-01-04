package csf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/series"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/subtitle_best_api"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_file_hash"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/mix_media_info"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/sirupsen/logrus"
)

type Supplier struct {
	log            *logrus.Logger
	fileDownloader *file_downloader.FileDownloader
	isAlive        bool
}

func NewSupplier(fileDownloader *file_downloader.FileDownloader) *Supplier {

	sup := Supplier{}
	sup.log = fileDownloader.Log
	sup.fileDownloader = fileDownloader
	sup.isAlive = true // 默认是可以使用的，如果 check 后，再调整状态

	if settings.Get().AdvancedSettings.Topic != common2.DownloadSubsPerSite {
		settings.Get().AdvancedSettings.Topic = common2.DownloadSubsPerSite
	}

	return &sup
}

func (s *Supplier) CheckAlive() (bool, int64) {
	// 计算当前时间
	startT := time.Now()
	err := s.fileDownloader.MediaInfoDealers.SubtitleBestApi.CheckAlive()
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

	if settings.Get().AdvancedSettings.SuppliersSettings.ChineseSubFinder.DailyDownloadLimit == 0 {
		s.log.Warningln(s.GetSupplierName(), "DailyDownloadLimit is 0, will Skip Download")
		return true
	}

	// 对于这个接口暂时没有限制
	return false
}

func (s *Supplier) GetLogger() *logrus.Logger {
	return s.log
}

func (s *Supplier) GetSupplierName() string {
	return common2.SubSiteChineseSubFinder
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if settings.Get().ExperimentalFunction.ShareSubSettings.ShareSubEnabled == false {
		return outSubInfos, nil
	}

	return s.findAndDownload(filePath, true, 0, 0)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if settings.Get().ExperimentalFunction.ShareSubSettings.ShareSubEnabled == false {
		return outSubInfos, nil
	}

	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		oneSubInfoList, err := s.findAndDownload(episodeInfo.FileFullPath, false, episodeInfo.Season, episodeInfo.Episode)
		if err != nil {
			return outSubInfos, errors.New("FindAndDownload error:" + err.Error())
		}
		outSubInfos = append(outSubInfos, oneSubInfoList...)
	}

	return outSubInfos, nil
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if settings.Get().ExperimentalFunction.ShareSubSettings.ShareSubEnabled == false {
		return outSubInfos, nil
	}

	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		oneSubInfoList, err := s.findAndDownload(episodeInfo.FileFullPath, false, episodeInfo.Season, episodeInfo.Episode)
		if err != nil {
			return outSubInfos, errors.New("FindAndDownload error:" + err.Error())
		}
		outSubInfos = append(outSubInfos, oneSubInfoList...)
	}

	return outSubInfos, nil
}

func (s *Supplier) findAndDownload(videoFPath string, isMovie bool, Season, Episode int) (outSubInfoList []supplier.SubInfo, err error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), videoFPath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), videoFPath, "Start...")

	outSubInfoList = make([]supplier.SubInfo, 0)
	fileHash, err := sub_file_hash.Calculate(videoFPath)
	if err != nil {
		err = errors.New(fmt.Sprintf("%s.Calculate %s %s", s.GetSupplierName(), videoFPath, err))
		return
	}
	mediaInfo, err := mix_media_info.GetMixMediaInfo(s.fileDownloader.MediaInfoDealers, videoFPath, isMovie)
	if err != nil {
		err = errors.New(fmt.Sprintf("%s.GetMixMediaInfo %s %s", s.GetSupplierName(), videoFPath, err))
		return
	}

	if isMovie == true {
		Season = 0
		Episode = 0
	}

	// 标记本次请求的归属性
	randomAuthToken := pkg.RandStringBytesMaskImprSrcSB(10)
	var bestOneSub subtitle_best_api.Subtitle
	var queueIsFull bool
	var waitTime int64
	reTryTimes := 0
	const maxRetryTimes = 5
	retryFail := false
	// 重试多次排队
	for {
		reTryTimes++
		if reTryTimes > maxRetryTimes {
			// 超过了最大重试次数，直接返回
			retryFail = true
			break
		}
		bestOneSub, queueIsFull, waitTime, err = s.askFindSubProcess(fileHash, mediaInfo.ImdbId, mediaInfo.TmdbId, Season, Episode, randomAuthToken, "")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("askFindSubProcess Error: %s", err.Error()))
		}
		if queueIsFull == true {
			// 队列满了，需要等待再次重试
			time.Sleep(5 * time.Second)
			continue
		}
		// 没有错误，没有队列满，可以继续
		break
	}

	if retryFail == true {
		// 没有等待时间，直接返回
		s.log.Warningln("ask find queue is full, over max retry times, skip this time")
		return
	}

	if bestOneSub.SubSha256 != "" {
		// 说明查询的时候名中了缓存，无需往下继续查询，直接去排队下载即可
	} else {
		// 没有命中缓存，需要主动去查询
		// 等待一定的时间去查询
		if waitTime <= 0 {
			waitTime = 5
		}
		s.log.Infoln("will wait", waitTime, "s 2 ask find sub")
		// 等待耗时的动作
		var sleepCounter int64
		sleepCounter = 0
		for true {
			if sleepCounter > waitTime {
				break
			}
			if sleepCounter%30 == 0 {
				s.log.Infoln("wait 2 ask find sub")
			}
			time.Sleep(1 * time.Second)
			sleepCounter++
		}
		// 直接查询
		findSubReply, err := s.fileDownloader.MediaInfoDealers.SubtitleBestApi.FindSub(fileHash, mediaInfo.ImdbId, mediaInfo.TmdbId, strconv.Itoa(Season), strconv.Itoa(Episode), randomAuthToken, "")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("FindSub Error: %s", err.Error()))
		}
		if len(findSubReply.Subtitle) < 1 {
			s.log.Warningln("FindSub Error: no sub found")
			return outSubInfoList, nil
		}
		bestOneSub = s.findBestSub(findSubReply.Subtitle)
	}

	tmpFolder, err := pkg.GetRootTmpFolder()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetRootTmpFolder Error: %s", err.Error()))
	}
	desSubSaveFPath := filepath.Join(tmpFolder, bestOneSub.SubSha256+bestOneSub.Ext)
	foundSubCache, cacheSubInfo, err := s.fileDownloader.GetCSF(bestOneSub.SubSha256)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetCSF Error: %s", err.Error()))
	}
	if foundSubCache == true {
		// 在本地缓存中找到了
		cacheSubInfo.Season = Season
		cacheSubInfo.Episode = Episode
		outSubInfoList = append(outSubInfoList, *cacheSubInfo)
		return
	}

	// 需要从服务器拿去
	// 得到查询结果，去排队下载
	reTryTimes = 0
	retryFail = false
	// 重试多次排队
	for {
		reTryTimes++
		if reTryTimes > maxRetryTimes {
			// 超过了最大重试次数，直接返回
			retryFail = true
			break
		}
		queueIsFull, waitTime, err = s.askDownloadSubProcess(bestOneSub.SubSha256, randomAuthToken, "")
		if err != nil {
			err = errors.New(fmt.Sprintf("AskDownloadSub Error: %s", err.Error()))
			return
		}
		if queueIsFull == true {
			// 队列满了，需要等待再次重试
			time.Sleep(5 * time.Second)
			continue
		}
		// 没有错误，没有队列满，可以继续
		break
	}

	if retryFail == true {
		// 没有等待时间，直接返回
		s.log.Warningln("ask download queue is full, over max retry times, skip this time")
		return
	}

	// 等待一定的时间去查询
	if waitTime <= 0 {
		waitTime = 5
	}
	s.log.Infoln("will wait", waitTime, "s 2 ask download sub")
	// 等待耗时的动作
	var sleepCounter int64
	sleepCounter = 0
	for true {
		if sleepCounter > waitTime {
			break
		}
		if sleepCounter%30 == 0 {
			s.log.Infoln("wait 2 ask download sub")
		}
		time.Sleep(1 * time.Second)
		sleepCounter++
	}
	// 直接下载，这里需要再前面的过程中搞清楚这个字幕是什么后缀名，然后写入到缓存目录中

	downloadSubReply, err := s.fileDownloader.MediaInfoDealers.SubtitleBestApi.DownloadSub(bestOneSub.SubSha256, randomAuthToken, "", desSubSaveFPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("DownloadSub Error: %s", err.Error()))
	}

	if downloadSubReply.Status == 0 {
		// 下载失败了
		err = errors.New(fmt.Sprintf("DownloadSub Error: %s", downloadSubReply.Message))
		return
	} else if downloadSubReply.Status == 1 {
		// 下载成功
	} else {
		// 不支持的返回值
		err = errors.New(fmt.Sprintf("DownloadSub Not Support Status: %d, Message: %s", downloadSubReply.Status, downloadSubReply.Message))
		return
	}

	// 下载成功后，需要判断一下这里的文件是否是字幕文件，如果不是，需要删除掉，然后跳过
	bok := false
	var subFileInfo *subparser.FileInfo
	bok, subFileInfo, err = s.fileDownloader.SubParserHub.DetermineFileTypeFromFile(desSubSaveFPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("DetermineFileTypeFromFile Error: %s", err.Error()))
	}
	if bok == false {
		// 不是字幕文件，需要删除掉
		err = os.Remove(desSubSaveFPath)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Remove File %s Error: %s", desSubSaveFPath, err.Error()))
		}
		s.log.Errorln("CSF.Download.DetermineFileTypeFromFile == false, Remove", desSubSaveFPath)
		return
	}

	subFileName := strings.ReplaceAll(filepath.Base(videoFPath), filepath.Ext(videoFPath), "")
	subBytes, err := ioutil.ReadFile(desSubSaveFPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ReadFile Error: %s", err.Error()))
	}
	oneSubInfo := supplier.NewSubInfo(s.GetSupplierName(), 1, subFileName, subFileInfo.Lang, bestOneSub.SubSha256, 0, 0, bestOneSub.Ext, subBytes)
	oneSubInfo.Season = Season
	oneSubInfo.Episode = Episode

	err = s.fileDownloader.AddCSF(oneSubInfo)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("AddCSF Error: %s", err.Error()))
	}

	outSubInfoList = append(outSubInfoList, *oneSubInfo)

	return
}

// askFindSubProcess 查找字幕
func (s *Supplier) askFindSubProcess(VideoFeature, ImdbId, TmdbId string, Season, Episode int, FindSubToken, ApiKey string) (bestOneSub subtitle_best_api.Subtitle, queueIsFull bool, waitTime int64, err error) {

	// 开始排队查询
	askFindSubReply, err := s.fileDownloader.MediaInfoDealers.SubtitleBestApi.AskFindSub(VideoFeature, ImdbId, TmdbId, strconv.Itoa(Season), strconv.Itoa(Episode), FindSubToken, ApiKey)
	if err != nil {
		err = errors.New(fmt.Sprintf("AskFindSub Error: %s", err.Error()))
		return
	}

	if askFindSubReply.Status == 0 {
		err = errors.New(fmt.Sprintf("AskFindSub Error: %s", askFindSubReply.Message))
		return
	} else if askFindSubReply.Status == 1 {
		// 成功，查询到了字幕列表（缓存有的）
		if len(askFindSubReply.Subtitle) < 1 {
			// 查询到了，但是返回的列表是空的，也直接返回
			return
		}
		// 那么获取到了字幕列表，接下来就要去排队请求这个字幕
		/*
			每个返回的字幕中有两个特别的字段
			1. match_video_feature，是否匹配传入的视频特征
			2. low_trust，是否是低可信
		*/
		bestOneSub = s.findBestSub(askFindSubReply.Subtitle)
		return
	} else if askFindSubReply.Status == 2 {
		// 放入队列，或者已经在队列中了，根据服务器安排的时间去请求排队下载
		// 得到目标时间与当前时间的差值，单位是s
		waitTime = askFindSubReply.ScheduledUnixTime - time.Now().Unix()
		return
	} else if askFindSubReply.Status == 3 {
		// 查询的队列满了
		queueIsFull = true
		return
	} else {
		// 不支持的返回值
		err = errors.New(fmt.Sprintf("AskFindSub Not Support Status: %d, Message: %s", askFindSubReply.Status, askFindSubReply.Message))
		return
	}
}

// askDownloadSubProcess 下载排队
func (s *Supplier) askDownloadSubProcess(SubSha256, DownloadSubToken, ApiKey string) (queueIsFull bool, waitTime int64, err error) {

	askDownloadSubReply, err := s.fileDownloader.MediaInfoDealers.SubtitleBestApi.AskDownloadSub(SubSha256, DownloadSubToken, ApiKey)
	if err != nil {
		err = errors.New(fmt.Sprintf("AskDownloadSub Error: %s", err.Error()))
		return
	}
	if askDownloadSubReply.Status == 0 {
		err = errors.New(fmt.Sprintf("AskDownloadSub Error: %s", askDownloadSubReply.Message))
		return
	} else if askDownloadSubReply.Status == 2 {
		// 放入队列，或者已在队列中，根据等待的时间再去下载即可
		waitTime = askDownloadSubReply.ScheduledUnixTime - time.Now().Unix()
		return
	} else if askDownloadSubReply.Status == 3 {
		// 队列满了
		queueIsFull = true
		return
	} else {
		// 不应该出现 status = 1 的情况，这个先不改了，懒
		// 其他情况也是
		// 不支持的返回值
		err = errors.New(fmt.Sprintf("AskFindSub Not Support Status: %d, Message: %s", askDownloadSubReply.Status, askDownloadSubReply.Message))
		return
	}
}

func (s *Supplier) findBestSub(subtitles []subtitle_best_api.Subtitle) (bestOneSub subtitle_best_api.Subtitle) {
	found := false
	for _, subtitle := range subtitles {
		if subtitle.MatchVideoFeature == true && subtitle.LowTrust == false {
			// 匹配视频 且 高可信
			bestOneSub = subtitle
			found = true
			s.log.Infoln("Find Best Subtitle, MatchVideoFeature == true and HighTrust", subtitle.SubSha256)
			break
		}
	}
	if found == false {
		// 没有找到，那么就按照高可信来查询
		for _, subtitle := range subtitles {
			if subtitle.LowTrust == false {
				// 高可信
				bestOneSub = subtitle
				found = true
				s.log.Infoln("Find Best Subtitle, HighTrust", subtitle.SubSha256)
				break
			}
		}
	}
	if found == false {
		// 没有找到，那么就按照高视频匹配的来查询
		for _, subtitle := range subtitles {
			if subtitle.MatchVideoFeature == false {
				// 匹配视频
				bestOneSub = subtitle
				s.log.Infoln("Find Best Subtitle, MatchVideoFeature == true", subtitle.SubSha256)
				found = true
				break
			}
		}
	}
	if found == false {
		// 上面的都没触发，那么就返回第一个字幕吧
		bestOneSub = subtitles[0]
		s.log.Infoln("Find Best Subtitle, LowTrust", bestOneSub.SubSha256)
	}
	return
}
