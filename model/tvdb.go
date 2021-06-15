package model

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/go-resty/resty/v2"
)

type TVDB struct {
	reqParam common.ReqParam
	httpClient *resty.Client
}

// TODO 从 TVDB ID 查找对应的影片信息，得到中文名称，然后再去搜索字幕
// 半泽直树 zho 半澤直樹 zhtw
// 到了 list 列表搜索 zho ，第一个元素，然后找它的父级，获取 text 应该就能拿到中文名称了
func NewTVDB(_reqParam ...common.ReqParam) *TVDB {
	tv := TVDB{}
	if len(_reqParam) > 0 {
		tv.reqParam = _reqParam[0]
	}
	tv.httpClient = NewHttpClient(tv.reqParam)
	return &tv
}

func (t TVDB) SearchAndGetChineseName() error {
	resp, err := t.httpClient.R().
		SetQueryParams(map[string]string{
			"query": ,
		}).
		Get(common.TVDBSearchUrl)
	if err != nil {
		return err
	}
	println(resp)
	return nil
}
