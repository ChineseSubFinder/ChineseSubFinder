package sub_supplier

import "github.com/allanpk716/ChineseSubFinder/common"

type SubInfo struct {
	Name 		string `json:"name"`		// 字幕的名称，这个比较随意，优先是影片的名称，然后才是从网上下载字幕的对应名称
	Language 	common.Language `json:"language"`	// 字幕的语言
	FileUrl     string `json:"file-url"`	// 字幕文件下载的路径
	Vote    	int64  `json:"vote"`		// 投票
	Offset  	int64  `json:"offset"`		// 字幕的偏移
	Ext			string `json:"ext"`			// 字幕文件的后缀名带点，有可能是直接能用的字幕文件，也可能是压缩包
	Data		[]byte	`json:"data"`		// 字幕文件的二进制数据
}

func NewSubInfo(name string, language common.Language, fileUrl string, vote int64, offset int64, ext string, data []byte) *SubInfo {
	return &SubInfo{Name: name, Language: language, FileUrl: fileUrl, Vote: vote, Offset: offset, Ext: ext, Data: data}
}

