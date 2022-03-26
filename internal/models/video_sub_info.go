package models

// VideoSubInfo 属于 IMDBInfo，IMDBInfoID 是外键，使用了 GORM 的 HasMany 关联
type VideoSubInfo struct {
	Feature     string `gorm:"primaryKey" json:"feature"` // 特征码，这个未必有，比如是蓝光格式，分散成多个视频文件的时候，暂定使用 shooter 的特征提前方式
	SubName     string `json:"sub_name"`                  // 字幕的文件名
	Season      int    `json:"season"`                    // 如果对应的是电影则可能是 0，没有
	Episode     int    `json:"episode"`                   // 如果对应的是电影则可能是 0，没有
	LanguageISO string `json:"language_iso"`              // 字幕的语言，目标语言，就算是双语，中英，也应该是中文。ISO_639-1_codes 标准，见 ISOLanguage.go 文件，这里无法区分简体繁体
	IsDouble    bool   `json:"is_double"`                 // 是否是双语，上面是主体语言，比如是中文，
	ChineseISO  string `json:"chinese_iso"`               // 中文语言编码变种，见 ISOLanguage.go 文件，这里区分简体、繁体等，如果语言是非中文则这里是空
	MyLanguage  string `json:"my_language"`               // 这个是本程序定义的语言类型，见 my_language.go 文件
	StoreFPath  string `json:"store_f_path"`              // 字幕存在出本地的哪里绝对路径上
	IMDBInfoID  string
}

func NewVideoSubInfo(feature string, subName string, languageISO string, isDouble bool, chineseISO string, myLanguage string, storeFPath string) *VideoSubInfo {
	return &VideoSubInfo{Feature: feature, SubName: subName, LanguageISO: languageISO, IsDouble: isDouble, ChineseISO: chineseISO, MyLanguage: myLanguage, StoreFPath: storeFPath}
}
