package xunlei

import (
	"crypto/sha1"
	"fmt"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"path/filepath"
)

type Supplier struct {
	reqParam types.ReqParam
	log      *logrus.Logger
	topic    int
}

func NewSupplier(_reqParam ...types.ReqParam) *Supplier {

	sup := Supplier{}
	sup.log = pkg.GetLogger()
	sup.topic = common2.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		sup.reqParam = _reqParam[0]
		if sup.reqParam.Topic > 0 && sup.reqParam.Topic != sup.topic {
			sup.topic = sup.reqParam.Topic
		}
	}
	return &sup
}

func (s Supplier) GetSupplierName() string {
	return common2.SubSiteXunLei
}

func (s Supplier) GetReqParam() types.ReqParam {
	return s.reqParam
}

func (s Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error){
	return s.getSubListFromFile(filePath)
}

func (s Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error){
	return s.downloadSub4Series(seriesInfo)
}

func (s Supplier) getSubListFromFile(filePath string) ([]supplier.SubInfo, error) {

	cid, err := s.getCid(filePath)
	var jsonList SublistSliceXunLei
	var tmpXunLeiSubListChinese = make([]SublistXunLei, 0)
	var outSubList []supplier.SubInfo
	if len(cid) == 0 {
		return outSubList, common2.XunLeiCIdIsEmpty
	}
	httpClient := pkg.NewHttpClient(s.reqParam)
	_, err = httpClient.R().
		SetResult(&jsonList).
		Get(fmt.Sprintf(common2.SubXunLeiRootUrl, cid))
	if err != nil {
		return outSubList, err
	}
	// 剔除空的
	for _, v := range jsonList.Sublist {
		if len(v.Scid) > 0 && v.Scid != "" {
			// 符合中文语言的先加入列表
			tmpLang := pkg.LangConverter(v.Language)
			if pkg.HasChineseLang(tmpLang) == true && pkg.IsSubTypeWanted(v.Sname) == true {
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
				tmpLang := pkg.LangConverter(v.Language)
				if pkg.HasChineseLang(tmpLang) == false {
					tmpXunLeiSubListChinese = append(tmpXunLeiSubListChinese, v)
				}
			}
		}
	}
	// 再开始下载字幕
	for i, v := range tmpXunLeiSubListChinese {
		tmpLang := pkg.LangConverter(v.Language)
		data, filename, err := pkg.DownFile(v.Surl)
		if err != nil {
			s.log.Error(err)
			continue
		}
		ext := ""
		if filename == "" {
			ext = filepath.Ext(v.Surl)
		} else {
			ext = filepath.Ext(filename)
		}

		outSubList = append(outSubList, *supplier.NewSubInfo(s.GetSupplierName(), int64(i), v.Sname, tmpLang, v.Surl, v.Svote, v.Roffset, ext, data))
	}


	return outSubList, nil
}

func (s Supplier) getSubListFromKeyword(keyword string) ([]supplier.SubInfo, error) {
	panic("not implemented")
}

//getCid 获取指定文件的唯一 cid
func (s Supplier) getCid(filePath string) (string, error) {
	hash := ""
	sha1Ctx := sha1.New()

	fp, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fp.Close()
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

func (s Supplier) downloadSub4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	var allSupplierSubInfo = make([]supplier.SubInfo, 0)
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {
		one, err := s.getSubListFromFile(episodeInfo.FileFullPath)
		if err != nil {
			return nil, err
		}
		// 需要赋值给字幕结构
		for i, _ := range one {
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

const LangUnknow = "未知语言"