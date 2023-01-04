package models

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/longbridgeapp/opencc"
)

type IMDBInfo struct {
	IMDBID        string         `gorm:"primaryKey" json:"imdb_id"  binding:"required"`                   // IMDB ID
	Name          string         `json:"name" binding:"required"`                                         // 视频名称
	Year          int            `gorm:"default:0" json:"year"  binding:"required"`                       // 发布的时间
	Description   string         `json:"description"  binding:"required"`                                 // 描述
	Languages     StringList     `gorm:"type:varchar(255);not null" json:"languages"  binding:"required"` // 语言
	AKA           StringList     `gorm:"type:varchar(255);not null" json:"AKA"  binding:"required"`       // 又名 xx xxx
	RootDirPath   string         `json:"root_dir_path"`                                                   // 这个电影或者连续剧（不是季的文件夹，而是这个连续剧的目录）路径
	IsMovie       bool           `json:"is_movie"`                                                        // 不是电影就是连续剧
	TmdbId        string         `gorm:"type:varchar(20)"`                                                // TMDB ID 也是 MediaInfo 的主键
	VideoSubInfos []VideoSubInfo `gorm:"foreignKey:IMDBInfoID"`                                           // 视频对应的字幕，外键约束
}

func NewIMDBInfo(IMDBID string, name string, year int, description string, languages StringList, AKA StringList) *IMDBInfo {
	return &IMDBInfo{IMDBID: IMDBID, Name: name, Year: year, Description: description, Languages: languages, AKA: AKA, VideoSubInfos: make([]VideoSubInfo, 0)}
}

func (i *IMDBInfo) GetChineseNameFromAKA() string {

	if len(i.AKA) == 0 {
		return ""
	}
	chsName := ""
	chtName := ""
	for _, akaWord := range i.AKA {
		// 0 不是简体和繁体，1 是简体，2 是繁体
		if language.WhichChineseType(akaWord) == 1 {
			chsName = akaWord
			break
		} else if language.WhichChineseType(akaWord) == 2 {
			chtName = akaWord
			break
		}
	}
	// 如果简体找到了，那么就返回
	if chsName != "" {
		return chsName
	}
	// 如果繁体找到了，那么进行一次简体转换
	if chtName != "" {

		t2s, err := opencc.New("t2s")
		if err != nil {
			return ""
		}
		// 繁体转简体
		newChs, err := t2s.Convert(chtName)
		if err != nil {
			return ""
		}

		return newChs
	}

	return ""
}
