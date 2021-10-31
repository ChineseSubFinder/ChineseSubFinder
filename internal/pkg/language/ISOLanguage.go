package language

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"strings"
)

// ChineseISOString2Lang 将 中文描述：zh、zho、chi 转换为 types.MyLanguage
func ChineseISOString2Lang(chineseStr string) language.MyLanguage {

	switch chineseStr {
	case language.ISO_639_1_Chinese, language.ISO_639_2T_Chinese, language.ISO_639_2B_Chinese:
		return language.ChineseSimple
	default:
		return language.Unknown
	}
}

// ISOString2SupportLang 从 639-2/B 的语言缩写字符串转换为内部的 MyLanguage 类型，值支持
// https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
func ISOString2SupportLang(isoString string) language.MyLanguage {
	switch strings.ToLower(isoString) {
	case language.ISO_639_2B_Chinese:
		return language.ChineseSimple
	case language.ISO_639_2B_English:
		return language.English
	case language.ISO_639_2B_Japanese:
		return language.Japanese
	case language.ISO_639_2B_Korean:
		return language.Korean
	default:
		return language.Unknown
	}
}

// IsSupportISOString 是否是受支持的  639-2/B 语言，中、英、日、韩
func IsSupportISOString(isoString string) bool {
	switch strings.ToLower(isoString) {
	case language.ISO_639_2B_Chinese, language.ISO_639_2B_English, language.ISO_639_2B_Japanese, language.ISO_639_2B_Korean:
		return true
	default:
		return false
	}
}
