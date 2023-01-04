package series

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
	"time"
)

/*
	这里只需要分为三层结构，因为有 sonarr 和 TMM 整理过
	所以命名很标注，使用 GetVideoInfoFromFileName 读取 SxxExx 问题不大
*/
type SeriesInfo struct {
	ImdbId           string
	Name             string
	Year             int
	ReleaseDate      string
	EpList           []EpisodeInfo
	DirPath          string
	SeasonDict       map[int]int
	NeedDlSeasonDict map[int]int
	NeedDlEpsKeyList map[string]EpisodeInfo // SxEx
}

type EpisodeInfo struct {
	MediaServerInsideVideoID string
	Title                    string
	Season                   int
	Episode                  int
	SubAlreadyDownloadedList []SubInfo // 已经下载在当前视频目录下的字幕列表
	Dir                      string    // 这里需要记录字幕的位置，因为需要在同级目录匹配相应的字幕才行
	FileFullPath             string    // 视频文件的全路径
	ModifyTime               time.Time // 创建时间
	AiredTime                string    // 播出的时间
}

type SubInfo struct {
	Title        string
	Season       int
	Episode      int
	Language     language.MyLanguage
	Dir          string // 这里需要记录字幕的位置，因为需要在同级目录匹配相应的视频才行
	FileFullPath string // 字幕文件的全路径
}
