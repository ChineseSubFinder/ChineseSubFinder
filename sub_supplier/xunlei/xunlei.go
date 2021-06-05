package xunlei

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier"
	"github.com/go-resty/resty/v2"
	"math"
	"os"
)



type Supplier struct {

}

func NewSupplier() *Supplier {
	return &Supplier{}
}

func (s Supplier) GetSubListFromFile(filePath string, httpProxy string) ([]sub_supplier.SubInfo, error) {
	cid, err := s.getCid(filePath)
	var jsonList SublistSliceXunLei
	var outSubList []sub_supplier.SubInfo
	if len(cid) == 0 {
		return outSubList, common.CIdIsEmpty
	}
	httpClient := resty.New()
	httpClient.SetTimeout(common.HTMLTimeOut)
	if httpProxy != "" {
		httpClient.SetProxy(httpProxy)
	}
	httpClient.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent": "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	})
	resp, err := httpClient.R().Get(fmt.Sprintf(common.SubXunleiRootUrl, cid))
	if err != nil {
		return outSubList, err
	}
	// 解析
	err = json.Unmarshal([]byte(resp.String()), &jsonList)
	if err != nil {
		return outSubList, err
	}
	// 剔除空的
	for _, v := range jsonList.Sublist {
		if len(v.Scid) > 0 {
			outSubList = append(outSubList, *sub_supplier.NewSubInfo(v.Sname, v.Language, v.Rate, v.Surl, v.Svote, v.Roffset))
		}
	}
	return outSubList, nil
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