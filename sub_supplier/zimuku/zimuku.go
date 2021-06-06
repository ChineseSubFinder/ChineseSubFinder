package zimuku

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier"
	"github.com/go-resty/resty/v2"
	"path/filepath"
	"strings"
)

type Supplier struct {

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

	httpClient := resty.New()
	httpClient.SetTimeout(common.HTMLTimeOut)
	if httpProxy != "" {
		httpClient.SetProxy(httpProxy)
	}
	httpClient.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent": "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	})
	// 第一级界面，有多少个字幕
	resp, err := httpClient.R().
		SetQueryParams(map[string]string{
			"q": keyword,
		}).
		Get(common.SubZimukuRootUrl)
	if err != nil {
		return nil, err
	}
	//println(resp.String())
	// 解析 html
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, err
	}
	doc.Find("div div table tbody tr").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		aa := s.Find("a[href]")
		//comicPicEpisode := aa.Text()
		println(aa.Text())
	})

	// 第二级界面，单个字幕详情

	// 第三级界面，具体字幕下载


	return nil, nil
}

type SubResult struct {
	Title string
	OtherName string
	SubList []SubInfo
}

type SubInfo struct {
	Score			float32
	DownloadTimes 	int
	Url				string

}