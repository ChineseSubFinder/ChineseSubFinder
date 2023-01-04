package a4k

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/series"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/mix_media_info"
	"github.com/Tnze/go.num/v2/zh"
	"github.com/go-resty/resty/v2"
	"github.com/jinzhu/now"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"

	"github.com/PuerkitoBio/goquery"

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
	httpClient, err := pkg.NewHttpClient()
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive.NewHttpClient", err)
		return false, 0
	}
	searPageUrl := fmt.Sprintf(settings.Get().AdvancedSettings.SuppliersSettings.A4k.RootUrl)
	resp, err := httpClient.R().Get(searPageUrl)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive.Get", err)
		return false, 0
	}
	if resp.StatusCode() != 200 {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive.StatusCode", resp.StatusCode())
		return false, 0
	}
	s.isAlive = true
	return true, time.Since(startT).Milliseconds()
}

func (s *Supplier) IsAlive() bool {
	return s.isAlive
}

func (s *Supplier) OverDailyDownloadLimit() bool {

	if settings.Get().AdvancedSettings.SuppliersSettings.A4k.DailyDownloadLimit == 0 {
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
	return common2.SubSiteA4K
}

func (s *Supplier) GetSubListFromFile4Movie(videoFPath string) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), videoFPath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), videoFPath, "Start...")

	outSubInfos := make([]supplier.SubInfo, 0)

	mediaInfo, err := mix_media_info.GetMixMediaInfo(s.fileDownloader.MediaInfoDealers,
		videoFPath, true)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "GetMixMediaInfo", err)
		return nil, err
	}
	// 需要找到中文名称去搜索
	keyWord, err := mix_media_info.KeyWordSelect(mediaInfo, videoFPath, true, "cn")
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "keyWordSelect", err)
		return nil, err
	}
	airTime, err := now.Parse(mediaInfo.Year)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "Parse airTime", err)
		return nil, err
	}
	searchKeyword := fmt.Sprintf("%s %d", keyWord, airTime.Year())
	s.log.Infoln(s.GetSupplierName(), "searchKeyword", searchKeyword)
	searchResultItems, err := s.searchKeyword(searchKeyword, true)
	if err != nil {
		return nil, err
	}
	if len(searchResultItems) == 0 {
		// 没有找到则返回
		s.log.Infoln(s.GetSupplierName(), "searchKeyword", searchKeyword, "not found")
		return nil, nil
	}
	// 开启下载
	downloadCounter := 0
	for _, searchResultItem := range searchResultItems {

		downloadPageUrl := settings.Get().AdvancedSettings.SuppliersSettings.A4k.RootUrl + searchResultItem.RUrl
		subInfo, err := s.downloadSub(videoFPath, downloadPageUrl, 0, 0)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "downloadSub", err)
			return nil, err
		}

		outSubInfos = append(outSubInfos, *subInfo)
		downloadCounter++

		if downloadCounter >= settings.Get().AdvancedSettings.Topic {
			break
		}
	}

	return outSubInfos, nil
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), seriesInfo.Name, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), seriesInfo.Name, "Start...")

	// 搜索的策略应该是 绝命律师 第五季  or 绝命律师 S05E06 这两种方式，优先后者，具体去搜索，如果找不到然后再切换关键词为全季
	outSubInfos := make([]supplier.SubInfo, 0)
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		mediaInfo, err := mix_media_info.GetMixMediaInfo(s.fileDownloader.MediaInfoDealers,
			episodeInfo.FileFullPath, false)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "GetMixMediaInfo", err)
			return nil, err
		}
		// 需要找到中文名称去搜索
		keyWord, err := mix_media_info.KeyWordSelect(mediaInfo, episodeInfo.FileFullPath, false, "cn")
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "keyWordSelect", err)
			return nil, err
		}
		// 第一次搜索 黄石 S04E01
		s.log.Infoln(s.GetSupplierName(), "searchKeyword", keyWord)
		searchResultItems, err := s.searchKeyword(keyWord, false)
		if err != nil {
			return nil, err
		}
		if len(searchResultItems) == 0 {
			// 没有找到则更换关键词
			// 黄石 第四季
			keyWord, err = mix_media_info.KeyWordSelect(mediaInfo, episodeInfo.FileFullPath, true, "cn")
			if err != nil {
				s.log.Errorln(s.GetSupplierName(), "keyWordSelect", err)
				return nil, err
			}
			searchKeyword := fmt.Sprintf("%s %s", keyWord, " 第"+zh.Uint64(episodeInfo.Season).String()+"季")
			s.log.Infoln(s.GetSupplierName(), "searchKeyword", searchKeyword)
			searchResultItems, err = s.searchKeyword(searchKeyword, false)
			if err != nil {
				return nil, err
			}
			if len(searchResultItems) == 0 {
				// 没有找到则返回
				s.log.Infoln(s.GetSupplierName(), episodeInfo.Season, episodeInfo.Episode, "no sub found")
				return nil, nil
			}
		}
		// 开启下载
		downloadCounter := 0
		for _, searchResultItem := range searchResultItems {

			if episodeInfo.Season == searchResultItem.Season && episodeInfo.Episode == searchResultItem.Episode {
				// Season 和 Eps 匹配上再继续下载
			} else if episodeInfo.Season == searchResultItem.Season && searchResultItem.IsFullSeason == true {
				// Season 匹配上，Eps 为 0 则下载，全季
			}
			downloadPageUrl := settings.Get().AdvancedSettings.SuppliersSettings.A4k.RootUrl + searchResultItem.RUrl
			// 注意这里传入的 Season Episode 是这个字幕下载时候解析出来的信息
			subInfo, err := s.downloadSub(episodeInfo.FileFullPath, downloadPageUrl, searchResultItem.Season, searchResultItem.Episode)
			if err != nil {
				s.log.Errorln(s.GetSupplierName(), "downloadSub", err)
				return nil, err
			}

			outSubInfos = append(outSubInfos, *subInfo)
			// 连续剧的时候至多下载 5 个即可
			downloadCounter++
			if downloadCounter >= 5 {
				break
			}
		}
	}

	return outSubInfos, nil
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	panic("not implemented")
}

// searchKeyword 通过关键词获取所有的字幕列表
func (s *Supplier) searchKeyword(keyword string, isMovie bool) (searchResultItems []SearchResultItem, err error) {

	var totalPage int
	totalPage = 0
	searchResultItems = make([]SearchResultItem, 0)
	// 先获取第一页
	nowPageIndex := 0
	for {
		var nowSearchResultItem []SearchResultItem
		if nowPageIndex == 0 {
			// 第一页才有获取总页数的操作
			nowSearchResultItem, totalPage, err = s.listPageItems(keyword, nowPageIndex, isMovie)
		} else {
			// 其他页面，跳过总页数的获取逻辑
			nowSearchResultItem, _, err = s.listPageItems(keyword, nowPageIndex, isMovie)
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

func (s *Supplier) listPageItems(keyword string, pageIndex int, isMovie bool) (searchResultItems []SearchResultItem, totalPage int, err error) {

	defer func() {
		time.Sleep(time.Second * 10)
	}()
	searchResultItems = make([]SearchResultItem, 0)
	httpClient, err := pkg.NewHttpClient()
	if err != nil {
		err = errors.New("NewHttpClient error:" + err.Error())
		return
	}
	// 先对第一页进行分析
	searPageUrl := fmt.Sprintf(settings.Get().AdvancedSettings.SuppliersSettings.A4k.RootUrl+"/search?term=%s&page=%d", url.QueryEscape(keyword), pageIndex)
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
			IsMovie:      isMovie,
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
	if pageIndex == 0 && lastPageSelection == nil {
		// 说明只有一页的结果
		return
	}
	if lastPageSelection != nil {
		// 说明至少有两页
		lastPageHrefUrl, ok := lastPageSelection.Attr("href")
		if pageIndex == 0 && ok == false {
			// 说明只有一页的结果
			return
		}
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

func (s *Supplier) downloadSub(videoFileName, downloadPageUrl string, season, eps int) (subInfo *supplier.SubInfo, err error) {

	defer func() {
		time.Sleep(time.Second * 5)
	}()

	var httpClient *resty.Client
	httpClient, err = pkg.NewHttpClient()
	if err != nil {
		err = errors.New("NewHttpClient error:" + err.Error())
		return
	}
	// 先对第一页进行分析
	var resp *resty.Response
	resp, err = httpClient.R().Get(downloadPageUrl)
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
	// 找到下载的 btn
	downloadBtSelection := doc.Find("div.buttons a.green")
	if downloadBtSelection == nil {
		err = errors.New("download btn is nil")
		return
	}
	downloadBtHrefUrl, ok := downloadBtSelection.Attr("href")
	if ok == false {
		err = errors.New("download btn href is nil")
		return
	}
	// 开始下载
	downloadFileUrl := settings.Get().AdvancedSettings.SuppliersSettings.A4k.RootUrl + downloadBtHrefUrl
	subInfo, err = s.fileDownloader.GetA4k(s.GetSupplierName(), 0, season, eps, videoFileName, downloadFileUrl)
	if err != nil {
		err = errors.New("fileDownloader.Get error:" + err.Error())
		return
	}

	return
}

type SearchResultItem struct {
	Title        string `json:"title"`
	RUrl         string `json:"r_url"`
	IsMovie      bool   `json:"is_movie"`
	Season       int    `json:"season"`
	Episode      int    `json:"episode"`
	IsFullSeason bool   `json:"is_full_season"`
}

const pageTageName = "&page="
