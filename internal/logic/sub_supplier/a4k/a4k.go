package a4k

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"

	"github.com/PuerkitoBio/goquery"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
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
	return common2.SubSiteA4K
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	//return s.findAndDownload(filePath, true, 0, 0)
	return outSubInfos, nil
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	// 搜索的策略应该是 绝命律师 第五季  or 绝命律师 S05E06 这两种方式，优先后者，具体去搜索，如果找不到然后再切换关键词为全季
	outSubInfos := make([]supplier.SubInfo, 0)
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	//for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {
	//
	//	oneSubInfoList, err := s.findAndDownload(episodeInfo.FileFullPath, false, episodeInfo.Season, episodeInfo.Episode)
	//	if err != nil {
	//		return outSubInfos, errors.New("FindAndDownload error:" + err.Error())
	//	}
	//	outSubInfos = append(outSubInfos, oneSubInfoList...)
	//}

	return outSubInfos, nil
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	outSubInfos := make([]supplier.SubInfo, 0)
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	//for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {
	//
	//	oneSubInfoList, err := s.findAndDownload(episodeInfo.FileFullPath, false, episodeInfo.Season, episodeInfo.Episode)
	//	if err != nil {
	//		return outSubInfos, errors.New("FindAndDownload error:" + err.Error())
	//	}
	//	outSubInfos = append(outSubInfos, oneSubInfoList...)
	//}

	return outSubInfos, nil
}

// searchKeyword 通过关键词获取所有的字幕列表
func (s *Supplier) searchKeyword(keyword string) (searchResultItems []SearchResultItem, err error) {

	var totalPage int
	totalPage = 0
	searchResultItems = make([]SearchResultItem, 0)
	// 先获取第一页
	nowPageIndex := 0
	for {
		var nowSearchResultItem []SearchResultItem
		if nowPageIndex == 0 {
			// 第一页才有获取总页数的操作
			nowSearchResultItem, totalPage, err = s.listPageItems(keyword, nowPageIndex)
		} else {
			// 其他页面，跳过总页数的获取逻辑
			nowSearchResultItem, _, err = s.listPageItems(keyword, nowPageIndex)
		}
		if err != nil {
			return
		}
		searchResultItems = append(searchResultItems, nowSearchResultItem...)
		if totalPage == 0 {
			// 说明只有一页
			break
		}
		nowPageIndex++
		if nowPageIndex > totalPage {
			// 超过总页数
			break
		}
	}
	return
}

func (s *Supplier) listPageItems(keyword string, pageIndex int) (searchResultItems []SearchResultItem, totalPage int, err error) {

	searchResultItems = make([]SearchResultItem, 0)
	httpClient, err := my_util.NewHttpClient(s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		err = errors.New("NewHttpClient error:" + err.Error())
		return
	}
	// 先对第一页进行分析
	searPageUrl := fmt.Sprintf(s.settings.AdvancedSettings.SuppliersSettings.A4k.RootUrl+"/search?term=%s&page=%d", keyword, pageIndex)
	resp, err := httpClient.R().Get(searPageUrl)
	if err != nil {
		err = errors.New("http get error:" + err.Error())
		return
	}
	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		err = errors.New("goquery NewDocumentFromReader error:" + err.Error())
		return
	}
	doc.Find(".sub-item-list li.item div.content h3").EachWithBreak(func(i int, selection *goquery.Selection) bool {

		subA := selection.Find("a")
		if subA == nil {
			s.log.Errorln(".sub-item-list li.item div.content h3 a is nil")
			return false
		}
		title := subA.Text()
		hrefUrl, ok := subA.Attr("href")
		if ok == false {
			s.log.Errorln("a href is nil")
			return false
		}
		var isFullSeason bool
		var season int
		var eps int
		isFullSeason, season, eps, err = decode.GetSeasonAndEpisodeFromSubFileName(title)
		if err != nil {
			s.log.Errorln("decode.GetSeasonAndEpisodeFromSubFileName error:" + err.Error())
			return false
		}
		searchResultItems = append(searchResultItems, SearchResultItem{
			Title:        title,
			RUrl:         hrefUrl,
			Season:       season,
			Episode:      eps,
			IsFullSeason: isFullSeason,
		})

		return true
	})
	if len(searchResultItems) < 1 {
		// 说明没有搜索的结果
		return
	}
	err = nil
	totalPage = 0
	if pageIndex != 0 {
		// 没有必要分析最后一页的信息
		return
	}
	// 判断一共有多少页，一页就是 0，第二页就是 1，以此类推
	lastPageSelection := doc.Find("a.pager__item--last")
	if lastPageSelection != nil {
		// 说明至少有两页
		lastPageHrefUrl, ok := lastPageSelection.Attr("href")
		if ok == false {
			err = errors.New("last page a href is nil")
			return
		}
		if strings.Contains(lastPageHrefUrl, pageTageName) == false {
			err = errors.New("last page a href is not correct, not found page tag")
			return
		}
		lastPageParts := strings.Split(lastPageHrefUrl, pageTageName)
		if len(lastPageParts) != 2 {
			err = errors.New("last page a href is not correct, split parts error")
			return
		}
		totalPage, err = strconv.Atoi(lastPageParts[1])
		if err != nil {
			err = errors.New("last page a href is not correct, convert to int error")
			return
		}
	}

	return
}

func (s Supplier) downloadSub(downloadPageUrl string) {

	httpClient, err := my_util.NewHttpClient(s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		return
	}
	// 先对第一页进行分析
	resp, err := httpClient.R().Get(downloadPageUrl)
	if err != nil {
		return
	}
	if err != nil {
		err = errors.New("goquery NewDocumentFromReader error:" + err.Error())
		return
	}
	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		err = errors.New("goquery NewDocumentFromReader error:" + err.Error())
		return
	}

	doc.Find(".sub-item-list li.item div.content h3")
}

type SearchResultItem struct {
	Title        string `json:"title"`
	RUrl         string `json:"r_url"`
	Season       int    `json:"season"`
	Episode      int    `json:"episode"`
	IsFullSeason bool   `json:"is_full_season"`
}

const pageTageName = "&page="
