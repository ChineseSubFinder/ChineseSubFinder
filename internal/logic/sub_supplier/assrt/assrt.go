package assrt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/models"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/mix_media_info"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/sirupsen/logrus"
)

type Supplier struct {
	settings          *settings.Settings
	log               *logrus.Logger
	fileDownloader    *file_downloader.FileDownloader
	topic             int
	isAlive           bool
	theSearchInterval time.Duration
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

	sup.theSearchInterval = 20 * time.Second

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

	outSubInfos := make([]supplier.SubInfo, 0)
	if s.settings.SubtitleSources.AssrtSettings.Enabled == false {
		return outSubInfos, nil
	}

	if s.settings.SubtitleSources.AssrtSettings.Token == "" {
		return nil, errors.New("Token is empty")
	}

	return s.getSubListFromFile(filePath, true)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if s.settings.SubtitleSources.AssrtSettings.Enabled == false {
		return outSubInfos, nil
	}

	if s.settings.SubtitleSources.AssrtSettings.Token == "" {
		return nil, errors.New("Token is empty")
	}

	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	if s.settings.SubtitleSources.AssrtSettings.Enabled == false {
		return outSubInfos, nil
	}

	if s.settings.SubtitleSources.AssrtSettings.Token == "" {
		return nil, errors.New("Token is empty")
	}

	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) getSubListFromFile(videoFPath string, isMovie bool) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), videoFPath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), videoFPath, "Start...")

	outSubInfoList := make([]supplier.SubInfo, 0)
	mediaInfo, err := mix_media_info.GetMixMediaInfo(s.log, s.fileDownloader.SubtitleBestApi, videoFPath, isMovie, s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), videoFPath, "GetMixMediaInfo", err)
		return nil, err
	}
	// 需要找到中文名称去搜索，找不到就是用英文名称，还找不到就是 OriginalTitle
	found, searchSubResult, err := s.getSubInfoEx(mediaInfo, videoFPath, isMovie, "cn")
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), videoFPath, "GetSubInfoEx", err)
		return nil, err
	}
	if found == false {
		// 没有找到中文名称，就用英文名称去搜索
		found, searchSubResult, err = s.getSubInfoEx(mediaInfo, videoFPath, isMovie, "en")
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), videoFPath, "GetSubInfoEx", err)
			return nil, err
		}
		if found == false {
			// 没有找到英文名称，就用原名称去搜索
			found, searchSubResult, err = s.getSubInfoEx(mediaInfo, videoFPath, isMovie, "org")
			if err != nil {
				s.log.Errorln(s.GetSupplierName(), videoFPath, "GetSubInfoEx", err)
				return nil, err
			}
			if found == false {
				return nil, nil
			}
		}
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
	}

	return outSubInfoList, nil
}

func (s *Supplier) getSubInfoEx(mediaInfo *models.MediaInfo, videoFPath string, isMovie bool, keyWordType string) (bool, *SearchSubResult, error) {

	var searchSubResult *SearchSubResult
	var err error
	keyWord, err := mix_media_info.KeyWordSelect(mediaInfo, videoFPath, isMovie, keyWordType)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), videoFPath, "keyWordSelect", err)
		return false, searchSubResult, err
	}
	searchSubResult, err = s.getSubByKeyWord(keyWord)
	if err != nil {
		s.log.Errorln("getSubByKeyWord", err)
		return false, searchSubResult, err
	}

	videoFileName := filepath.Base(videoFPath)
	if searchSubResult.Sub.Subs == nil || len(searchSubResult.Sub.Subs) == 0 {
		s.log.Infoln(s.GetSupplierName(), videoFileName, "No subtitle found", "KeyWord:", keyWord)
		return false, searchSubResult, nil
	} else {
		return true, searchSubResult, nil
	}
}

func (s *Supplier) downloadSub4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	var allSupplierSubInfo = make([]supplier.SubInfo, 0)

	index := 0
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		index++
		one, err := s.getSubListFromFile(episodeInfo.FileFullPath, false)
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

func (s *Supplier) getSubByKeyWord(keyword string) (*SearchSubResult, error) {

	defer func() {
		time.Sleep(s.theSearchInterval)
	}()

	var searchSubResult SearchSubResult

	s.log.Infoln("Search KeyWord:", keyword)
	tt := url.QueryEscape(keyword)
	httpClient, err := my_util.NewHttpClient(s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		return nil, err
	}
	var errKnow error
	resp, err := httpClient.R().
		Get(s.settings.AdvancedSettings.SuppliersSettings.Assrt.RootUrl +
			"/sub/search?q=" + tt +
			"&cnt=15&pos=0" +
			"&token=" + s.settings.SubtitleSources.AssrtSettings.Token)
	if err != nil {
		return nil, err
	}
	/*
		这里有个梗， Sub 有值的时候是一个列表，但是如果为空的时候，又是一个空的结构体
		所以出现两个结构体需要去尝试解析
		SearchSubResultEmpty
		SearchSubResult
		比如这个情况：
		jsonString := "{\"sub\":{\"action\":\"search\",\"subs\":{},\"result\":\"succeed\",\"keyword\":\"追杀夏娃 S04E07\"},\"status\":0}"
	*/
	err = json.Unmarshal([]byte(resp.String()), &searchSubResult)
	if err != nil {
		// 再此尝试解析空列表
		var searchSubResultEmpty SearchSubResultEmpty
		err = json.Unmarshal([]byte(resp.String()), &searchSubResultEmpty)
		if err != nil {
			// 如果还是解析错误，那么就要把现在的错误和上面的错误仪器返回出去
			s.log.Errorln(s.GetSupplierName(), "NewHttpClient:", keyword, errKnow.Error())
			s.log.Errorln(s.GetSupplierName(), "json.Unmarshal", err)
			notify_center.Notify.Add(s.GetSupplierName()+" NewHttpClient", fmt.Sprintf("keyword: %s, resp: %s, error: %s", keyword, resp.String(), errKnow.Error()))
			return nil, err
		}
		// 赋值过去
		searchSubResult.Sub.Action = searchSubResultEmpty.Sub.Action
		searchSubResult.Sub.Result = searchSubResultEmpty.Sub.Result
		searchSubResult.Sub.Keyword = searchSubResultEmpty.Sub.Keyword
		searchSubResult.Status = searchSubResultEmpty.Status

		return &searchSubResult, nil
	}

	return &searchSubResult, nil
}

func (s *Supplier) getSubDetail(subID int) (OneSubDetail, error) {

	defer func() {
		time.Sleep(s.theSearchInterval)
	}()

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

			// 输出调试文件
			cacheCenterFolder, err := my_folder.GetRootCacheCenterFolder()
			if err != nil {
				s.log.Errorln(s.GetSupplierName(), "GetRootCacheCenterFolder", err)
			}
			desJsonInfo := filepath.Join(cacheCenterFolder, strconv.Itoa(subID)+"--assrt_search_error_getSubDetail.json")
			// 写字符串到文件种
			file, _ := os.Create(desJsonInfo)
			defer func() {
				_ = file.Close()
			}()
			file.WriteString(resp.String())
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

type SearchSubResultEmpty struct {
	Sub struct {
		Action string `json:"action"`
		Subs   struct {
		} `json:"subs"`
		Result  string `json:"result"`
		Keyword string `json:"keyword"`
	} `json:"sub"`
	Status int `json:"status"`
}

type SearchSubResult struct {
	Sub struct {
		Action string `json:"action"`
		Subs   []struct {
			Lang struct {
				Desc     string `json:"desc,omitempty"`
				Langlist struct {
					Langcht bool `json:"langcht,omitempty"`
					Langdou bool `json:"langdou,omitempty"`
					Langeng bool `json:"langeng,omitempty"`
					Langchs bool `json:"langchs,omitempty"`
				} `json:"langlist,omitempty"`
			} `json:"lang,omitempty"`
			Id          int    `json:"id,omitempty"`
			VoteScore   int    `json:"vote_score,omitempty"`
			Videoname   string `json:"videoname,omitempty"`
			ReleaseSite string `json:"release_site,omitempty"`
			Revision    int    `json:"revision,omitempty"`
			Subtype     string `json:"subtype,omitempty"`
			NativeName  string `json:"native_name,omitempty"`
			UploadTime  string `json:"upload_time,omitempty"`
		} `json:"subs,omitempty"`
		Result  string `json:"result,omitempty"`
		Keyword string `json:"keyword,omitempty"`
	} `json:"sub,omitempty"`
	Status int `json:"status,omitempty"`
}

type OneSubDetail struct {
	Sub struct {
		Action string `json:"action"`
		Subs   []struct {
			DownCount int `json:"down_count,omitempty"`
			ViewCount int `json:"view_count,omitempty"`
			Lang      struct {
				Desc     string `json:"desc,omitempty"`
				Langlist struct {
					Langcht bool `json:"langcht,omitempty"`
					Langdou bool `json:"langdou,omitempty"`
					Langeng bool `json:"langeng,omitempty"`
					Langchs bool `json:"langchs,omitempty"`
				} `json:"langlist,omitempty"`
			} `json:"lang,omitempty"`
			Size       int    `json:"size,omitempty"`
			Title      string `json:"title,omitempty"`
			Videoname  string `json:"videoname,omitempty"`
			Revision   int    `json:"revision,omitempty"`
			NativeName string `json:"native_name,omitempty"`
			UploadTime string `json:"upload_time,omitempty"`
			Producer   struct {
				Producer string `json:"producer,omitempty"`
				Verifier string `json:"verifier,omitempty"`
				Uploader string `json:"uploader,omitempty"`
				Source   string `json:"source,omitempty"`
			} `json:"producer,omitempty"`
			Subtype     string `json:"subtype,omitempty"`
			VoteScore   int    `json:"vote_score,omitempty"`
			ReleaseSite string `json:"release_site,omitempty"`
			Filelist    []struct {
				S   string `json:"s,omitempty"`
				F   string `json:"f,omitempty"`
				Url string `json:"url,omitempty"`
			} `json:"filelist,omitempty"`
			Id       int    `json:"id,omitempty"`
			Filename string `json:"filename,omitempty"`
			Url      string `json:"url,omitempty"`
		} `json:"subs,omitempty"`
		Result string `json:"result,omitempty"`
	} `json:"sub,omitempty"`
	Status int `json:"status,omitempty"`
}

type UserInfo struct {
	User struct {
		Action string `json:"action,omitempty"`
		Result string `json:"result,omitempty"`
		Quota  int    `json:"quota,omitempty"`
	} `json:"user,omitempty"`
	Status int `json:"status,omitempty"`
}
