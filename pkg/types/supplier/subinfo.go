package supplier

import (
	"crypto/sha256"
	"fmt"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
)

type SubInfo struct {
	FromWhere     string              `json:"from_where"`     // 从哪个网站下载来的
	TopN          int64               `json:"top_n"`          // 是 Top 几？
	Name          string              `json:"name"`           // 字幕的名称，这个比较随意，优先是影片的名称，然后才是从网上下载字幕的对应名称
	Language      language.MyLanguage `json:"language"`       // 字幕的语言
	FileUrl       string              `json:"file-url"`       // 字幕文件下载的路径
	Score         int64               `json:"score"`          // TODO 字幕的评分，需要有一个独立的评价体系。首先是每个网站自己的评价排序，然后再到统一的评分体系
	Offset        int64               `json:"offset"`         // 字幕的偏移
	Ext           string              `json:"ext"`            // 字幕文件的后缀名带点，有可能是直接能用的字幕文件，也可能是压缩包
	Data          []byte              `json:"data"`           // 字幕文件的二进制数据
	Season        int                 `json:"season"`         // 第几季，默认-1
	Episode       int                 `json:"episode"`        // 第几集，默认-1
	IsFullSeason  bool                `json:"is_full_season"` // 是否是全季的字幕
	fileUrlSha256 string              // 字幕文件的 FileUrl sha256 值
}

func NewSubInfo(fromWhere string, topN int64, name string, language language.MyLanguage, fileUrl string,
	score int64, offset int64, ext string, data []byte) *SubInfo {

	s := SubInfo{FromWhere: fromWhere, TopN: topN, Name: name, Language: language, FileUrl: fileUrl,
		Score: score, Offset: offset, Ext: ext, Data: data}

	s.Season = -1
	s.Episode = -1

	return &s
}

// GetUID 通过 FileUrl 获取字幕的唯一标识
func (s *SubInfo) GetUID() string {

	if s.fileUrlSha256 == "" {
		if s.FileUrl == "" {
			return ""
		}
		s.fileUrlSha256 = fmt.Sprintf("%x", sha256.Sum256([]byte(s.FileUrl)))
		return s.fileUrlSha256
	} else {
		return s.fileUrlSha256
	}
}

// SetFileUrlSha256 为了 ASSRT 这种下载连接是临时情况所准备的
func (s *SubInfo) SetFileUrlSha256(fileUrlSha256 string) {
	s.fileUrlSha256 = fileUrlSha256
}
