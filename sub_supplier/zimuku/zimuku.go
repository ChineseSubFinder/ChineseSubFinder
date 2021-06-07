package zimuku

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Supplier struct {
	_reqParam common.ReqParam
}

func NewSupplier() *Supplier {
	return &Supplier{}
}

func (s Supplier) GetSubListFromFile(filePath string, httpProxy string) ([]sub_supplier.SubInfo, error) {

	/*
		虽然是传入视频文件路径，但是其实需要读取对应的视频文件目录下的
		movie.xml 以及 *.nfo，找到 IMDB id
		优先通过 IMDB id 去查找字幕
		如果找不到，再靠文件名提取影片名称去查找
	*/
	// 得到这个视频文件名中的信息
	info, err := common.GetVideoInfo(filePath)
	if err != nil {
		return nil, err
	}
	// 找到这个视频文件，然后读取它目录下的文件，尝试得到 IMDB ID
	fileRootDirPath := filepath.Dir(filePath)
	imdbId, err := common.GetImdbId(fileRootDirPath)
	if err != nil {
		return nil, err
	}

	// 先用 imdb id 找
	subInfoList, err := s.GetSubListFromKeyword(imdbId, httpProxy)
	if err != nil {
		return nil, err
	}
	// 如果有就优先返回
	if len(subInfoList) >0 {
		return subInfoList, nil
	}
	// 如果没有，那么就用文件名查找
	subInfoList, err = s.GetSubListFromKeyword(info.Title, httpProxy)
	if err != nil {
		return nil, err
	}

	return subInfoList, nil
}

func (s Supplier) GetSubListFromKeyword(keyword string, httpProxy string) ([]sub_supplier.SubInfo, error) {

	// 第一级界面，有多少个字幕
	subResult, err := s.Step1(keyword, httpProxy)
	if err != nil {
		return nil, err
	}
	// 第二级界面，单个字幕详情
	err = s.Step2(&subResult, httpProxy)
	if err != nil {
		return nil, err
	}
	// 第三级界面，具体字幕下载
	err = s.Step3(&subResult, httpProxy)
	if err != nil {
		return nil, err
	}
	// TODO 需要把查询到的信息转换到 []sub_supplier.SubInfo 再输出
	// 注意要做一次排序，根据优先级
	return nil, nil
}

// Step1 第一级界面，有多少个字幕
func (s Supplier) Step1(keyword string, _reqParam ...common.ReqParam) (SubResult, error) {
	httpClient := common.NewHttpClient()
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
						number, err := common.GetNumber2Folat(vote)
						if err != nil {
							return
						}
						subInfo.Score = number
					} else if i == 2{
						// 下载量
						number, err := common.GetNumber2int(sTd.Text())
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
		return SubResult{}, common.ZiMuKuSearchKeyWordStep1NotFound
	}
	return subResult, nil
}

// Step2 第二级界面，单个字幕详情
func (s Supplier) Step2(subResult *SubResult, httpProxy string) error {
	// 找到最大的优先级的字幕下载
	sort.Sort(SortByPriority{subResult.SubInfos})
	// 排序后的第一个
	httpClient := common.NewHttpClient(httpProxy)
	resp, err := httpClient.R().
		Get(common.SubZiMuKuRootUrl + subResult.SubInfos[0].DetailUrl)
	if err != nil {
		return err
	}
	// 找到下载地址
	re := regexp.MustCompile(`<a\s+id="down1"\s+href="([^"]*/dld/[\w]+\.html)"`)
	matched := re.FindAllStringSubmatch(resp.String(), -1)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		return common.ZiMuKuDownloadUrlStep2NotFound
	}
	if strings.Contains(matched[0][1], "://") {
		subResult.SubInfos[0].SubDownloadPageUrl = matched[0][1]
	} else {
		subResult.SubInfos[0].SubDownloadPageUrl = fmt.Sprintf("%s%s", common.SubZiMuKuRootUrl, matched[0][1])
	}
	return nil
}

// Step3 第三级界面，具体字幕下载
func (s Supplier) Step3(subResult *SubResult, httpProxy string) error {

	subDownloadPageUrl := common.AddBaseUrl(common.SubZiMuKuRootUrl, subResult.SubInfos[0].SubDownloadPageUrl)
	httpClient := common.NewHttpClient(httpProxy)
	resp, err := httpClient.R().
		Get(subDownloadPageUrl)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`<li><a\s+rel="nofollow"\s+href="([^"]*/download/[^"]+)"`)
	matched := re.FindAllStringSubmatch(resp.String(), -1)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		return nil
	}
	var filename string
	var data []byte
	for i := 0; i < len(matched); i++ {

		data, filename, err = common.DownFile(common.AddBaseUrl(common.SubZiMuKuRootUrl, matched[i][1]), common.ReqParam{
			HttpProxy: httpProxy,
			Referer:   subDownloadPageUrl,
		})
		if err != nil {
			println("ZiMuKu Step3 DownloadFile", err)
			continue
		}
	}
	return nil
}

type SubResult struct {
	Title string			// 字幕的标题
	OtherName string		// 影片又名
	SubInfos SubInfos		// 字幕的列表
}

type SubInfo struct {
	Score				float32	// 评分
	DownloadTimes 		int		// 下载的次数
	Priority			float32	// 优先级，使用评分和次数乘积而来
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