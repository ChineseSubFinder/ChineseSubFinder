package models

type MediaInfo struct {
	TmdbId           string `gorm:"type:varchar(20);primarykey"`
	ImdbId           string `gorm:"type:varchar(20)"`
	OriginalTitle    string `gorm:"type:varchar(100)"`
	OriginalLanguage string `gorm:"type:varchar(100)"` // 视频的原始语言  en zh
	TitleEn          string `gorm:"type:varchar(100)"` // 英文标题
	TitleCn          string `gorm:"type:varchar(100)"` // 中文的标题
	Year             string `gorm:"type:varchar(20)"`  // 播出的时间，如果是连续剧是第一次播出的时间 2019-01-01  2022-01-01
}
