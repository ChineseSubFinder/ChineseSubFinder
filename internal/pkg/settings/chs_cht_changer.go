package settings

type ChsChtChanger struct {
	Enable                 bool `json:"enable"`
	DesChineseLanguageType int  `json:"des_chinese_language_type"` // 默认 0 是 简体 ，1 是 繁体
}
