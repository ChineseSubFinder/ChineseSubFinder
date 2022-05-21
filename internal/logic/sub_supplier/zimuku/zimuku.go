package zimuku

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Tnze/go.num/v2/zh"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	language2 "github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

type Supplier struct {
	settings         *settings.Settings
	log              *logrus.Logger
	fileDownloader   *file_downloader.FileDownloader
	tt               time.Duration
	debugMode        bool
	httpProxyAddress string
	topic            int
	isAlive          bool
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

	// 默认超时是 2 * 60s，如果是调试模式则是 5 min
	sup.tt = common2.HTMLTimeOut
	sup.debugMode = sup.settings.AdvancedSettings.DebugMode
	if sup.debugMode == true {
		sup.tt = common2.OneMovieProcessTimeOut
	}
	// 判断是否启用代理
	sup.httpProxyAddress = sup.settings.AdvancedSettings.ProxySettings.GetLocalHttpProxyUrl()
	return &sup
}

func (s *Supplier) CheckAlive() (bool, int64) {

	// TODO 是用本地的 Browser 还是远程的，推荐是远程的
	browser, err := rod_helper.NewBrowserEx(s.log, true, s.settings, s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl)
	if err != nil {
		return false, 0
	}
	defer func() {
		_ = browser.Close()
	}()

	begin := time.Now() //判断代理访问时间
	_, page, err := rod_helper.HttpGetFromBrowser(browser, s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl, 15*time.Second)
	if err != nil {
		return false, 0
	}
	_ = page.Close()
	speed := time.Now().Sub(begin).Nanoseconds() / 1000 / 1000 //ms
	s.isAlive = true
	return true, speed
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
		s.log.Warningln(s.GetSupplierName(), "DailyDownloadLimit:", s.settings.AdvancedSettings.SuppliersSettings.Zimuku.DailyDownloadLimit, "Now Is:", count)
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
	return common2.SubSiteZiMuKu
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {

	// TODO 是用本地的 Browser 还是远程的，推荐是远程的
	browser, err := rod_helper.NewBrowserEx(s.log, true, s.settings, s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = browser.Close()
	}()

	return s.getSubListFromMovie(browser, filePath)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), seriesInfo.Name, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), seriesInfo.Name, "Start...")

	var err error
	// TODO 是用本地的 Browser 还是远程的，推荐是远程的
	browser, err := rod_helper.NewBrowserEx(s.log, true, s.settings, s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = browser.Close()
	}()
	/*
		去网站搜索的时候，有个比较由意思的逻辑，有些剧集，哪怕只有一季，sonarr 也会给它命名为 Season 1
		但是在 zimuku 搜索的时候，如果你加上 XXX 第一季 就搜索不出来，那么目前比较可行的办法是查询两次
		第一次优先查询 XXX 第一季 ，如果返回的列表是空的，那么再查询 XXX
	*/
	// 这里打算牺牲效率，提高代码的复用度，不然后续得维护一套电影的查询逻辑，一套剧集的查询逻辑
	// 比如，其实可以搜索剧集名称，应该可以得到多个季的列表，然后分析再继续
	// 现在粗暴点，直接一季搜索一次，跟电影的搜索一样，在首个影片就停止，然后继续往下
	AllSeasonSubResult := SubResult{}
	for value := range seriesInfo.NeedDlSeasonDict {

		/*
			经过网友的测试反馈，每一季 zimuku 是支持这一季的第一集的 IMDB ID 可以搜索到这一季的信息 #253
			1. 那么在搜索某一集的时候，需要根据这一集去找这一季的第一集，然后读取它的 IMDB ID 信息，然后优先用于搜索这一季的信息
			2. 如果是搜索季，就直接推算到达季文件夹的位置，搜索所有文件找到第一集，获取它的 IMDB ID
			是不是有点绕···
		*/
		findSeasonFirstEpsIMDBId := ""
		videoList, err := my_util.SearchMatchedVideoFile(s.log, seriesInfo.DirPath)
		if err != nil {
			s.log.Errorln("GetSubListFromFile4Series.SearchMatchedVideoFile, Season:", value, "Error:", err)
			continue
		}
		for _, oneVideoFPath := range videoList {
			oneVideoInfo, err := decode.GetVideoInfoFromFileName(filepath.Base(oneVideoFPath))
			if err != nil {
				s.log.Errorln("GetVideoInfoFromFileName", oneVideoInfo, err)
				continue
			}
			if oneVideoInfo.Season == value && oneVideoInfo.Episode == 1 {
				// 这一季的第一集
				episodeInfo, err := decode.GetImdbInfo4OneSeriesEpisode(oneVideoFPath)
				if err != nil {
					s.log.Errorln("GetImdbInfo4OneSeriesEpisode", oneVideoFPath, err)
					break
				}
				findSeasonFirstEpsIMDBId = episodeInfo.ImdbId
				break
			}
		}

		filmDetailPageUrl := ""
		if findSeasonFirstEpsIMDBId != "" {
			// 第一级界面，找到影片的详情界面
			// 使用上面得到的这一季第一集的 IMDB ID 进行搜索这一季的信息
			keyword := findSeasonFirstEpsIMDBId
			s.log.Debugln(s.GetSupplierName(), "step 0", "1 times", "keyword:", keyword)
			filmDetailPageUrl, err = s.step0(browser, keyword)
			if err != nil {
				s.log.Errorln(s.GetSupplierName(), "step 0", "0 times", "keyword:", keyword, err)
				// 如果只是搜索不到，则继续换关键词
				if err != common2.ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound {
					s.log.Errorln(s.GetSupplierName(), "ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound", keyword, err)
					continue
				}
			}
		}
		// 如果上面找到了，那么 filmDetailPageUrl 就应该不为空，如果没有找到就是空
		if filmDetailPageUrl == "" {
			// 第一级界面，找到影片的详情界面
			keyword := seriesInfo.Name + " 第" + zh.Uint64(value).String() + "季"
			s.log.Debugln(s.GetSupplierName(), "step 0", "0 times", "keyword:", keyword)
			filmDetailPageUrl, err = s.step0(browser, keyword)
			if err != nil {
				s.log.Errorln(s.GetSupplierName(), "step 0", "0 times", "keyword:", keyword, err)
				// 如果只是搜索不到，则继续换关键词
				if err != common2.ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound {
					s.log.Errorln(s.GetSupplierName(), "ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound", keyword, err)
					continue
				}

				// 直接更换为这个剧目的 Name 的搜索，不带季度关键词信息
				keyword = seriesInfo.Name
				s.log.Debugln(s.GetSupplierName(), "step 0", "1 times", "keyword:", keyword)
				filmDetailPageUrl, err = s.step0(browser, keyword)
				if err != nil {
					s.log.Errorln(s.GetSupplierName(), "1 times", "keyword:", keyword, err)
					continue
				}
			}
		}

		// 第二级界面，有多少个字幕
		s.log.Debugln(s.GetSupplierName(), "step 1", filmDetailPageUrl)
		subResult, err := s.step1(browser, filmDetailPageUrl)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "step 1", filmDetailPageUrl, err)
			continue
		}

		if AllSeasonSubResult.Title == "" {
			AllSeasonSubResult = subResult
		} else {
			AllSeasonSubResult.SubInfos = append(AllSeasonSubResult.SubInfos, subResult.SubInfos...)
		}
	}
	// 找到最大的优先级的字幕下载
	sort.Sort(SortByPriority{AllSeasonSubResult.SubInfos})
	// 找到那些 Eps 需要下载字幕的
	subInfoNeedDownload := s.whichEpisodeNeedDownloadSub(seriesInfo, AllSeasonSubResult)
	// 剩下的部分跟 GetSubListFroKeyword 一样，就是去下载了
	outSubInfoList := s.whichSubInfoNeedDownload(browser, subInfoNeedDownload, err)

	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return outSubInfoList, nil
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	panic("not implemented")
}

func (s *Supplier) getSubListFromMovie(browser *rod.Browser, fileFPath string) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), fileFPath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), fileFPath, "Start...")
	/*
		虽然是传入视频文件路径，但是其实需要读取对应的视频文件目录下的
		movie.xml 以及 *.nfo，找到 IMDB id
		优先通过 IMDB id 去查找字幕
		如果找不到，再靠文件名提取影片名称去查找
	*/
	// 得到这个视频文件名中的信息
	info, _, err := decode.GetVideoInfoFromFileFullPath(fileFPath)
	if err != nil {
		return nil, err
	}
	s.log.Debugln(s.GetSupplierName(), fileFPath, "GetVideoInfoFromFileFullPath -> Title:", info.Title)
	// 找到这个视频文件，尝试得到 IMDB ID
	// 目前测试来看，加入 年 这个关键词去搜索，对 2020 年后的影片有利，因为网站有统一的详细页面了，而之前的，没有，会影响识别
	// 所以，year >= 2020 年，则可以多加一个关键词（年）去搜索影片
	imdbInfo, err := decode.GetImdbInfo4Movie(fileFPath)
	if err != nil {
		// 允许的错误，跳过，继续进行文件名的搜索
		s.log.Errorln("model.GetImdbInfo", err)
	} else {
		s.log.Debugln(s.GetSupplierName(), fileFPath, "GetImdbInfo4Movie -> Title:", imdbInfo.Title)
		s.log.Debugln(s.GetSupplierName(), fileFPath, "GetImdbInfo4Movie -> OriginalTitle:", imdbInfo.OriginalTitle)
		s.log.Debugln(s.GetSupplierName(), fileFPath, "GetImdbInfo4Movie -> Year:", imdbInfo.Year)
		s.log.Debugln(s.GetSupplierName(), fileFPath, "GetImdbInfo4Movie -> ImdbId:", imdbInfo.ImdbId)
	}

	var subInfoList []supplier.SubInfo

	if imdbInfo.ImdbId != "" {
		// 先用 imdb id 找
		s.log.Debugln(s.GetSupplierName(), fileFPath, "getSubListFromKeyword -> Search By IMDB ID:", imdbInfo.ImdbId)
		subInfoList, err = s.getSubListFromKeyword(browser, imdbInfo.ImdbId)
		if err != nil {
			// 允许的错误，跳过，继续进行文件名的搜索
			s.log.Errorln(s.GetSupplierName(), "keyword:", imdbInfo.ImdbId)
			s.log.Errorln("getSubListFromKeyword", "IMDBID can not found sub", fileFPath, err)
		}

		s.log.Debugln(s.GetSupplierName(), fileFPath, "getSubListFromKeyword -> Search By IMDB ID, subInfoList Count:", len(subInfoList))
		// 如果有就优先返回
		if len(subInfoList) > 0 {
			return subInfoList, nil
		}
	}
	// 如果没有，那么就用文件名查找
	searchKeyword := my_util.VideoNameSearchKeywordMaker(s.log, info.Title, imdbInfo.Year)

	s.log.Debugln(s.GetSupplierName(), fileFPath, "VideoNameSearchKeywordMaker Keyword:", searchKeyword)

	subInfoList, err = s.getSubListFromKeyword(browser, searchKeyword)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "keyword:", searchKeyword)
		return nil, err
	}

	s.log.Debugln(s.GetSupplierName(), fileFPath, "getSubListFromKeyword -> Search By Keyword, subInfoList Count:", len(subInfoList))
	return subInfoList, nil
}

// getSubListFromKeyword 目前是给电影使用的，搜索返回的字幕列表可能很多，需要挑选一下，比如 Top 1 下来就好了
func (s *Supplier) getSubListFromKeyword(browser *rod.Browser, keyword string) ([]supplier.SubInfo, error) {

	s.log.Infoln("Search Keyword:", keyword)
	var outSubInfoList []supplier.SubInfo
	// 第一级界面，找到影片的详情界面
	filmDetailPageUrl, err := s.step0(browser, keyword)
	if err != nil {
		return nil, err
	}
	s.log.Debugln(s.GetSupplierName(), "getSubListFromKeyword -> step0 -> filmDetailPageUrl:", filmDetailPageUrl)
	// 第二级界面，有多少个字幕
	subResult, err := s.step1(browser, filmDetailPageUrl)
	if err != nil {
		return nil, err
	}
	// 第三级界面，单个字幕详情
	// 找到最大的优先级的字幕下载
	sort.Sort(SortByPriority{subResult.SubInfos})

	// 强制把找到的列表缩少到 Top 5
	subResult.SubInfos = subResult.SubInfos[:5]

	s.log.Debugln(s.GetSupplierName(), "getSubListFromKeyword -> step1 -> subResult.Title:", subResult.Title)
	s.log.Debugln(s.GetSupplierName(), "getSubListFromKeyword -> step1 -> subResult.OtherName:", subResult.OtherName)
	for i, info := range subResult.SubInfos {
		s.log.Debugln(s.GetSupplierName(), "getSubListFromKeyword -> step1 -> info.Name", i, info.Name)
		s.log.Debugln(s.GetSupplierName(), "getSubListFromKeyword -> step1 -> info.DownloadUrl:", i, info.DownloadUrl)
		s.log.Debugln(s.GetSupplierName(), "getSubListFromKeyword -> step1 -> info.DetailUrl:", i, info.DetailUrl)
		s.log.Debugln(s.GetSupplierName(), "getSubListFromKeyword -> step1 -> info.DownloadTimes:", i, info.DownloadTimes)
	}

	outSubInfoList = s.whichSubInfoNeedDownload(browser, subResult.SubInfos, err)

	return outSubInfoList, nil
}

func (s *Supplier) whichEpisodeNeedDownloadSub(seriesInfo *series.SeriesInfo, AllSeasonSubResult SubResult) []SubInfo {
	// 字幕很多，考虑效率，需要做成字典
	// key SxEx - SubInfos
	var allSubDict = make(map[string]SubInfos)
	// 全季的字幕列表
	var oneSeasonSubDict = make(map[string]SubInfos)
	for _, subInfo := range AllSeasonSubResult.SubInfos {
		_, season, episode, err := decode.GetSeasonAndEpisodeFromSubFileName(subInfo.Name)
		if err != nil {
			s.log.Errorln("whichEpisodeNeedDownloadSub.GetVideoInfoFromFileFullPath", subInfo.Name, err)
			continue
		}
		subInfo.Season = season
		subInfo.Episode = episode
		epsKey := my_util.GetEpisodeKeyName(season, episode)
		_, ok := allSubDict[epsKey]
		if ok == false {
			// 初始化
			allSubDict[epsKey] = SubInfos{}
			if season != 0 && episode == 0 {
				oneSeasonSubDict[epsKey] = SubInfos{}
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
	var subInfoNeedDownload = make([]SubInfo, 0)
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

func (s *Supplier) whichSubInfoNeedDownload(browser *rod.Browser, subInfos SubInfos, err error) []supplier.SubInfo {

	var outSubInfoList = make([]supplier.SubInfo, 0)
	for i := range subInfos {

		err = s.step2(browser, &subInfos[i])
		if err != nil {
			s.log.Error(s.GetSupplierName(), "step 2", subInfos[i].Name, err)
			continue
		}
		s.log.Debugln(s.GetSupplierName(), "whichSubInfoNeedDownload -> step2 -> info.SubDownloadPageUrl:", i, subInfos[i].SubDownloadPageUrl)
	}

	// TODO 这里需要考虑，可以设置为高级选项，不够就用 unknow 来补充
	// 首先过滤出中文的字幕，同时需要满足是支持的字幕
	var tmpSubInfo = make([]SubInfo, 0)
	for _, subInfo := range subInfos {
		tmpLang := language.LangConverter4Sub_Supplier(subInfo.Lang)
		if language.HasChineseLang(tmpLang) == true && sub_parser_hub.IsSubTypeWanted(subInfo.Ext) == true {
			tmpSubInfo = append(tmpSubInfo, subInfo)
		}
	}

	// 看字幕够不够
	if len(tmpSubInfo) < s.topic {
		for _, subInfo := range subInfos {
			if len(tmpSubInfo) >= s.topic {
				break
			}
			tmpLang := language.LangConverter4Sub_Supplier(subInfo.Lang)
			if language.HasChineseLang(tmpLang) == false {
				tmpSubInfo = append(tmpSubInfo, subInfo)
			}
		}
	}

	s.log.Debugln(s.GetSupplierName(), "step2 -> tmpSubInfo.Count", len(tmpSubInfo))
	for i, info := range tmpSubInfo {

		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.Name:", i, info.Name)
		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.DownloadUrl:", i, info.DownloadUrl)
		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.DetailUrl:", i, info.DetailUrl)
		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.DownloadTimes:", i, info.DownloadTimes)
		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.SubDownloadPageUrl:", i, info.SubDownloadPageUrl)
	}

	// 看字幕是不是太多了，超出 topic 的限制了
	if len(tmpSubInfo) > s.topic {
		tmpSubInfo = tmpSubInfo[:s.topic]
	}
	s.log.Debugln(s.GetSupplierName(), "step2 -> tmpSubInfo.Count with topic limit", len(tmpSubInfo))
	for i, info := range tmpSubInfo {

		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.Name:", i, info.Name)
		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.DownloadUrl:", i, info.DownloadUrl)
		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.DetailUrl:", i, info.DetailUrl)
		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.DownloadTimes:", i, info.DownloadTimes)
		s.log.Debugln(s.GetSupplierName(), "ChineseSubs -> tmpSubInfo.SubDownloadPageUrl:", i, info.SubDownloadPageUrl)
	}

	// 第四级界面，具体字幕下载
	for i, subInfo := range tmpSubInfo {

		s.log.Debugln(s.GetSupplierName(), "GetEx:", i, subInfo.SubDownloadPageUrl)

		getSubInfo, err := s.fileDownloader.GetEx(s.GetSupplierName(), browser, subInfo.SubDownloadPageUrl, int64(i), subInfo.Season, subInfo.Episode, s.DownFile)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "GetEx", "GetEx", subInfo.Name, subInfo.Season, subInfo.Episode, err)
			continue
		}

		outSubInfoList = append(outSubInfoList, *getSubInfo)
	}

	for i, info := range outSubInfoList {
		s.log.Debugln(s.GetSupplierName(), "DownFile -> Downloaded File Info", i, "FileName:", info.Name, "FileUrl:", info.FileUrl)
	}

	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return outSubInfoList
}

// step0 先在查询界面找到字幕对应第一个影片的详情界面，需要解决自定义错误 ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound
func (s *Supplier) step0(browser *rod.Browser, keyword string) (string, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("zimuku_step0", err.Error())
		}
	}()

	desUrl := fmt.Sprintf(s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl+common2.SubZiMuKuSearchFormatUrl, url.QueryEscape(keyword))
	result, page, err := rod_helper.HttpGetFromBrowser(browser, desUrl, s.tt)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = page.Close()
	}()
	// 找到对应影片的详情界面
	re := regexp.MustCompile(`<p\s+class="tt\s+clearfix"><a\s+href="(/subs/[\w]+\.html)"\s+target="_blank"><b>(.*?)</b></a></p>`)
	matched := re.FindAllStringSubmatch(result, -1)
	if matched == nil || len(matched) < 1 {
		return "", common2.ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound
	}
	// 影片的详情界面 url
	filmDetailPageUrl := matched[0][1]
	return filmDetailPageUrl, nil
}

// step1 分析详情界面，找到有多少个字幕
func (s *Supplier) step1(browser *rod.Browser, filmDetailPageUrl string) (SubResult, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("zimuku_step1", err.Error())
		}
	}()

	var subResult SubResult
	subResult.SubInfos = SubInfos{}

	filmDetailPageUrl = my_util.AddBaseUrl(s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl, filmDetailPageUrl)

	result, page, err := rod_helper.HttpGetFromBrowser(browser, filmDetailPageUrl, s.tt)
	if err != nil {
		return subResult, err
	}
	defer func() {
		_ = page.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		return SubResult{}, err
	}

	counterIndex := 3
	// 先找到页面”下载“关键词是第几列，然后下面的下载量才能正确的解析。否则，电影是[3]，而在剧集中，因为多了字幕组的筛选，则为[4]
	doc.Find("#subtb thead tr th").Each(func(i int, th *goquery.Selection) {
		if th.Text() == "下载" {
			counterIndex = i
		}
	})

	doc.Find("#subtb tbody tr").Each(func(i int, tr *goquery.Selection) {
		// 字幕下载页面地址
		href, exists := tr.Find("a").Attr("href")
		if !exists {
			return
		}
		// 标题
		title, exists := tr.Find("a").Attr("title")
		if !exists {
			return
		}
		// 扩展名
		ext := tr.Find(".label-info").Text()
		// 作者信息
		authorInfos := tr.Find(".gray")
		authorInfo := ""
		authorInfos.Each(func(a_i int, a_lb *goquery.Selection) {
			authorInfo += a_lb.Text() + "，"
		})
		authorInfoLen := len(authorInfo)
		if authorInfoLen > 0 {
			authorInfo = authorInfo[0 : authorInfoLen-3]
		}
		// 语言
		lang, exists := tr.Find("img").First().Attr("alt")
		if !exists {
			lang = ""
		}
		// 投票
		rate, exists := tr.Find(".rating-star").First().Attr("data-original-title")
		if !exists {
			rate = ""
		}
		vote, err := decode.GetNumber2Float(rate)
		if err != nil {
			return
		}
		// 下载次数统计
		downCountNub := 0
		downCount := tr.Find("td").Eq(counterIndex).Text()
		if strings.Contains(downCount, "万") {
			fNumb, err := decode.GetNumber2Float(downCount)
			if err != nil {
				return
			}
			downCountNub = int(fNumb * 10000)
		} else {
			downCountNub, err = decode.GetNumber2int(downCount)
			if err != nil {
				return
			}
		}

		var subInfo SubInfo
		subResult.Title = title
		subInfo.Name = title
		subInfo.DetailUrl = href
		subInfo.Ext = ext
		subInfo.AuthorInfo = authorInfo
		subInfo.Lang = lang
		subInfo.DownloadTimes = downCountNub

		subInfo.Score = vote
		// 计算优先级
		subInfo.Priority = subInfo.Score * float32(subInfo.DownloadTimes)

		subResult.SubInfos = append(subResult.SubInfos, subInfo)
	})
	return subResult, nil
}

// step2 第二级界面，单个字幕详情，需要判断 ZiMuKuDownloadUrlStep2NotFound 这个自定义错误
func (s *Supplier) step2(browser *rod.Browser, subInfo *SubInfo) error {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("zimuku_step2", err.Error())
		}
	}()
	detailUrl := my_util.AddBaseUrl(s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl, subInfo.DetailUrl)
	result, page, err := rod_helper.HttpGetFromBrowser(browser, detailUrl, s.tt)
	if err != nil {
		return err
	}
	defer func() {
		_ = page.Close()
	}()
	// 找到下载地址
	re := regexp.MustCompile(`<a\s+id="down1"\s+href="([^"]*/dld/[\w]+\.html)"`)
	matched := re.FindAllStringSubmatch(result, -1)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		s.log.Warnln("Step2,sub download url not found", detailUrl)
		return common2.ZiMuKuDownloadUrlStep2NotFound
	}
	if strings.Contains(matched[0][1], "://") {
		subInfo.SubDownloadPageUrl = matched[0][1]
	} else {
		subInfo.SubDownloadPageUrl = fmt.Sprintf("%s%s", s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl, matched[0][1])
	}
	return nil
}

// DownFile 第三级界面，具体字幕下载 ZiMuKuDownloadUrlStep3NotFound ZiMuKuDownloadUrlDownFileFailed
func (s *Supplier) DownFile(browser *rod.Browser, subDownloadPageUrl string, TopN int64, Season, Episode int) (*supplier.SubInfo, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("zimuku_DownFile", err.Error())
		}
	}()
	subDownloadPageFullUrl := my_util.AddBaseUrl(s.settings.AdvancedSettings.SuppliersSettings.Zimuku.RootUrl, subDownloadPageUrl)
	result, page, err := rod_helper.HttpGetFromBrowser(browser, subDownloadPageFullUrl, s.tt)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = page.Close()
	}()
	re := regexp.MustCompile(`<li><a\s+rel="nofollow"\s+href="([^"]*/download/[^"]+)"`)
	matched := re.FindAllStringSubmatch(result, -1)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		s.log.Debugln("Step3,sub download url not found", subDownloadPageFullUrl)
		return nil, common2.ZiMuKuDownloadUrlStep3NotFound
	}

	fileName := ""
	fileByte := []byte{0}
	downloadSuccess := false
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
		element := page.MustElement(btnClickDownload)
		// 直接可以下载
		element.MustClick()
		fileByte, fileName, err = getDownloadFile()
		if err != nil {
			return
		}

		downloadSuccess = true
	})
	if err != nil {
		s.log.Errorln("ZiMuKu DownFile DownloadFile", err)
		return nil, err
	}
	if downloadSuccess == true {
		s.log.Debugln("Step3,DownFile, FileName:", fileName, "DataLen:", len(fileByte))

		inSubInfo := supplier.NewSubInfo(s.GetSupplierName(), 1, fileName, language2.ChineseSimple,
			subDownloadPageUrl, 0, 0, filepath.Ext(fileName), fileByte)

		inSubInfo.TopN = TopN
		inSubInfo.Season = Season
		inSubInfo.Episode = Episode

		return inSubInfo, nil
	} else {
		s.log.Debugln("Step3,sub download url not found", subDownloadPageFullUrl)
		return nil, common2.ZiMuKuDownloadUrlDownFileFailed
	}
}

type SubResult struct {
	Title     string   // 字幕的标题
	OtherName string   // 影片又名
	SubInfos  SubInfos // 字幕的列表
}

type SubInfo struct {
	Name               string  // 字幕的名称
	Lang               string  // 语言
	AuthorInfo         string  // 作者
	Ext                string  // 后缀名
	Score              float32 // 评分
	DownloadTimes      int     // 下载的次数
	Priority           float32 // 优先级，使用评分和次数乘积而来，类似于 Score 投票
	DetailUrl          string  // 字幕的详情界面，需要再次分析具体的下载地址，地址需要拼接网站的根地址上去
	SubDownloadPageUrl string  // 字幕的具体的下载页面，会有多个下载可用的链接
	DownloadUrl        string  // 字幕的下载地址
	Season             int     // 第几季，默认-1
	Episode            int     // 第几集，默认-1
}

// SubInfos 实现自定义排序
type SubInfos []SubInfo

func (s SubInfos) Len() int {
	return len(s)
}
func (s SubInfos) Less(i, j int) bool {
	return s[i].Priority > s[j].Priority
}
func (s SubInfos) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type SortByPriority struct{ SubInfos }

// Less 根据元素的优先级降序排序
func (s SortByPriority) Less(i, j int) bool {
	return s.SubInfos[i].Priority > s.SubInfos[j].Priority
}

const btnClickDownload = "a.btn-danger"
