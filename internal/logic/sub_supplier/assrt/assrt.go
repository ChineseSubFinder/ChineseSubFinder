package assrt

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/imdb_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/sirupsen/logrus"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
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

	// 如果没有设置这个 API 接口，那么就任务是不可用的
	if s.settings.SubtitleSources.AssrtSettings.Token == "" {
		s.isAlive = false
		return false, 0
	}

	// 计算当前时间
	startT := time.Now()
	userInfo, err := s.getUserInfo()
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Error", err)
		s.isAlive = false
		return false, 0
	}
	s.log.Infoln(s.GetSupplierName(), "CheckAlive", "UserInfo.Status:", userInfo.Status, "UserInfo.Quota:", userInfo.User.Quota)
	// 计算耗时
	s.isAlive = true
	return true, time.Since(startT).Milliseconds()
}

func (s *Supplier) IsAlive() bool {

	if s.settings.SubtitleSources.AssrtSettings.Token == "" {
		s.log.Debugln(s.GetSupplierName(), "IsAlive", "Token is empty")
		return false
	}

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
	return common2.SubSiteAssrt
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {
	return s.getSubListFromFile(filePath, true)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) getSubListFromFile(videoFPath string, isMovie bool) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), videoFPath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), videoFPath, "Start...")

	outSubInfoList := make([]supplier.SubInfo, 0)

	imdbInfo, err := imdb_helper.GetIMDBInfo(s.log, videoFPath, isMovie, s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), videoFPath, "GetIMDBInfo", err)
		return nil, err
	}

	// 需要找到中文名称去搜索
	keyWord := imdbInfo.GetChineseNameFromAKA()
	if keyWord == "" {
		return nil, errors.New("No Chinese name")
	}
	if isMovie == false {
		// 连续剧需要额外补充 S01E01 这样的信息
		infoFromFileName, err := decode.GetVideoInfoFromFileName(videoFPath)
		if err != nil {
			return nil, err
		}
		keyWord += " " + my_util.GetEpisodeKeyName(infoFromFileName.Season, infoFromFileName.Episode, true)
	}

	var searchSubResult SearchSubResult
	searchSubResult, err = s.getSubByKeyWord(keyWord)
	if err != nil {
		s.log.Errorln("getSubByKeyWord", err)
		return nil, err
	}

	videoFileName := filepath.Base(videoFPath)

	for index, subInfo := range searchSubResult.Sub.Subs {

		// 获取具体的下载地址
		oneSubDetail, err := s.getSubDetail(subInfo.Id)
		if err != nil {
			s.log.Errorln("getSubDetail", err)
			continue
		}

		if len(oneSubDetail.Sub.Subs) < 1 {
			continue
		}
		// 这里需要注意的是 ASSRT 说明了，下载的地址是有时效性的，那么如果缓存整个地址则不是正确的
		// 需要缓存的应该是这个字幕的 ID
		nowSubDownloadUrl := oneSubDetail.Sub.Subs[0].Url
		subInfo, err := s.fileDownloader.Get(s.GetSupplierName(), int64(index), videoFileName, nowSubDownloadUrl,
			0, 0,
			// 得到一个特殊的替代 FileDownloadUrl 的特征字符串
			fmt.Sprintf("%s-%s-%d", s.GetSupplierName(), subInfo.NativeName, subInfo.Id),
		)
		if err != nil {
			s.log.Error("FileDownloader.Get", err)
			continue
		}

		outSubInfoList = append(outSubInfoList, *subInfo)
		// 如果够了那么多个字幕就返回
		if len(outSubInfoList) >= s.topic {
			return outSubInfoList, nil
		}

		if index >= 1 {
			// 因为每分钟只有 5 次的限额
			time.Sleep(20 * time.Second)
		}
	}

	return outSubInfoList, nil
}

func (s *Supplier) downloadSub4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	var allSupplierSubInfo = make([]supplier.SubInfo, 0)

	index := 0
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		index++
		one, err := s.getSubListFromFile(episodeInfo.FileFullPath, false)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "getSubListFromFile", episodeInfo.FileFullPath)
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

func (s *Supplier) getSubByKeyWord(keyword string) (SearchSubResult, error) {

	var searchSubResult SearchSubResult

	tt := url.QueryEscape(keyword)
	println(tt)

	httpClient, err := my_util.NewHttpClient(s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		return searchSubResult, err
	}
	resp, err := httpClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetResult(&searchSubResult).
		Get(s.settings.AdvancedSettings.SuppliersSettings.Assrt.RootUrl +
			"/sub/search?q=" + tt +
			"&cnt=15&pos=0" +
			"&token=" + s.settings.SubtitleSources.AssrtSettings.Token)
	if err != nil {
		if resp != nil {
			s.log.Errorln(s.GetSupplierName(), "NewHttpClient:", keyword, err.Error())
			notify_center.Notify.Add(s.GetSupplierName()+" NewHttpClient", fmt.Sprintf("keyword: %s, resp: %s, error: %s", keyword, resp.String(), err.Error()))
		}
		return searchSubResult, err
	}

	return searchSubResult, nil
}

func (s *Supplier) getSubDetail(subID int) (OneSubDetail, error) {
	var subDetail OneSubDetail

	httpClient, err := my_util.NewHttpClient(s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		return subDetail, err
	}
	resp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"token": s.settings.SubtitleSources.AssrtSettings.Token,
			"id":    strconv.Itoa(subID),
		}).
		SetResult(&subDetail).
		Get(s.settings.AdvancedSettings.SuppliersSettings.Assrt.RootUrl + "/sub/detail")
	if err != nil {
		if resp != nil {
			s.log.Errorln(s.GetSupplierName(), "NewHttpClient:", subID, err.Error())
			notify_center.Notify.Add(s.GetSupplierName()+" NewHttpClient", fmt.Sprintf("subID: %d, resp: %s, error: %s", subID, resp.String(), err.Error()))
		}
		return subDetail, err
	}

	return subDetail, nil
}

func (s *Supplier) getUserInfo() (UserInfo, error) {

	var userInfo UserInfo

	httpClient, err := my_util.NewHttpClient(s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		return userInfo, err
	}
	resp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"token": s.settings.SubtitleSources.AssrtSettings.Token,
		}).
		SetResult(&userInfo).
		Get(s.settings.AdvancedSettings.SuppliersSettings.Assrt.RootUrl + "/user/quota")
	if err != nil {
		if resp != nil {
			s.log.Errorln(s.GetSupplierName(), "NewHttpClient:", err.Error())
			notify_center.Notify.Add(s.GetSupplierName()+" NewHttpClient", fmt.Sprintf("resp: %s, error: %s", resp.String(), err.Error()))
		}
		return userInfo, err
	}

	return userInfo, nil
}

type SearchSubResult struct {
	Sub struct {
		Action string `json:"action"`
		Subs   []struct {
			Lang struct {
				Desc     string `json:"desc"`
				Langlist struct {
					Langcht bool `json:"langcht"`
					Langdou bool `json:"langdou"`
					Langeng bool `json:"langeng"`
					Langchs bool `json:"langchs"`
				} `json:"langlist"`
			} `json:"lang"`
			Id          int    `json:"id"`
			VoteScore   int    `json:"vote_score"`
			Videoname   string `json:"videoname"`
			ReleaseSite string `json:"release_site"`
			Revision    int    `json:"revision"`
			Subtype     string `json:"subtype"`
			NativeName  string `json:"native_name"`
			UploadTime  string `json:"upload_time"`
		} `json:"subs"`
		Result  string `json:"result"`
		Keyword string `json:"keyword"`
	} `json:"sub"`
	Status int `json:"status"`
}

type OneSubDetail struct {
	Sub struct {
		Action string `json:"action"`
		Subs   []struct {
			DownCount int `json:"down_count"`
			ViewCount int `json:"view_count"`
			Lang      struct {
				Desc     string `json:"desc"`
				Langlist struct {
					Langcht bool `json:"langcht"`
					Langdou bool `json:"langdou"`
					Langeng bool `json:"langeng"`
					Langchs bool `json:"langchs"`
				} `json:"langlist"`
			} `json:"lang"`
			Size       int    `json:"size"`
			Title      string `json:"title"`
			Videoname  string `json:"videoname"`
			Revision   int    `json:"revision"`
			NativeName string `json:"native_name"`
			UploadTime string `json:"upload_time"`
			Producer   struct {
				Producer string `json:"producer"`
				Verifier string `json:"verifier"`
				Uploader string `json:"uploader"`
				Source   string `json:"source"`
			} `json:"producer"`
			Subtype     string `json:"subtype"`
			VoteScore   int    `json:"vote_score"`
			ReleaseSite string `json:"release_site"`
			Filelist    []struct {
				S   string `json:"s"`
				F   string `json:"f"`
				Url string `json:"url"`
			} `json:"filelist"`
			Id       int    `json:"id"`
			Filename string `json:"filename"`
			Url      string `json:"url"`
		} `json:"subs"`
		Result string `json:"result"`
	} `json:"sub"`
	Status int `json:"status"`
}

type UserInfo struct {
	User struct {
		Action string `json:"action"`
		Result string `json:"result"`
		Quota  int    `json:"quota"`
	} `json:"user"`
	Status int `json:"status"`
}
