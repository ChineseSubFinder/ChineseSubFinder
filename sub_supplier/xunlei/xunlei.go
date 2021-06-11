package xunlei

import (
	"crypto/sha1"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"path/filepath"
)

type Supplier struct {
	reqParam common.ReqParam
	log *logrus.Logger
	topic int
}

func NewSupplier(_reqParam ... common.ReqParam) *Supplier {

	sup := Supplier{}
	sup.log = common.GetLogger()
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
	return common.SubSiteXunLei
}

func (s Supplier) GetSubListFromFile(filePath string) ([]common.SupplierSubInfo, error) {

	cid, err := s.getCid(filePath)
	var jsonList SublistSliceXunLei
	var tmpXunLeiSubListChinese = make([]SublistXunLei, 0)
	var outSubList []common.SupplierSubInfo
	if len(cid) == 0 {
		return outSubList, common.XunLeiCIdIsEmpty
	}
	httpClient := common.NewHttpClient(s.reqParam)
	_, err = httpClient.R().
		SetResult(&jsonList).
		Get(fmt.Sprintf(common.SubXunLeiRootUrl, cid))
	if err != nil {
		return outSubList, err
	}
	// 剔除空的
	for _, v := range jsonList.Sublist {
		if len(v.Scid) > 0 {
			// 符合中文语言的先加入列表
			tmpLang := common.LangConverter(v.Language)
			if common.HasChineseLang(tmpLang) == true && common.IsSubTypeWanted(v.Sname) == true {
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
			tmpLang := common.LangConverter(v.Language)
			if common.HasChineseLang(tmpLang) == false {
				tmpXunLeiSubListChinese = append(tmpXunLeiSubListChinese, v)
			}
		}
	}
	// 再开始下载字幕
	for i, v := range tmpXunLeiSubListChinese {
		tmpLang := common.LangConverter(v.Language)
		data, filename, err := common.DownFile(v.Surl)
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

		outSubList = append(outSubList, *common.NewSupplierSubInfo(s.GetSupplierName(), int64(i), v.Sname, tmpLang, v.Surl, v.Svote, v.Roffset, ext, data))
	}


	return outSubList, nil
}

func (s Supplier) GetSubListFromKeyword(keyword string) ([]common.SupplierSubInfo, error) {
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