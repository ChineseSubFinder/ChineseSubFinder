package subhd

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

type Supplier struct {
	reqParam common.ReqParam
	topic int
}

func NewSupplier(_reqParam ... common.ReqParam) *Supplier {

	sup := Supplier{}
	sup.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		sup.reqParam = _reqParam[0]
		if sup.reqParam.Topic > 0 && sup.reqParam.Topic != sup.topic {
			sup.topic = sup.reqParam.Topic
		}
	}
	return &sup
}

func (s Supplier) GetSubListFromFile(filePath string) ([]sub_supplier.SubInfo, error) {
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
	if err != nil && err != common.CanNotFindIMDBID {
		return nil, err
	}

	var subInfoList []sub_supplier.SubInfo

	if imdbId != "" {
		// 先用 imdb id 找
		subInfoList, err = s.GetSubListFromKeyword(imdbId)
		if err != nil {
			return nil, err
		}
		// 如果有就优先返回
		if len(subInfoList) >0 {
			return subInfoList, nil
		}
	}

	// 如果没有，那么就用文件名查找
	subInfoList, err = s.GetSubListFromKeyword(info.Title)
	if err != nil {
		return nil, err
	}

	return subInfoList, nil
}

func (s Supplier) GetSubListFromKeyword(keyword string) ([]sub_supplier.SubInfo, error) {

	var subInfos  []sub_supplier.SubInfo
	detailPageUrl, err := s.Step0(keyword)
	if err != nil {
		return nil, err
	}
	subList, err := s.Step1(detailPageUrl)
	if err != nil {
		return nil, err
	}

	for _, item := range subList {
		hdContent, err := s.Step2(item.Url)
		if err != nil {
			return nil, err
		}

		var subInfo sub_supplier.SubInfo
		subInfo.Name = hdContent.Filename
		subInfo.Ext = hdContent.Ext
		subInfo.Language = common.ChineseSimple
		subInfo.Vote = 0
		subInfo.FileUrl = common.AddBaseUrl(common.SubSubHDRootUrl, item.Url)
		subInfo.Offset = 0
		subInfo.Data = hdContent.Data

		subInfos = append(subInfos, subInfo)
	}

	return subInfos, nil
}

// Step0 找到这个影片的详情列表
func (s Supplier) Step0(keyword string) (string, error) {

	result, err := s.httpGet(fmt.Sprintf(common.SubSubHDSearchUrl, url.QueryEscape(keyword)))
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`<a\shref="(/d/[\w]+)"><img`)
	matched := re.FindAllStringSubmatch(result, -1)
	if len(matched) < 1 || len(matched[0]) < 2{
		return "",  common.SubHDStep0HrefIsNull
	}
	return matched[0][1], nil
}
// Step1 获取影片的详情字幕列表
func (s Supplier) Step1(detailPageUrl string) ([]HdListItem, error) {
	detailPageUrl = common.AddBaseUrl(common.SubSubHDRootUrl, detailPageUrl)
	result, err := s.httpGet(detailPageUrl)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		return nil, err
	}
	var lists []HdListItem
	doc.Find(".table-sm tr").EachWithBreak(func(i int, tr *goquery.Selection) bool {
		if tr.Find("a.text-dark").Size() == 0 {
			return true
		}
		downUrl, exists := tr.Find("a.text-dark").Eq(0).Attr("href")
		if !exists {
			return true
		}
		title := strings.TrimSpace(tr.Find("a.text-dark").Text())

		downCount, err := common.GetNumber2int(tr.Find("td.p-3").Eq(1).Text())
		if err != nil {
			return true
		}

		ext := ""
		tr.Find(".text-secondary span").Each(func(a_i int, a_lb *goquery.Selection) {
			ext += a_lb.Text() + "，"
		})
		extLen := len(ext)
		if len(ext) > 0 {
			ext = ext[0 : extLen - 3]
		}

		authorInfo := tr.Find("a.text-dark").Eq(2).Text()

		rate := ""

		listItem := HdListItem{}
		listItem.Url = downUrl
		listItem.BaseUrl = common.SubSubHDRootUrl
		listItem.Title = title
		listItem.Ext = ext
		listItem.AuthorInfo = authorInfo
		listItem.Rate = rate
		listItem.DownCount = downCount

		if len(lists) > s.topic {
			return false
		}

		lists = append(lists, listItem)

		return true
	})

	return lists, nil
}
// Step2 下载字幕
func (s Supplier) Step2(subDownloadPageUrl string) (*HdContent, error) {
	subDownloadPageUrl = common.AddBaseUrl(common.SubSubHDRootUrl, subDownloadPageUrl)
	result, err := s.httpGet(subDownloadPageUrl)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		return nil, err
	}
	// 是否有腾讯的防水墙
	matchList := doc.Find("#TencentCaptcha")
	if len(matchList.Nodes) < 1 {
		println("qiang")
	}
	//matchList = doc.Find("#down")
	//if len(matchList.Nodes) < 1 {
	//	println("not found down")
	//}
	postData := make(map[string]string)
	sid, exists := matchList.Attr("sid")
	if !exists {
		return nil, common.SubHDStep2SidIsNull
	}
	postData["sub_id"] = sid
	dToken, exists := matchList.Attr("dtoken1")
	if !exists {
		return nil, common.SubHDStep2DTokenIsNull
	}
	postData["dtoken1"] = dToken
	url2 := fmt.Sprintf("%s%s", common.SubSubHDRootUrl, "/ajax/down_ajax")
	result, err = s.httpPost(url2, postData, subDownloadPageUrl)
	if err != nil {
		return nil, err
	}
	if result == "" || strings.Contains(result, "true") == false {
		return nil, common.SubHDStep2ResultIsNullOrNotTrue
	}
	reg := regexp.MustCompile(`"url":"([^"]+)"`)
	arr := reg.FindStringSubmatch(result)
	if len(arr) == 0 {
		return nil, common.SubHDStep2PostResultGetUrlNotFound
	}
	downUrl := arr[1]
	downUrl = strings.ReplaceAll(downUrl, "\\", "")
	var filename = filepath.Base(downUrl)
	var data []byte
	data, filename, err = common.DownFile(downUrl, s.reqParam)
	if err != nil {
		return nil, err
	}
	return &HdContent{
		Filename: filename,
		Ext:      strings.ToLower(filepath.Ext(filename)),
		Data:     data,
	}, nil
}

func (s Supplier) httpGet(url string) (string, error) {
	s.reqParam.Referer = url
	httpClient := common.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().Get(url)
	if err != nil {
		return "", err
	}
	//搜索验证 点击继续搜索
	if strings.Contains(resp.String(), "搜索验证") {
		println("搜索验证 reload", url)
		return s.httpGet(url)
	}
	return resp.String(), nil
}

func (s Supplier) httpPost(url string, postData map[string]string, referer string) (string, error) {

	s.reqParam.Referer = referer
	httpClient := common.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().
		SetFormData(postData).
		Post(url)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
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
}

type HdContent struct {
	Filename string `json:"filename"`
	Ext      string `json:"ext"`
	Data     []byte `json:"data"`
}