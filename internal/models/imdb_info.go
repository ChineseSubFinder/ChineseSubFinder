package models

type IMDBInfo struct {
	IMDBID        string         `gorm:"primaryKey"` // IMDB ID
	Name          string         // 视频名称
	Year          int            `gorm:"default:0"` // 发布的时间
	Description   string         // 描述
	Languages     StringList     `gorm:"type:varchar(255);not null"` // 语言
	AKA           StringList     `gorm:"type:varchar(255);not null"` // 又名 xx xxx
	VideoSubInfos []VideoSubInfo `gorm:"foreignKey:IMDBInfoID"`      // 视频对应的字幕，外键约束
}

func NewIMDBInfo(IMDBID string, name string, year int, description string, languages StringList, AKA StringList) *IMDBInfo {
	return &IMDBInfo{IMDBID: IMDBID, Name: name, Year: year, Description: description, Languages: languages, AKA: AKA, VideoSubInfos: make([]VideoSubInfo, 0)}
}
