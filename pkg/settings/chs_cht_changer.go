package settings

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
)

type ChsChtChanger struct {
	Enable                 bool `json:"enable"`
	DesChineseLanguageType int  `json:"des_chinese_language_type"` // 默认 0 是 简体 ，1 是 繁体
}

func (c ChsChtChanger) GetDesChineseLanguageTypeString() string {
	if c.DesChineseLanguageType == 0 {
		return language.MatchLangChs
	} else {
		return language.MatchLangCht
	}
}
