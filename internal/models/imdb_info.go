package models

type IMDBInfo struct {
	IMDBID                  string                   `gorm:"primaryKey" json:"imdb_id"  binding:"required"`                   // IMDB ID
	Name                    string                   `json:"name" binding:"required"`                                         // 视频名称
	Year                    int                      `gorm:"default:0" json:"year"  binding:"required"`                       // 发布的时间
	Description             string                   `json:"description"  binding:"required"`                                 // 描述
	Languages               StringList               `gorm:"type:varchar(255);not null" json:"languages"  binding:"required"` // 语言
	AKA                     StringList               `gorm:"type:varchar(255);not null" json:"AKA"  binding:"required"`       // 又名 xx xxx
	VideoSubInfos           []VideoSubInfo           `gorm:"foreignKey:IMDBInfoID"`                                           // 视频对应的字幕，外键约束
	MovieOrSeriesLocalInfos []MovieOrSeriesLocalInfo `gorm:"foreignKey:IMDBInfoID"`                                           // 视频对应的本地信息，外键约束
}

func NewIMDBInfo(IMDBID string, name string, year int, description string, languages StringList, AKA StringList) *IMDBInfo {
	return &IMDBInfo{IMDBID: IMDBID, Name: name, Year: year, Description: description, Languages: languages, AKA: AKA, VideoSubInfos: make([]VideoSubInfo, 0)}
}
