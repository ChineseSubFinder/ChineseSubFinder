package subhd

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
	"image/jpeg"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Supplier struct {
	reqParam    common.ReqParam
	log         *logrus.Logger
	topic       int
	rodLauncher *launcher.Launcher
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
	return common.SubSiteSubHd
}

func (s Supplier) GetSubListFromFile4Movie(filePath string) ([]common.SupplierSubInfo, error){
	return s.GetSubListFromFile(filePath)
}

func (s Supplier) GetSubListFromFile4Series(seriesPath string) ([]common.SupplierSubInfo, error) {
	panic("not implemented")
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
	info, _, err := model.GetVideoInfoFromFileName(filePath)
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

	var subInfos  []common.SupplierSubInfo
	detailPageUrl, err := s.Step0(keyword)
	if err != nil {
		return nil, err
	}
	// 没有搜索到字幕
	if detailPageUrl == "" {
		return nil, nil
	}
	subList, err := s.Step1(detailPageUrl)
	if err != nil {
		return nil, err
	}

	var browser *rod.Browser
	// 是用本地的 Browser 还是远程的，推荐是远程的
	//if s.reqParam.RemoteBrowserDockerURL != "" {
		browser, err = model.NewBrowserFromDocker(s.reqParam.HttpProxy, "ws://192.168.50.135:9222")
	//} else {
	//browser, err = model.NewBrowser(s.reqParam.HttpProxy)
	////}
	if err != nil {
		return nil, err
	}

	for i, item := range subList {
		hdContent, err := s.Step2Ex(browser, item.Url)
		if err != nil {
			return nil, err
		}
		subInfos = append(subInfos, *common.NewSupplierSubInfo(s.GetSupplierName(), int64(i), hdContent.Filename, common.ChineseSimple, model.AddBaseUrl(common.SubSubHDRootUrl, item.Url), 0, 0, hdContent.Ext, hdContent.Data))
	}

	return subInfos, nil
}

// Step0 找到这个影片的详情列表
func (s Supplier) Step0(keyword string) (string, error) {

	result, err := s.httpGet(fmt.Sprintf(common.SubSubHDSearchUrl, url.QueryEscape(keyword)))
	if err != nil {
		return "", err
	}
	// 是否有查找到的结果，至少要有结果。根据这里这样下面才能判断是分析失效了，还是就是没有结果而已
	re := regexp.MustCompile(`共\s*(\d+)\s*条`)
	matched := re.FindAllStringSubmatch(result, -1)
	if len(matched) < 1 {
		return "",  common.SubHDStep0SubCountNotFound
	}
	subCount, err := model.GetNumber2int(matched[0][0])
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
	imgUrl, ok := imgSelection.Attr("src")
	if ok == true{
		imgName := filepath.Base(imgUrl)
		imgExt := filepath.Ext(imgUrl)
		if strings.Contains(imgName, "_") == true {
			items := strings.Split(imgName, "_")
			return "/d/" + items[0], nil
		} else {
			return "/d/" + strings.ReplaceAll(imgName, imgExt, ""), nil
		}
	} else{
		return "",  common.SubHDStep0HrefIsNull
	}
	//re = regexp.MustCompile(`<a\shref="(/d/[\w]+)">\s?<img`)
	//matched = re.FindAllStringSubmatch(result, -1)
	//if len(matched) < 1 || len(matched[0]) < 2{
	//	return "",  common.SubHDStep0HrefIsNull
	//}
	//return matched[0][1], nil
}
// Step1 获取影片的详情字幕列表
func (s Supplier) Step1(detailPageUrl string) ([]HdListItem, error) {
	detailPageUrl = model.AddBaseUrl(common.SubSubHDRootUrl, detailPageUrl)
	result, err := s.httpGet(detailPageUrl)
	if err != nil {
		return nil, err
	}
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
		if model.IsSubTypeWanted(insideSubType) == false {
			return true
		}
		// 下载的次数
		downCount, err := model.GetNumber2int(tr.Find(oneSubTrDownloadCountKeyword).Eq(1).Text())
		if err != nil {
			return true
		}

		listItem := HdListItem{}
		listItem.Url = downUrl
		listItem.BaseUrl = common.SubSubHDRootUrl
		listItem.Title = title
		listItem.DownCount = downCount

		if len(lists) >= s.topic {
			return false
		}

		lists = append(lists, listItem)

		return true
	})

	return lists, nil
}

// Step2Ex 下载字幕 过防水墙
func (s Supplier) Step2Ex(browser *rod.Browser, subDownloadPageUrl string) (*HdContent, error)  {
	subDownloadPageUrl = model.AddBaseUrl(common.SubSubHDRootUrl, subDownloadPageUrl)
	// TODO 需要提取出 rod 的超时时间和重试次数，注意，这里的超时时间，在调试的时候也算进去的，所以···
	page, err := model.NewPageNavigate(browser, subDownloadPageUrl, 300*time.Second, 5)
	if err != nil {
		return nil, err
	}
	err = page.WaitLoad()
	if err != nil {
		return nil, err
	}
	pageString, err := page.HTML()
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageString))
	if err != nil {
		return nil, err
	}
	// 是否有腾讯的防水墙
	hasWaterWall := true
	waterWall := doc.Find("#TencentCaptcha")
	if len(waterWall.Nodes) < 1 {
		hasWaterWall = false
	}
	// 是否有下载按钮
	hasDownBtn := true
	downBtn := doc.Find("#down")
	if len(downBtn.Nodes) < 1 {
		hasDownBtn = false
	}
	if hasWaterWall == false && hasDownBtn == false {
		// 都没有，则返回故障，无法下载
		return nil, common.SubHDStep2ExCannotFindDownloadBtn
	}
	// 下载字幕
	content, err := s.downloadSubFile(browser, page, hasWaterWall)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (s Supplier) downloadSubFile(browser *rod.Browser, page *rod.Page, hasWaterWall bool) (*HdContent, error) {
	var err error
	fileName := ""
	fileByte := []byte{0}
	err = rod.Try(func() {
		tmpDir := filepath.Join(os.TempDir(), "rod", "downloads")
		wait := browser.WaitDownload(tmpDir)
		getDownloadFile:= func() ([]byte, string, error) {
			info := wait()
			downloadPath := filepath.Join(tmpDir, info.GUID)
			defer func() { _ = os.Remove(downloadPath) }()
			b, err := ioutil.ReadFile(downloadPath)
			if err != nil {
				return nil, "", err
			}
			return b,info.SuggestedFilename, nil
		}

		// 点击下载按钮
		if hasWaterWall == true {
			page.MustElement("#TencentCaptcha").MustClick()
		} else {
			page.MustElement("#down").MustClick()
		}
		// 过墙
		if hasWaterWall == true {
			s.passWaterWall(page)
		}
		fileByte, fileName, err = getDownloadFile()
		if err != nil {
			panic(err)
		}
	})
	if err != nil {
		return nil, err
	}

	var hdContent HdContent
	hdContent.Filename = fileName
	hdContent.Ext = filepath.Ext(fileName)
	hdContent.Data = fileByte

	return &hdContent, nil
}

func (s Supplier) passWaterWall(page *rod.Page)  {
	//等待驗證碼窗體載入
	page.MustElement("#tcaptcha_iframe").MustWaitLoad()
	//進入到iframe
	iframe := page.MustElement("#tcaptcha_iframe").MustFrame()
	//等待拖動條加載, 延遲500秒檢測變化, 以確認加載完畢
	iframe.MustElement("#tcaptcha_drag_button").MustWaitStable()
	//等待缺口圖像載入
	slideBgEl := iframe.MustElement("#slideBg").MustWaitLoad()
	slideBgEl = slideBgEl.MustWaitStable()
	//取得帶缺口圖像
	shadowbg := slideBgEl.MustResource()
	// 取得原始圖像
	src := slideBgEl.MustProperty("src")
	fullbg, _, err := model.DownFile(strings.Replace(src.String(), "img_index=1", "img_index=0", 1))
	if err != nil {
		panic(err)
	}
	//取得img展示的真實尺寸
	shape, err := slideBgEl.Shape()
	if err != nil {
		panic(err)
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
				s.log.Debug("對比完畢, 偏移量:", distance)
				break search
			}
		}
	}
	//獲取拖動按鈕形狀
	dragBtnBox := iframe.MustElement("#tcaptcha_drag_thumb").MustShape().Box()
	//启用滑鼠功能
	mouse := page.Mouse
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

	if s.reqParam.DebugMode == true {
		//截圖保存
		nowProcessRoot, err := model.GetDebugFolder()
		if err == nil {
			page.MustScreenshot(path.Join(nowProcessRoot, "result.png"))
		} else {
			s.log.Errorln("model.GetDebugFolder", err)
		}
	}
}

func (s Supplier) httpGet(url string) (string, error) {
	s.reqParam.Referer = url
	httpClient := model.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().Get(url)
	if err != nil {
		return "", err
	}
	//搜索验证 点击继续搜索
	if strings.Contains(resp.String(), "搜索验证") {
		s.log.Debug("搜索验证 reload", url)
		return s.httpGet(url)
	}
	return resp.String(), nil
}

//httpPost  没用了，弃了
func (s Supplier) httpPost(url string, postData map[string]string, referer string) (string, error) {

	s.reqParam.Referer = referer
	httpClient := model.NewHttpClient(s.reqParam)
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