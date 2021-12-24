package zimuku

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/Tnze/go.num/v2/zh"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	language2 "github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Supplier struct {
	reqParam types.ReqParam
	log      *logrus.Logger
	topic    int
}

func NewSupplier(_reqParam ...types.ReqParam) *Supplier {

	sup := Supplier{}
	sup.log = log_helper.GetLogger()
	sup.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		sup.reqParam = _reqParam[0]
		if sup.reqParam.Topic > 0 && sup.reqParam.Topic != sup.topic {
			sup.topic = sup.reqParam.Topic
		}
	}
	return &sup
}

func (s Supplier) GetSupplierName() string {
	return common.SubSiteZiMuKu
}

func (s Supplier) GetReqParam() types.ReqParam {
	return s.reqParam
}

func (s Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {
	return s.getSubListFromMovie(filePath)
}

func (s Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), seriesInfo.Name, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), seriesInfo.Name, "Start...")

	var err error
	/*
		去网站搜索的时候，有个比较由意思的逻辑，有些剧集，哪怕只有一季，sonarr 也会给它命名为 Season 1
		但是在 zimuku 搜索的时候，如果你加上 XXX 第一季 就搜索不出来，那么目前比较可行的办法是查询两次
		第一次优先查询 XXX 第一季 ，如果返回的列表是空的，那么再查询 XXX
	*/
	// 这里打算牺牲效率，提高代码的复用度，不然后续得维护一套电影的查询逻辑，一套剧集的查询逻辑
	// 比如，其实可以搜索剧集名称，应该可以得到多个季的列表，然后分析再继续
	// 现在粗暴点，直接一季搜索一次，跟电影的搜索一样，在首个影片就停止，然后继续往下
	AllSeasonSubResult := SubResult{}
	for value := range seriesInfo.SeasonDict {
		// 第一级界面，找到影片的详情界面
		keyword := seriesInfo.Name + " 第" + zh.Uint64(value).String() + "季"
		s.log.Debugln(s.GetSupplierName(), "step 0", "0 times", "keyword:", keyword)
		filmDetailPageUrl, err := s.step0(keyword)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "step 0", "0 times", "keyword:", keyword, err)
			// 如果只是搜索不到，则继续换关键词
			if err != common.ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound {
				s.log.Errorln(s.GetSupplierName(), "ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound", keyword, err)
				continue
			}
			keyword = seriesInfo.Name
			s.log.Debugln(s.GetSupplierName(), "step 0", "1 times", "keyword:", keyword)
			filmDetailPageUrl, err = s.step0(keyword)
			if err != nil {
				s.log.Errorln(s.GetSupplierName(), "1 times", "keyword:", keyword, err)
				continue
			}
		}
		// 第二级界面，有多少个字幕
		s.log.Debugln(s.GetSupplierName(), "step 1", filmDetailPageUrl)
		subResult, err := s.step1(filmDetailPageUrl)
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
	outSubInfoList := s.whichSubInfoNeedDownload(subInfoNeedDownload, err)

	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return outSubInfoList, nil
}

func (s Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	panic("not implemented")
}

func (s Supplier) getSubListFromMovie(fileFPath string) ([]supplier.SubInfo, error) {

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
	// 找到这个视频文件，尝试得到 IMDB ID
	// 目前测试来看，加入 年 这个关键词去搜索，对 2020 年后的影片有利，因为网站有统一的详细页面了，而之前的，没有，会影响识别
	// 所以，year >= 2020 年，则可以多加一个关键词（年）去搜索影片
	imdbInfo, err := decode.GetImdbInfo4Movie(fileFPath)
	if err != nil {
		// 允许的错误，跳过，继续进行文件名的搜索
		s.log.Errorln("model.GetImdbInfo", err)
	}
	var subInfoList []supplier.SubInfo

	if imdbInfo.ImdbId != "" {
		// 先用 imdb id 找
		subInfoList, err = s.getSubListFromKeyword(imdbInfo.ImdbId)
		if err != nil {
			// 允许的错误，跳过，继续进行文件名的搜索
			s.log.Errorln(s.GetSupplierName(), "keyword:", imdbInfo.ImdbId)
			s.log.Errorln("getSubListFromKeyword", "IMDBID can not found sub", fileFPath, err)
		}
		// 如果有就优先返回
		if len(subInfoList) > 0 {
			return subInfoList, nil
		}
	}

	// 如果没有，那么就用文件名查找
	searchKeyword := my_util.VideoNameSearchKeywordMaker(info.Title, imdbInfo.Year)
	subInfoList, err = s.getSubListFromKeyword(searchKeyword)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "keyword:", searchKeyword)
		return nil, err
	}

	return subInfoList, nil
}

func (s Supplier) getSubListFromKeyword(keyword string) ([]supplier.SubInfo, error) {

	var outSubInfoList []supplier.SubInfo
	// 第一级界面，找到影片的详情界面
	filmDetailPageUrl, err := s.step0(keyword)
	if err != nil {
		return nil, err
	}
	// 第二级界面，有多少个字幕
	subResult, err := s.step1(filmDetailPageUrl)
	if err != nil {
		return nil, err
	}
	// 第三级界面，单个字幕详情
	// 找到最大的优先级的字幕下载
	sort.Sort(SortByPriority{subResult.SubInfos})

	outSubInfoList = s.whichSubInfoNeedDownload(subResult.SubInfos, err)

	return outSubInfoList, nil
}

func (s Supplier) whichEpisodeNeedDownloadSub(seriesInfo *series.SeriesInfo, AllSeasonSubResult SubResult) []SubInfo {
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
		} else {
			s.log.Infoln(s.GetSupplierName(), "Not Find Sub can be download",
				epsInfo.Title, epsInfo.Season, epsInfo.Episode)
		}
	}
	// 全季的字幕列表，也拼进去，后面进行下载
	for _, infos := range oneSeasonSubDict {
		subInfoNeedDownload = append(subInfoNeedDownload, infos[0])
	}

	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return subInfoNeedDownload
}

func (s Supplier) whichSubInfoNeedDownload(subInfos SubInfos, err error) []supplier.SubInfo {

	var outSubInfoList = make([]supplier.SubInfo, 0)
	for i := range subInfos {
		err = s.step2(&subInfos[i])
		if err != nil {
			s.log.Error(s.GetSupplierName(), "step 2", subInfos[i].Name, err)
			continue
		}
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

	// 第四级界面，具体字幕下载
	for i, subInfo := range tmpSubInfo {
		fileName, data, err := s.step3(subInfo.SubDownloadPageUrl)
		if err != nil {
			s.log.Error(s.GetSupplierName(), "step 3", err)
			continue
		}
		// 默认都是包含中文字幕的，然后具体使用的时候再进行区分

		oneSubInfo := supplier.NewSubInfo(s.GetSupplierName(), int64(i), fileName, language2.ChineseSimple, my_util.AddBaseUrl(common.SubZiMuKuRootUrl, subInfo.SubDownloadPageUrl), 0,
			0, filepath.Ext(fileName), data)

		oneSubInfo.Season = subInfo.Season
		oneSubInfo.Episode = subInfo.Episode
		outSubInfoList = append(outSubInfoList, *oneSubInfo)
	}

	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return outSubInfoList
}

// step0 先在查询界面找到字幕对应第一个影片的详情界面，需要解决自定义错误 ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound
func (s Supplier) step0(keyword string) (string, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("zimuku_step0", err.Error())
		}
	}()
	httpClient := my_util.NewHttpClient(s.reqParam)
	// 第一级界面，有多少个字幕
	resp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"q": keyword,
		}).
		Get(common.SubZiMuKuSearchUrl)
	if err != nil {
		return "", err
	}
	// 找到对应影片的详情界面
	re := regexp.MustCompile(`<p\s+class="tt\s+clearfix"><a\s+href="(/subs/[\w]+\.html)"\s+target="_blank"><b>(.*?)</b></a></p>`)
	matched := re.FindAllStringSubmatch(resp.String(), -1)
	if matched == nil || len(matched) < 1 {
		return "", common.ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound
	}
	// 影片的详情界面 url
	filmDetailPageUrl := matched[0][1]
	return filmDetailPageUrl, nil
}

// step1 分析详情界面，找到有多少个字幕
func (s Supplier) step1(filmDetailPageUrl string) (SubResult, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("zimuku_step1", err.Error())
		}
	}()
	filmDetailPageUrl = my_util.AddBaseUrl(common.SubZiMuKuRootUrl, filmDetailPageUrl)
	httpClient := my_util.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().
		Get(filmDetailPageUrl)
	if err != nil {
		return SubResult{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return SubResult{}, err
	}
	var subResult SubResult
	subResult.SubInfos = SubInfos{}

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
		rate, exists := tr.Find(".rating-star").First().Attr("title")
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
func (s Supplier) step2(subInfo *SubInfo) error {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("zimuku_step2", err.Error())
		}
	}()
	detailUrl := my_util.AddBaseUrl(common.SubZiMuKuRootUrl, subInfo.DetailUrl)
	httpClient := my_util.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().
		Get(detailUrl)
	if err != nil {
		return err
	}
	// 找到下载地址
	re := regexp.MustCompile(`<a\s+id="down1"\s+href="([^"]*/dld/[\w]+\.html)"`)
	matched := re.FindAllStringSubmatch(resp.String(), -1)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		s.log.Warnln("Step2,sub download url not found", detailUrl)
		return common.ZiMuKuDownloadUrlStep2NotFound
	}
	if strings.Contains(matched[0][1], "://") {
		subInfo.SubDownloadPageUrl = matched[0][1]
	} else {
		subInfo.SubDownloadPageUrl = fmt.Sprintf("%s%s", common.SubZiMuKuRootUrl, matched[0][1])
	}
	return nil
}

// step3 第三级界面，具体字幕下载 ZiMuKuDownloadUrlStep3NotFound ZiMuKuDownloadUrlStep3AllFailed
func (s Supplier) step3(subDownloadPageUrl string) (string, []byte, error) {
	var err error
	defer func() {
		if err != nil {
			notify_center.Notify.Add("zimuku_step3", err.Error())
		}
	}()
	subDownloadPageUrl = my_util.AddBaseUrl(common.SubZiMuKuRootUrl, subDownloadPageUrl)
	httpClient := my_util.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().
		Get(subDownloadPageUrl)
	if err != nil {
		return "", nil, err
	}
	re := regexp.MustCompile(`<li><a\s+rel="nofollow"\s+href="([^"]*/download/[^"]+)"`)
	matched := re.FindAllStringSubmatch(resp.String(), -1)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		s.log.Debugln("Step3,sub download url not found", subDownloadPageUrl)
		return "", nil, common.ZiMuKuDownloadUrlStep3NotFound
	}
	var filename string
	var data []byte

	s.reqParam.Referer = subDownloadPageUrl
	for i := 0; i < len(matched); i++ {
		data, filename, err = my_util.DownFile(my_util.AddBaseUrl(common.SubZiMuKuRootUrl, matched[i][1]), s.reqParam)
		if err != nil {
			s.log.Errorln("ZiMuKu step3 DownloadFile", err)
			continue
		}
		return filename, data, nil
	}
	s.log.Debugln("Step3,sub download url not found", subDownloadPageUrl)
	return "", nil, common.ZiMuKuDownloadUrlStep3AllFailed
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
