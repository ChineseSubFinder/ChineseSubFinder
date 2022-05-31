package subhd

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/mix_media_info"

	"github.com/PuerkitoBio/goquery"
	"github.com/Tnze/go.num/v2/zh"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/url_connectedness_helper"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
)

type Supplier struct {
	settings       *settings.Settings
	log            *logrus.Logger
	fileDownloader *file_downloader.FileDownloader
	topic          int
	tt             time.Duration
	debugMode      bool
	isAlive        bool
}

func NewSupplier(fileDownloader *file_downloader.FileDownloader) *Supplier {

	sup := Supplier{}
	sup.log = fileDownloader.Log
	sup.fileDownloader = fileDownloader
	sup.topic = common2.DownloadSubsPerSite

	sup.settings = fileDownloader.Settings
	if sup.settings.AdvancedSettings.Topic > 0 && sup.settings.AdvancedSettings.Topic != sup.topic {
		sup.topic = sup.settings.AdvancedSettings.Topic
	}
	sup.isAlive = true // 默认是可以使用的，如果 check 后，再调整状态

	// 默认超时是 2 * 60s，如果是调试模式则是 5 min
	sup.tt = common2.HTMLTimeOut
	sup.debugMode = sup.settings.AdvancedSettings.DebugMode
	if sup.debugMode == true {
		sup.tt = common2.OneMovieProcessTimeOut
	}

	return &sup
}

func (s *Supplier) CheckAlive() (bool, int64) {

	proxyStatus, proxySpeed, err := url_connectedness_helper.UrlConnectednessTest(s.settings.AdvancedSettings.SuppliersSettings.SubHD.RootUrl,
		s.settings.AdvancedSettings.ProxySettings.GetLocalHttpProxyUrl())
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Error", err)
		s.isAlive = false
		return false, 0
	}
	if proxyStatus == false {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Status != 200")
		s.isAlive = false
		return false, proxySpeed
	}

	s.isAlive = true
	return true, proxySpeed
}

func (s *Supplier) IsAlive() bool {
	return s.isAlive
}

func (s *Supplier) OverDailyDownloadLimit() bool {

	// 需要查询今天的限额
	count, err := s.fileDownloader.CacheCenter.DailyDownloadCountGet(s.GetSupplierName(),
		my_util.GetPublicIP(s.log, s.settings.AdvancedSettings.TaskQueue, s.settings.AdvancedSettings.ProxySettings))
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "DailyDownloadCountGet", err)
		return true
	}
	if count > s.settings.AdvancedSettings.SuppliersSettings.Zimuku.DailyDownloadLimit {
		// 超限了
		s.log.Warningln(s.GetSupplierName(), "DailyDownloadLimit:", s.settings.AdvancedSettings.SuppliersSettings.SubHD.DailyDownloadLimit, "Now Is:", count)
		return true
	} else {
		// 没有超限
		s.log.Infoln(s.GetSupplierName(), "DailyDownloadLimit:", s.settings.AdvancedSettings.SuppliersSettings.Zimuku.DailyDownloadLimit, "Now Is:", count)
		return false
	}
}

func (s *Supplier) GetLogger() *logrus.Logger {
	return s.log
}

func (s *Supplier) GetSettings() *settings.Settings {
	return s.settings
}

func (s *Supplier) GetSupplierName() string {
	return common2.SubSiteSubHd
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {
	return s.getSubListFromFile4Movie(filePath)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	var browser *rod.Browser
	// TODO 是用本地的 Browser 还是远程的，推荐是远程的
	browser, err := rod_helper.NewBrowserEx(s.log, true, s.settings)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = browser.Close()
	}()

	mediaInfo, err := mix_media_info.GetMixMediaInfo(s.log, s.fileDownloader.SubtitleBestApi,
		seriesInfo.EpList[0].FileFullPath, false,
		s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), seriesInfo.EpList[0].FileFullPath, "GetMixMediaInfo", err)
		return nil, err
	}
	// 优先中文查询
	keyWord, err := mix_media_info.KeyWordSelect(mediaInfo, seriesInfo.EpList[0].FileFullPath, true, "cn")
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), seriesInfo.EpList[0].FileFullPath, "keyWordSelect", err)
		return nil, err
	}
	if keyWord == "" {
		// 更换英文译名
		keyWord, err = mix_media_info.KeyWordSelect(mediaInfo, seriesInfo.EpList[0].FileFullPath, true, "en")
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), seriesInfo.EpList[0].FileFullPath, "keyWordSelect", err)
			return nil, err
		}
	}
	var subInfos = make([]supplier.SubInfo, 0)
	var subList = make([]HdListItem, 0)
	for value := range seriesInfo.NeedDlSeasonDict {
		// 第一级界面，找到影片的详情界面
		//keyword := seriesInfo.Name + " 第" + zh.Uint64(value).String() + "季"
		keyword := keyWord + " 第" + zh.Uint64(value).String() + "季"
		s.log.Infoln("Search Keyword:", keyword)
		detailPageUrl, err := s.step0(browser, keyword)
		if err != nil {
			s.log.Errorln("subhd step0", keyword)
			return nil, err
		}
		if detailPageUrl == "" {
			// 如果只是搜索不到，则继续换关键词
			s.log.Warning("subhd first search keyword", keyword, "not found")
			keyword = seriesInfo.Name
			s.log.Warning("subhd Retry", keyword)
			s.log.Infoln("Search Keyword:", keyword)
			detailPageUrl, err = s.step0(browser, keyword)
			if err != nil {
				s.log.Errorln("subhd step0", keyword)
				return nil, err
			}
		}
		if detailPageUrl == "" {
			s.log.Warning("subhd search keyword", keyword, "not found")
			continue
		}
		// 列举字幕
		oneSubList, err := s.step1(browser, detailPageUrl, false)
		if err != nil {
			s.log.Errorln("subhd step1", keyword)
			return nil, err
		}

		subList = append(subList, oneSubList...)
	}
	// 与剧集需要下载的集 List 进行比较，找到需要下载的列表
	// 找到那些 Eps 需要下载字幕的
	subInfoNeedDownload := s.whichEpisodeNeedDownloadSub(seriesInfo, subList)
	// 下载字幕
	for i, item := range subInfoNeedDownload {

		subInfo, err := s.fileDownloader.GetEx(s.GetSupplierName(), browser, item.Url, int64(i), item.Season, item.Episode, s.DownFile)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "GetEx", item.Title, item.Season, item.Episode, err)
			continue
		}

		subInfos = append(subInfos, *subInfo)
	}

	return subInfos, nil
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	panic("not implemented")
}

func (s *Supplier) getSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {
	/*
		虽然是传入视频文件路径，但是其实需要读取对应的视频文件目录下的
		movie.xml 以及 *.nfo，找到 IMDB id
		优先通过 IMDB id 去查找字幕
		如果找不到，再靠文件名提取影片名称去查找
	*/
	// 找到这个视频文件，尝试得到 IMDB ID
	// 目前测试来看，加入 年 这个关键词去搜索，对 2020 年后的影片有利，因为网站有统一的详细页面了，而之前的，没有，会影响识别
	// 所以，year >= 2020 年，则可以多加一个关键词（年）去搜索影片
	imdbInfo, err := decode.GetImdbInfo4Movie(filePath)
	if err != nil {
		// 允许的错误，跳过，继续进行文件名的搜索
		s.log.Errorln("model.GetImdbInfo", err)
	}
	var subInfoList []supplier.SubInfo

	if imdbInfo.ImdbId != "" {
		// 先用 imdb id 找
		subInfoList, err = s.getSubListFromKeyword4Movie(imdbInfo.ImdbId)
		if err != nil {
			// 允许的错误，跳过，继续进行文件名的搜索
			s.log.Errorln(s.GetSupplierName(), "keyword:", imdbInfo.ImdbId)
			s.log.Errorln("getSubListFromKeyword4Movie", "IMDBID can not found sub", filePath, err)
		}
		// 如果有就优先返回
		if len(subInfoList) > 0 {
			return subInfoList, nil
		}
	}
	s.log.Infoln(s.GetSupplierName(), filePath, "No subtitle found", "KeyWord:", imdbInfo.ImdbId)
	mediaInfo, err := mix_media_info.GetMixMediaInfo(s.log, s.fileDownloader.SubtitleBestApi,
		filePath, true,
		s.settings.AdvancedSettings.ProxySettings)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), filePath, "GetMixMediaInfo", err)
		return nil, err
	}
	// 优先中文查询
	keyWord, err := mix_media_info.KeyWordSelect(mediaInfo, filePath, true, "cn")
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), filePath, "keyWordSelect", err)
		return nil, err
	}
	// 如果没有，那么就用文件名查找
	searchKeyword := my_util.VideoNameSearchKeywordMaker(s.log, keyWord, imdbInfo.Year)
	subInfoList, err = s.getSubListFromKeyword4Movie(searchKeyword)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "keyword:", searchKeyword)
		return nil, err
	}
	if len(subInfoList) < 1 {
		// 切换到英文查询
		s.log.Infoln(s.GetSupplierName(), filePath, "No subtitle found", "KeyWord:", searchKeyword)
		keyWord, err = mix_media_info.KeyWordSelect(mediaInfo, filePath, true, "cn")
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), filePath, "keyWordSelect", err)
			return nil, err
		}
		// 如果没有，那么就用文件名查找
		searchKeyword = my_util.VideoNameSearchKeywordMaker(s.log, keyWord, imdbInfo.Year)
		subInfoList, err = s.getSubListFromKeyword4Movie(searchKeyword)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "keyword:", searchKeyword)
			return nil, err
		}
		if len(subInfoList) < 1 {
			s.log.Infoln(s.GetSupplierName(), filePath, "No subtitle found", "KeyWord:", searchKeyword)
		}
	}

	return subInfoList, nil
}

func (s *Supplier) getSubListFromKeyword4Movie(keyword string) ([]supplier.SubInfo, error) {

	s.log.Infoln("Search Keyword:", keyword)
	var browser *rod.Browser
	// TODO 是用本地的 Browser 还是远程的，推荐是远程的
	browser, err := rod_helper.NewBrowserEx(s.log, true, s.settings)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = browser.Close()
	}()
	var subInfos []supplier.SubInfo
	detailPageUrl, err := s.step0(browser, keyword)
	if err != nil {
		return nil, err
	}
	// 没有搜索到字幕
	if detailPageUrl == "" {
		return nil, nil
	}
	subList, err := s.step1(browser, detailPageUrl, true)
	if err != nil {
		return nil, err
	}

	for i, item := range subList {

		subInfo, err := s.fileDownloader.GetEx(s.GetSupplierName(), browser, item.Url, int64(i), 0, 0, s.DownFile)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "GetEx", item.Title, item.Season, item.Episode, err)
			continue
		}

		subInfos = append(subInfos, *subInfo)
	}

	return subInfos, nil
}

func (s *Supplier) whichEpisodeNeedDownloadSub(seriesInfo *series.SeriesInfo, allSubList []HdListItem) []HdListItem {
	// 字幕很多，考虑效率，需要做成字典
	// key SxEx - SubInfos
	var allSubDict = make(map[string][]HdListItem)
	// 全季的字幕列表
	var oneSeasonSubDict = make(map[string][]HdListItem)
	for _, subInfo := range allSubList {
		_, season, episode, err := decode.GetSeasonAndEpisodeFromSubFileName(subInfo.Title)
		if err != nil {
			s.log.Errorln("whichEpisodeNeedDownloadSub.GetVideoInfoFromFileFullPath", subInfo.Title, err)
			continue
		}
		subInfo.Season = season
		subInfo.Episode = episode
		epsKey := my_util.GetEpisodeKeyName(season, episode)
		_, ok := allSubDict[epsKey]
		if ok == false {
			// 初始化
			allSubDict[epsKey] = make([]HdListItem, 0)
			if season != 0 && episode == 0 {
				oneSeasonSubDict[epsKey] = make([]HdListItem, 0)
			}
		}
		// 添加
		allSubDict[epsKey] = append(allSubDict[epsKey], subInfo)
		if season != 0 && episode == 0 {
			oneSeasonSubDict[epsKey] = append(oneSeasonSubDict[epsKey], subInfo)
		}
	}
	// 本地的视频列表，找到没有字幕的
	// 需要进行下载字幕的列表
	var subInfoNeedDownload = make([]HdListItem, 0)
	// 有那些 Eps 需要下载的，按 SxEx 反回 epsKey
	for epsKey, epsInfo := range seriesInfo.NeedDlEpsKeyList {
		// 从一堆字幕里面找合适的
		value, ok := allSubDict[epsKey]
		// 是否有
		if ok == true && len(value) > 0 {
			value[0].Season = epsInfo.Season
			value[0].Episode = epsInfo.Episode
			subInfoNeedDownload = append(subInfoNeedDownload, value[0])
		}
	}
	// 全季的字幕列表，也拼进去，后面进行下载
	for _, infos := range oneSeasonSubDict {

		if len(infos) < 1 {
			continue
		}
		subInfoNeedDownload = append(subInfoNeedDownload, infos[0])
	}

	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return subInfoNeedDownload
}

// step0 找到这个影片的详情列表
func (s *Supplier) step0(browser *rod.Browser, keyword string) (string, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("subhd_step0", err.Error())
		}
	}()

	result, page, err := rod_helper.HttpGetFromBrowser(browser, fmt.Sprintf(s.settings.AdvancedSettings.SuppliersSettings.SubHD.RootUrl+common2.SubSubHDSearchUrl, url.QueryEscape(keyword)), s.tt)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = page.Close()
	}()
	// 是否有查找到的结果，至少要有结果。根据这里这样下面才能判断是分析失效了，还是就是没有结果而已
	re := regexp.MustCompile(`共\s*(\d+)\s*条`)
	matched := re.FindAllStringSubmatch(result, -1)
	if matched == nil || len(matched) < 1 {
		return "", common2.SubHDStep0SubCountElementNotFound
	}
	subCount, err := decode.GetNumber2int(matched[0][0])
	if err != nil {
		return "", err
	}
	// 如果所搜没有找到字幕，就要返回
	if subCount < 1 {
		return "", nil
	}
	// 这里是确认能继续分析的详细连接
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		return "", err
	}
	imgSelection := doc.Find("img.rounded-start")
	_, ok := imgSelection.Attr("src")
	if ok == true {

		if len(imgSelection.Nodes) < 1 {
			return "", common2.SubHDStep0ImgParentLessThan1
		}
		step1Url := ""
		if imgSelection.Nodes[0].Parent.Data == "a" {
			// 第一个父级是不是超链接
			for _, attribute := range imgSelection.Nodes[0].Parent.Attr {
				if attribute.Key == "href" {
					step1Url = attribute.Val
					break
				}
			}
		} else if imgSelection.Nodes[0].Parent.Parent.Data == "a" {
			// 第二个父级是不是超链接
			for _, attribute := range imgSelection.Nodes[0].Parent.Parent.Attr {
				if attribute.Key == "href" {
					step1Url = attribute.Val
					break
				}
			}
		}
		if step1Url == "" {
			return "", common2.SubHDStep0HrefIsNull
		}
		return step1Url, nil
	} else {
		return "", common2.SubHDStep0HrefIsNull
	}
}

// step1 获取影片的详情字幕列表
func (s *Supplier) step1(browser *rod.Browser, detailPageUrl string, isMovieOrSeries bool) ([]HdListItem, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("subhd_step1", err.Error())
		}
	}()
	detailPageUrl = my_util.AddBaseUrl(s.settings.AdvancedSettings.SuppliersSettings.SubHD.RootUrl, detailPageUrl)
	result, page, err := rod_helper.HttpGetFromBrowser(browser, detailPageUrl, s.tt)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = page.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		return nil, err
	}
	var lists []HdListItem

	const subTableKeyword = ".pt-2"
	const oneSubTrTitleKeyword = "a.link-dark"
	const oneSubTrDownloadCountKeyword = "div.px-3"
	const oneSubLangAndTypeKeyword = ".text-secondary"

	doc.Find(subTableKeyword).EachWithBreak(func(i int, tr *goquery.Selection) bool {
		if tr.Find(oneSubTrTitleKeyword).Size() == 0 {
			return true
		}
		// 文件的下载页面，还需要分析
		downUrl, exists := tr.Find(oneSubTrTitleKeyword).Eq(0).Attr("href")
		if !exists {
			return true
		}
		// 文件名
		title := strings.TrimSpace(tr.Find(oneSubTrTitleKeyword).Text())
		// 字幕类型
		insideSubType := tr.Find(oneSubLangAndTypeKeyword).Text()
		if sub_parser_hub.IsSubTypeWanted(insideSubType) == false {
			return true
		}
		// 下载的次数
		downCount, err := decode.GetNumber2int(tr.Find(oneSubTrDownloadCountKeyword).Eq(1).Text())
		if err != nil {
			return true
		}

		listItem := HdListItem{}
		listItem.Url = downUrl
		listItem.BaseUrl = s.settings.AdvancedSettings.SuppliersSettings.SubHD.RootUrl
		listItem.Title = title
		listItem.DownCount = downCount

		// 电影，就需要第一个
		// 连续剧，需要多个
		if isMovieOrSeries == true {

			if len(lists) >= s.topic {
				return false
			}
		}
		lists = append(lists, listItem)
		return true
	})

	return lists, nil
}

// DownFile 下载字幕 过防水墙
func (s *Supplier) DownFile(browser *rod.Browser, subDownloadPageUrl string, TopN int64, Season, Episode int) (*supplier.SubInfo, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("subhd_DownFile", err.Error())
		}
	}()
	subDownloadPageFullUrl := my_util.AddBaseUrl(s.settings.AdvancedSettings.SuppliersSettings.SubHD.RootUrl, subDownloadPageUrl)

	_, page, err := rod_helper.HttpGetFromBrowser(browser, subDownloadPageFullUrl, s.tt)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = page.Close()
	}()

	// 需要先判断是否先要输入验证码，然后才到下载界面
	// 下载字幕
	subInfo, err := s.downloadSubFile(browser, page, subDownloadPageUrl)
	if err != nil {
		return nil, err
	}

	subInfo.TopN = TopN
	subInfo.Season = Season
	subInfo.Episode = Episode

	return subInfo, nil
}

func (s *Supplier) downloadSubFile(browser *rod.Browser, page *rod.Page, subDownloadPageUrl string) (*supplier.SubInfo, error) {

	var err error
	var doc *goquery.Document
	downloadSuccess := false
	fileName := ""
	fileByte := []byte{0}
	err = rod.Try(func() {
		tmpDir := filepath.Join(global_value.DefTmpFolder(), "downloads")
		wait := browser.Timeout(30 * time.Second).WaitDownload(tmpDir)
		getDownloadFile := func() ([]byte, string, error) {
			info := wait()

			if info == nil {
				return nil, "", errors.New("download sub timeout")
			}

			downloadPath := filepath.Join(tmpDir, info.GUID)
			defer func() { _ = os.Remove(downloadPath) }()
			b, err := os.ReadFile(downloadPath)
			if err != nil {
				return nil, "", err
			}
			return b, info.SuggestedFilename, nil
		}
		// 初始化页面用于查询元素
		pString := page.MustHTML()
		doc, err = goquery.NewDocumentFromReader(strings.NewReader(pString))
		if err != nil {
			return
		}
		// 点击“验证获取下载地址”
		s.log.Debugln("click '验证获取下载地址'")
		clickCodeBtn := doc.Find(btnClickCodeBtn)
		if len(clickCodeBtn.Nodes) < 1 {
			return
		}
		element := page.MustElement(btnClickCodeBtn)
		BtnCodeText := element.MustText()
		if strings.Contains(BtnCodeText, "验证") == true {
			s.log.Debugln("find '验证' 关键词")
			// 那么需要填写验证码
			element.MustClick()
			time.Sleep(time.Second * 2)
			// 填写“验证码”
			s.log.Debugln("填写验证码")
			el := page.MustElement("#gzhcode")
			el.MustInput(common2.SubhdCode)
			//page.MustEval(`$("#gzhcode").attr("value","` + common2.SubhdCode + `");`)
			// 是否有“完成验证”按钮
			s.log.Debugln("查找是否有交验证码按钮1")
			downBtn := doc.Find(btnCommitCode)
			if len(downBtn.Nodes) < 1 {
				return
			}
			s.log.Debugln("查找是否有交验证码按钮2")
			element = page.MustElement(btnCommitCode)
			benCommit := element.MustText()
			if strings.Contains(benCommit, "验证") == false {
				s.log.Errorln("btn not found 完整验证")
				return
			}
			s.log.Debugln("点击提交验证码")
			element.MustClick()
			time.Sleep(time.Second * 2)

			s.log.Debugln("点击下载按钮")
			// 点击下载按钮
			page.MustElement(btnClickCodeBtn).MustClick()

			time.Sleep(time.Second * 2)
		} else if strings.Contains(BtnCodeText, "下载") == true {

			s.log.Debugln("点击下载按钮")
			// 直接可以下载
			element.MustClick()
			time.Sleep(time.Second * 2)
		} else {

			s.log.Errorln("btn not found 下载验证 or 下载")
			return
		}
		// 更新 page 的实例对应的 doc Content
		pString = page.MustHTML()
		doc, err = goquery.NewDocumentFromReader(strings.NewReader(pString))
		if err != nil {
			return
		}
		// 是否有腾讯的防水墙
		hasWaterWall := false
		waterWall := doc.Find(TCode)
		if len(waterWall.Nodes) >= 1 {
			hasWaterWall = true
		}
		s.log.Debugln("Need pass WaterWall", hasWaterWall)
		// 过墙
		if hasWaterWall == true {
			s.passWaterWall(page)
		}
		fileByte, fileName, err = getDownloadFile()
		if err != nil {
			return
		}
		downloadSuccess = true
	})
	if err != nil {
		return nil, err
	}

	inSubInfo := supplier.NewSubInfo(s.GetSupplierName(), 1, fileName, language.ChineseSimple, subDownloadPageUrl, 0, 0, filepath.Ext(fileName), fileByte)
	if downloadSuccess == false {
		return nil, common2.SubHDStep2ExCannotFindDownloadBtn
	}

	return inSubInfo, nil
}

func (s *Supplier) passWaterWall(page *rod.Page) {

	const (
		waterIFrame = "#tcaptcha_iframe"
		dragBtn     = "#tcaptcha_drag_button"
		slideBg     = "#slideBg"
	)

	//等待驗證碼窗體載入
	page.MustElement(waterIFrame).MustWaitLoad()
	//進入到iframe
	iframe := page.MustElement(waterIFrame).MustFrame()
	// see iframe bug, see  https://github.com/go-rod/rod/issues/548
	p := page.Browser().MustPageFromTargetID(proto.TargetTargetID(iframe.FrameID))

	//等待拖動條加載, 延遲500秒檢測變化, 以確認加載完畢
	p.MustElement(dragBtn).MustWaitStable()
	//等待缺口圖像載入
	slideBgEl := p.MustElement(slideBg).MustWaitLoad()
	slideBgEl = slideBgEl.MustWaitStable()
	//取得帶缺口圖像
	shadowbg := slideBgEl.MustResource()
	// 取得原始圖像
	src := slideBgEl.MustProperty("src")
	fullbg, _, err := my_util.DownFile(s.log, strings.Replace(src.String(), "img_index=1", "img_index=0", 1))
	if err != nil {
		s.log.Errorln("passWaterWall.DownFile", err)
		return
	}
	//取得img展示的真實尺寸
	shape, err := slideBgEl.Shape()
	if err != nil {
		s.log.Errorln("passWaterWall.Shape", err)
		return
	}
	bgbox := shape.Box()
	height, width := uint(math.Round(bgbox.Height)), uint(math.Round(bgbox.Width))
	//裁剪圖像
	shadowbgImg, _ := jpeg.Decode(bytes.NewReader(shadowbg))
	shadowbgImg = resize.Resize(width, height, shadowbgImg, resize.Lanczos3)
	fullbgImg, _ := jpeg.Decode(bytes.NewReader(fullbg))
	fullbgImg = resize.Resize(width, height, fullbgImg, resize.Lanczos3)

	//啓始left，排除干擾部份，所以右移10個像素
	left := fullbgImg.Bounds().Min.X + 10
	//啓始top, 排除干擾部份, 所以下移10個像素
	top := fullbgImg.Bounds().Min.Y + 10
	//最大left, 排除干擾部份, 所以左移10個像素
	maxleft := fullbgImg.Bounds().Max.X - 10
	//最大top, 排除干擾部份, 所以上移10個像素
	maxtop := fullbgImg.Bounds().Max.Y - 10
	//rgb比较阈值, 超出此阈值及代表找到缺口位置
	threshold := 20
	//缺口偏移, 拖動按鈕初始會偏移27.5
	distance := -27.5
	//取絕對值方法
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}
search:
	for i := left; i <= maxleft; i++ {
		for j := top; j <= maxtop; j++ {
			colorAR, colorAG, colorAB, _ := fullbgImg.At(i, j).RGBA()
			colorBR, colorBG, colorBB, _ := shadowbgImg.At(i, j).RGBA()
			colorAR, colorAG, colorAB = colorAR>>8, colorAG>>8, colorAB>>8
			colorBR, colorBG, colorBB = colorBR>>8, colorBG>>8, colorBB>>8
			if abs(int(colorAR)-int(colorBR)) > threshold ||
				abs(int(colorAG)-int(colorBG)) > threshold ||
				abs(int(colorAB)-int(colorBB)) > threshold {
				distance += float64(i)
				s.log.Debugln("對比完畢, 偏移量:", distance)
				break search
			}
		}
	}
	//獲取拖動按鈕形狀
	dragBtnBox := p.MustElement("#tcaptcha_drag_thumb").MustShape().Box()
	//启用滑鼠功能
	mouse := p.Mouse
	//模擬滑鼠移動至拖動按鈕處, 右移3的原因: 拖動按鈕比滑塊圖大3個像素
	mouse.MustMove(dragBtnBox.X+3, dragBtnBox.Y+(dragBtnBox.Height/2))
	//按下滑鼠左鍵
	mouse.MustDown("left")
	//開始拖動
	err = mouse.Move(dragBtnBox.X+distance, dragBtnBox.Y+(dragBtnBox.Height/2), 20)
	if err != nil {
		s.log.Errorln("mouse.Move", err)
	}
	//鬆開滑鼠左鍵, 拖动完毕
	mouse.MustUp("left")

	if s.debugMode == true {
		//截圖保存
		page.MustScreenshot(global_value.DefDebugFolder(), "result.png")
	}
}

type HdListItem struct {
	Url        string `json:"url"`
	BaseUrl    string `json:"baseUrl"`
	Title      string `json:"title"`
	Ext        string `json:"ext"`
	AuthorInfo string `json:"authorInfo"`
	Lang       string `json:"lang"`
	Rate       string `json:"rate"`
	DownCount  int    `json:"downCount"`
	Season     int    // 第几季，默认-1
	Episode    int    // 第几集，默认-1
}

//type HdContent struct {
//	Filename string `json:"filename"`
//	Ext      string `json:"ext"`
//	Data     []byte `json:"data"`
//}

const TCode = "#TencentCaptcha"
const btnClickCodeBtn = "button.btn-danger"
const btnCommitCode = "button.btn-primary"
