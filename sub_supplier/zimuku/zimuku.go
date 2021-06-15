package zimuku

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Supplier struct {
	reqParam common.ReqParam
	log      *logrus.Logger
	topic    int
}

func NewSupplier(_reqParam ...common.ReqParam) *Supplier {

	sup := Supplier{}
	sup.log = model.GetLogger()
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

func (s Supplier) GetSubListFromFile4Movie(filePath string) ([]common.SupplierSubInfo, error){
	return s.GetSubListFromFile(filePath)
}

func (s Supplier) GetSubListFromFile4Series(seriesPath string) ([]common.SupplierSubInfo, error) {
	// 只考虑 IMDB 去查询，文件名目前发现可能会跟电影重复，导致很麻烦，本来也有前置要求要削刮器处理的
	// TODO ！GetImdbInfo 需要输出 title 以及 originaltitle
	imdbId, _, err := model.GetImdbInfo(seriesPath)
	if err != nil {
		return nil, err
	}
	// 使用 IMDB 去
	imdbInfo, err := model.GetVideoInfoFromIMDB(imdbId)
	if err != nil {
		return nil, err
	}
	subInfoList, err := s.GetSubListFromKeyword(imdbInfo.Name)
	if err != nil {
		return nil, err
	}

	return subInfoList, nil
}

func (s Supplier) GetSubListFromFile4Anime(animePath string) ([]common.SupplierSubInfo, error){
	panic("not implemented")
}

func (s Supplier) GetSubListFromFile(filePath string) ([]common.SupplierSubInfo, error) {

	/*
		虽然是传入视频文件路径，但是其实需要读取对应的视频文件目录下的
		movie.xml 以及 *.nfo，找到 IMDB id
		优先通过 IMDB id 去查找字幕
		如果找不到，再靠文件名提取影片名称去查找
	*/
	// 得到这个视频文件名中的信息
	info, err := model.GetVideoInfoFromFileName(filePath)
	if err != nil {
		return nil, err
	}
	// 找到这个视频文件，然后读取它目录下的文件，尝试得到 IMDB ID
	fileRootDirPath := filepath.Dir(filePath)
	// 目前测试来看，加入 年 这个关键词去搜索，对 2020 年后的影片有利，因为网站有统一的详细页面了，而之前的，没有，会影响识别
	// 所以，year >= 2020 年，则可以多加一个关键词（年）去搜索影片
	imdbInfo, err := model.GetImdbInfo(fileRootDirPath)
	if err != nil {
		// 允许的错误，跳过，继续进行文件名的搜索
		s.log.Errorln("model.GetImdbInfo", err)
	}
	var subInfoList []common.SupplierSubInfo

	if imdbInfo.ImdbId != "" {
		// 先用 imdb id 找
		subInfoList, err = s.GetSubListFromKeyword(imdbInfo.ImdbId)
		if err != nil {
			// 允许的错误，跳过，继续进行文件名的搜索
			s.log.Errorln("GetSubListFromKeyword", "IMDBID can not found sub", filePath, err)
		}
		// 如果有就优先返回
		if len(subInfoList) >0 {
			return subInfoList, nil
		}
	}

	// 如果没有，那么就用文件名查找
	searchKeyword := model.VideoNameSearchKeywordMaker(info.Title, imdbInfo.Year)
	subInfoList, err = s.GetSubListFromKeyword(searchKeyword)
	if err != nil {
		return nil, err
	}

	return subInfoList, nil
}

func (s Supplier) GetSubListFromKeyword(keyword string) ([]common.SupplierSubInfo, error) {

	var outSubInfoList []common.SupplierSubInfo
	// 第一级界面，找到影片的详情界面
	filmDetailPageUrl, err := s.Step0(keyword)
	if err != nil {
		return nil, err
	}
	// 第二级界面，有多少个字幕
	subResult, err := s.Step1(filmDetailPageUrl)
	if err != nil {
		return nil, err
	}
	// 第三级界面，单个字幕详情
	// 找到最大的优先级的字幕下载
	sort.Sort(SortByPriority{subResult.SubInfos})

	for i := range subResult.SubInfos {
		err = s.Step2(&subResult.SubInfos[i])
		if err != nil {
			s.log.Error(err)
			continue
		}
	}

	// TODO 这里需要考虑，可以设置为高级选项，不够就用 unknow 来补充
	// 首先过滤出中文的字幕，同时需要满足是支持的字幕
	var tmpSubInfo = make([]SubInfo, 0)
	for _, subInfo := range subResult.SubInfos {
		tmpLang := model.LangConverter(subInfo.Lang)
		if model.HasChineseLang(tmpLang) == true && model.IsSubTypeWanted(subInfo.Ext) == true {
			tmpSubInfo = append(tmpSubInfo, subInfo)
		}
	}
	// 看字幕够不够
	if len(tmpSubInfo) < s.topic {
		for _, subInfo := range subResult.SubInfos {
			if len(tmpSubInfo) >= s.topic {
				break
			}
			tmpLang := model.LangConverter(subInfo.Lang)
			if model.HasChineseLang(tmpLang) == false {
				tmpSubInfo = append(tmpSubInfo, subInfo)
			}
		}
	}

	// 第四级界面，具体字幕下载
	for i, subInfo := range tmpSubInfo {
		fileName, data, err := s.Step3(subInfo.SubDownloadPageUrl)
		if err != nil {
			s.log.Error(err)
			continue
		}
		// 默认都是包含中文字幕的，然后具体使用的时候再进行区分
		outSubInfoList = append(outSubInfoList, *common.NewSupplierSubInfo(s.GetSupplierName(), int64(i), fileName, common.ChineseSimple, model.AddBaseUrl(common.SubZiMuKuRootUrl, subInfo.SubDownloadPageUrl), 0,
			0, filepath.Ext(fileName), data))
	}

	return outSubInfoList, nil
}

// Step0 先在查询界面找到字幕对应第一个影片的详情界面，需要解决自定义错误 ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound
func (s Supplier) Step0(keyword string) (string, error) {
	httpClient := model.NewHttpClient(s.reqParam)
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
	if len(matched) < 1 {
		return "", common.ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound
	}
	// 影片的详情界面 url
	filmDetailPageUrl := matched[0][1]
	return filmDetailPageUrl, nil
}

// Step1 分析详情界面，找到有多少个字幕
func (s Supplier) Step1(filmDetailPageUrl string) (SubResult, error) {
	filmDetailPageUrl = model.AddBaseUrl(common.SubZiMuKuRootUrl, filmDetailPageUrl)
	httpClient := model.NewHttpClient(s.reqParam)
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
		vote, err := model.GetNumber2Float(rate)
		if err != nil {
			return
		}
		// 下载次数统计
		downCountNub := 0
		downCount := tr.Find("td").Eq(3).Text()
		if strings.Contains(downCount, "万") {
			fNumb, err := model.GetNumber2Float(downCount)
			if err != nil {
				return
			}
			downCountNub = int(fNumb * 10000)
		} else {
			downCountNub, err = model.GetNumber2int(downCount)
			if err != nil {
				return
			}
		}

		var subInfo SubInfo
		subResult.Title = title
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

// Step2 第二级界面，单个字幕详情，需要判断 ZiMuKuDownloadUrlStep2NotFound 这个自定义错误
func (s Supplier) Step2(subInfo *SubInfo) error {

	detailUrl := model.AddBaseUrl(common.SubZiMuKuRootUrl, subInfo.DetailUrl)
	httpClient := model.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().
		Get(detailUrl)
	if err != nil {
		return err
	}
	// 找到下载地址
	re := regexp.MustCompile(`<a\s+id="down1"\s+href="([^"]*/dld/[\w]+\.html)"`)
	matched := re.FindAllStringSubmatch(resp.String(), -1)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		s.log.Debug(detailUrl)
		return common.ZiMuKuDownloadUrlStep2NotFound
	}
	if strings.Contains(matched[0][1], "://") {
		subInfo.SubDownloadPageUrl = matched[0][1]
	} else {
		subInfo.SubDownloadPageUrl = fmt.Sprintf("%s%s", common.SubZiMuKuRootUrl, matched[0][1])
	}
	return nil
}

// Step3 第三级界面，具体字幕下载 ZiMuKuDownloadUrlStep3NotFound ZiMuKuDownloadUrlStep3AllFailed
func (s Supplier) Step3(subDownloadPageUrl string) (string, []byte, error) {

	subDownloadPageUrl = model.AddBaseUrl(common.SubZiMuKuRootUrl, subDownloadPageUrl)
	httpClient := model.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().
		Get(subDownloadPageUrl)
	if err != nil {
		return "", nil, err
	}
	re := regexp.MustCompile(`<li><a\s+rel="nofollow"\s+href="([^"]*/download/[^"]+)"`)
	matched := re.FindAllStringSubmatch(resp.String(), -1)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		s.log.Debug(subDownloadPageUrl)
		return "", nil, common.ZiMuKuDownloadUrlStep3NotFound
	}
	var filename string
	var data []byte

	s.reqParam.Referer = subDownloadPageUrl
	for i := 0; i < len(matched); i++ {
		data, filename, err = model.DownFile(model.AddBaseUrl(common.SubZiMuKuRootUrl, matched[i][1]), s.reqParam)
		if err != nil {
			s.log.Errorln("ZiMuKu Step3 DownloadFile", err)
			continue
		}
		return filename, data, nil
	}
	s.log.Debug(subDownloadPageUrl)
	return "", nil, common.ZiMuKuDownloadUrlStep3AllFailed
}

// Step1Discard 第一级界面，有多少个字幕，弃用，直接再搜索出来的结果界面匹配会遇到一个问题，就是 “还有8个字幕，点击查看” 类似此问题
func (s Supplier) Step1Discard(keyword string) (SubResult, error) {
	httpClient := model.NewHttpClient(s.reqParam)
	// 第一级界面，有多少个字幕
	resp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"q": keyword,
		}).
		Get(common.SubZiMuKuSearchUrl)
	if err != nil {
		return SubResult{}, err
	}
	// 解析 html
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return SubResult{}, err
	}
	// 具体解析这个页面
	var subResult SubResult
	subResult.SubInfos = SubInfos{}
	// 这一级找到的是这个关键词出来的所有的影片的信息，可能有多个 Title，但是仅仅处理第一个
	doc.Find("div[class=title]").EachWithBreak(func(i int, selectionTitleRoot *goquery.Selection) bool {

		// 找到"又名"，是第二个 P
		selectionTitleRoot.Find("p").Each(func(i int, s *goquery.Selection) {
			if i == 1 {
				subResult.OtherName = s.Text()
			}
		})
		// 找到字幕的列表，读取相应的信息
		selectionTitleRoot.Find("div table tbody tr").Each(func(i int, sTr *goquery.Selection) {
			aa := sTr.Find("a[href]")
			subDetailUrl, ok := aa.Attr("href")
			var subInfo SubInfo
			if ok {
				// 字幕的标题
				subResult.Title = aa.Text()
				// 字幕的详情界面
				subInfo.DetailUrl = subDetailUrl
				// 找到这个 tr 下面的第二个和第三个 td
				sTr.Find("td").Each(func(i int, sTd *goquery.Selection) {
					if i == 1 {
						// 评分
						vote, ok := sTd.Find("i").Attr("title")
						if ok == false {
							return
						}
						number, err := model.GetNumber2Float(vote)
						if err != nil {
							return
						}
						subInfo.Score = number
					} else if i == 2{
						// 下载量
						number, err := model.GetNumber2int(sTd.Text())
						if err != nil {
							return
						}
						subInfo.DownloadTimes = number
					}
				})
				// 计算优先级
				subInfo.Priority = subInfo.Score * float32(subInfo.DownloadTimes)
				// 加入列表
				subResult.SubInfos = append(subResult.SubInfos, subInfo)
			}

		})
		// EachWithBreak 使用这个，就能阻断继续遍历
		return false
	})
	// 这里要判断，一级界面是否OK 了，不行就返回
	if subResult.Title == "" || len(subResult.SubInfos) == 0 {
		return SubResult{}, nil
	}
	return subResult, nil
}

type SubResult struct {
	Title     string   // 字幕的标题
	OtherName string   // 影片又名
	SubInfos  SubInfos // 字幕的列表
}

type SubInfo struct {
	Lang				string	// 语言
	AuthorInfo			string	// 作者
	Ext					string	// 后缀名
	Score				float32	// 评分
	DownloadTimes 		int		// 下载的次数
	Priority			float32	// 优先级，使用评分和次数乘积而来，类似于 Score 投票
	DetailUrl			string	// 字幕的详情界面，需要再次分析具体的下载地址，地址需要拼接网站的根地址上去
	SubDownloadPageUrl 	string	// 字幕的具体的下载页面，会有多个下载可用的链接
	DownloadUrl			string	// 字幕的下载地址
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