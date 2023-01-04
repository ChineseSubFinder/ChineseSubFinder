package xunlei

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/series"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/notify_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/sirupsen/logrus"
)

type Supplier struct {
	log            *logrus.Logger
	fileDownloader *file_downloader.FileDownloader
	topic          int
	isAlive        bool
}

func NewSupplier(fileDownloader *file_downloader.FileDownloader) *Supplier {

	sup := Supplier{}
	sup.log = fileDownloader.Log
	sup.fileDownloader = fileDownloader
	sup.topic = common.DownloadSubsPerSite
	sup.isAlive = true // 默认是可以使用的，如果 check 后，再调整状态

	if settings.Get().AdvancedSettings.Topic > 0 && settings.Get().AdvancedSettings.Topic != sup.topic {
		sup.topic = settings.Get().AdvancedSettings.Topic
	}

	return &sup
}

func (s *Supplier) CheckAlive() (bool, int64) {

	// 计算当前时间
	startT := time.Now()
	jsonList, err := s.getSubInfos(checkFileName, checkCID)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Error", err)
		s.isAlive = false
		return false, 0
	}

	if len(jsonList.Sublist) < 1 {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Sublist < 1")
		s.isAlive = false
		return false, 0
	}

	s.isAlive = true
	return true, time.Since(startT).Milliseconds()
}

func (s *Supplier) IsAlive() bool {
	return s.isAlive
}

func (s *Supplier) OverDailyDownloadLimit() bool {

	if settings.Get().AdvancedSettings.SuppliersSettings.Xunlei.DailyDownloadLimit == 0 {
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
	return common.SubSiteXunLei
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {
	return s.getSubListFromFile(filePath)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) getSubListFromFile(filePath string) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), filePath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), filePath, "Start...")

	if pkg.IsFile(filePath) == false {
		// 这里传入的可能是蓝光结构的伪造存在的视频文件，需要检查一次这个文件是否存在
		bok, _, _ := decode.IsFakeBDMVWorked(filePath)
		if bok == false {

			nowError := errors.New(fmt.Sprintf("%s %s %s",
				s.GetSupplierName(),
				filePath,
				"not exist, and it`s not a Blue ray Video FakeFileName"))

			s.log.Errorln(nowError)

			return nil, nowError
		}
	}

	cid, err := s.getCid(filePath)
	var jsonList SublistSliceXunLei
	var tmpXunLeiSubListChinese = make([]SublistXunLei, 0)
	var outSubList []supplier.SubInfo
	if len(cid) == 0 {
		return nil, common.XunLeiCIdIsEmpty
	}

	jsonList, err = s.getSubInfos(filePath, cid)
	if err != nil {
		return nil, err
	}

	// 剔除空的
	for _, v := range jsonList.Sublist {
		if len(v.Scid) > 0 && v.Scid != "" {
			// 符合中文语言的先加入列表
			tmpLang := language.LangConverter4Sub_Supplier(v.Language)
			if language.HasChineseLang(tmpLang) == true && sub_parser_hub.IsSubTypeWanted(v.Sname) == true {
				tmpXunLeiSubListChinese = append(tmpXunLeiSubListChinese, v)
			}
		}
	}
	// TODO 这里需要考虑，可以设置为高级选项，不够就用 unknow 来补充
	// 如果不够，再补 unknow
	if len(tmpXunLeiSubListChinese) < s.topic {
		for _, v := range jsonList.Sublist {
			if len(tmpXunLeiSubListChinese) >= s.topic {
				break
			}
			if len(v.Scid) > 0 && v.Scid != "" {
				tmpLang := language.LangConverter4Sub_Supplier(v.Language)
				if language.HasChineseLang(tmpLang) == false {
					tmpXunLeiSubListChinese = append(tmpXunLeiSubListChinese, v)
				}
			}
		}
	}

	videoFileName := filepath.Base(filePath)
	// 再开始下载字幕
	for i, v := range tmpXunLeiSubListChinese {

		// 解析 xunlei 列表中的这个字幕类型转换到内部的字幕类型
		//tmpLang := language.LangConverter4Sub_Supplier(v.Language)
		subInfo, err := s.fileDownloader.Get(s.GetSupplierName(), int64(i), videoFileName, v.Surl, v.Svote, v.Roffset)
		if err != nil {
			s.log.Error("FileDownloader.Get", err)
			continue
		}

		outSubList = append(outSubList, *subInfo)
	}

	return outSubList, nil
}

func (s *Supplier) getSubInfos(filePath, cid string) (SublistSliceXunLei, error) {
	var jsonList SublistSliceXunLei

	httpClient, err := pkg.NewHttpClient()
	if err != nil {
		return jsonList, err
	}
	resp, err := httpClient.R().
		SetResult(&jsonList).
		Get(fmt.Sprintf(settings.Get().AdvancedSettings.SuppliersSettings.Xunlei.RootUrl, cid))
	if err != nil {
		if resp != nil {
			s.log.Errorln(s.GetSupplierName(), "NewHttpClient:", filePath, err.Error())
			notify_center.Notify.Add(s.GetSupplierName()+" NewHttpClient", fmt.Sprintf("filePath: %s, resp: %s, error: %s", filePath, resp.String(), err.Error()))
		}
		return jsonList, err
	}

	return jsonList, nil
}

//getCid 获取指定文件的唯一 cid
func (s *Supplier) getCid(filePath string) (string, error) {
	hash := ""
	sha1Ctx := sha1.New()

	fp, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = fp.Close()
	}()
	stat, err := fp.Stat()
	if err != nil {
		return "", err
	}
	fileLength := stat.Size()
	if fileLength < 0xF000 {
		return "", err
	}
	bufferSize := int64(0x5000)
	positions := []int64{0, int64(math.Floor(float64(fileLength) / 3)), fileLength - bufferSize}
	for _, position := range positions {
		var buffer = make([]byte, bufferSize)
		_, err = fp.Seek(position, 0)
		if err != nil {
			return "", err
		}
		_, err = fp.Read(buffer)
		if err != nil {
			return "", err
		}
		sha1Ctx.Write(buffer)
	}

	hash = fmt.Sprintf("%X", sha1Ctx.Sum(nil))
	return hash, nil
}

func (s *Supplier) downloadSub4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	var allSupplierSubInfo = make([]supplier.SubInfo, 0)
	index := 0
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		index++
		one, err := s.getSubListFromFile(episodeInfo.FileFullPath)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "getSubListFromFile", episodeInfo.Season, episodeInfo.Episode,
				episodeInfo.FileFullPath)
			continue
		}
		if one == nil {
			// 没有搜索到字幕
			s.log.Infoln(s.GetSupplierName(), "Not Find Sub can be download",
				episodeInfo.Title, episodeInfo.Season, episodeInfo.Episode)
			continue
		}
		// 需要赋值给字幕结构
		for i := range one {
			one[i].Season = episodeInfo.Season
			one[i].Episode = episodeInfo.Episode
		}
		allSupplierSubInfo = append(allSupplierSubInfo, one...)
	}
	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return allSupplierSubInfo, nil
}

type SublistXunLei struct {
	Scid     string `json:"scid"`
	Sname    string `json:"sname"`
	Language string `json:"language"`
	Rate     string `json:"rate"`
	Surl     string `json:"surl"`
	Svote    int64  `json:"svote"`
	Roffset  int64  `json:"roffset"`
}

type SublistSliceXunLei struct {
	Sublist []SublistXunLei
}

const (
	checkFileName = "CheckFileName"
	checkCID      = "FB4E2AFF106112136DFC5ACC7339EB29D1EC0CF8"
)
