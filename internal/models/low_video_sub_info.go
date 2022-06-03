package models

type LowVideoSubInfo struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	IMDBID       string `json:"imdb_id"`
	TMDBID       string `json:"tmdb_id"`
	Feature      string `json:"feature"  binding:"required"`       // 特征码，这个未必有，比如是蓝光格式，分散成多个视频文件的时候，暂定使用本程序的特征提前方式
	SubName      string `json:"sub_name" binding:"required"`       // 字幕的文件名
	Season       int    `json:"season"`                            // 如果对应的是电影则可能是 0，没有
	Episode      int    `json:"episode"`                           // 如果对应的是电影则可能是 0，没有
	LanguageISO  string `json:"language_iso" binding:"required"`   // 字幕的语言，目标语言，就算是双语，中英，也应该是中文。ISO_639-1_codes 标准，见 ISOLanguage.go 文件，这里无法区分简体繁体
	IsDouble     bool   `json:"is_double" binding:"required"`      // 是否是双语，上面是主体语言，比如是中文，
	ChineseISO   string `json:"chinese_iso" binding:"required"`    // 中文语言编码变种，见 ISOLanguage.go 文件，这里区分简体、繁体等，如果语言是非中文则这里是空
	MyLanguage   string `json:"my_language" binding:"required"`    // 这个是本程序定义的语言类型，见 my_language.go 文件
	StoreRPath   string `json:"store_r_path"`                      // 字幕存在出本地的哪里相对路径上，cache/CSF-ShareSubCache
	ExtraPreName string `json:"extra_pre_name" binding:"required"` // 字幕额外的命名信息，指 Emby 字幕命名格式(简英,subhd)，的 subhd
	SHA256       string `json:"sha_256" binding:"required"`        // 当前文件的 sha256 的值
	IsSend       bool   `json:"is_send"`                           // 是否已经发送
}

func NewLowVideoSubInfo(imdbID, tmdbID, feature string, subName string, languageISO string, isDouble bool, chineseISO string, myLanguage string, storeFPath string, extraPreName string, sha256String string) *LowVideoSubInfo {
	return &LowVideoSubInfo{IMDBID: imdbID, TMDBID: tmdbID, Feature: feature, SubName: subName, LanguageISO: languageISO, IsDouble: isDouble, ChineseISO: chineseISO, MyLanguage: myLanguage, StoreRPath: storeFPath, ExtraPreName: extraPreName, SHA256: sha256String}
}
