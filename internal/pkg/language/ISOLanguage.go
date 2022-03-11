package language

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"strings"
)

// ISOString2SupportLang 从语言缩写字符串转换为内部的 MyLanguage 类型
// 1. 支持 ISO 639-1、639-2/B、639-2/T、639-3
// 2. 支持中文的多种变种编码
// https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
func ISOString2SupportLang(isoString string) language.MyLanguage {

	lowerString := strings.ToLower(isoString)

	// 639-1
	switch lowerString {
	case language.ISO_639_1_Chinese:
		return language.ChineseSimple
	case language.ISO_639_1_English:
		return language.English
	case language.ISO_639_1_Korean:
		return language.Korean
	case language.ISO_639_1_Japanese:
		return language.Japanese
	}
	// 639-2/B
	switch lowerString {
	case language.ISO_639_2B_Chinese:
		return language.ChineseSimple
	case language.ISO_639_2B_English:
		return language.English
	case language.ISO_639_2B_Japanese:
		return language.Japanese
	case language.ISO_639_2B_Korean:
		return language.Korean
	}
	// 639-2/T
	switch lowerString {
	case language.ISO_639_2T_Chinese:
		return language.ChineseSimple
	case language.ISO_639_2T_English:
		return language.English
	case language.ISO_639_2T_Japanese:
		return language.Japanese
	case language.ISO_639_2T_Korean:
		return language.Korean
	}
	// 639-3
	switch lowerString {
	case language.ISO_639_3_Chinese:
		return language.ChineseSimple
	case language.ISO_639_3_English:
		return language.English
	case language.ISO_639_3_Korean:
		return language.Korean
	case language.ISO_639_3_Japanese:
		return language.Japanese
	}
	// 中文编码变种
	switch lowerString {
	case language.ChineseISO_Hans:
		return language.ChineseSimple
	case language.ChineseISO_Hant:
		return language.ChineseTraditional
	case language.ChineseISO_CN:
		return language.ChineseSimple
	case language.ChineseISO_TW:
		return language.ChineseTraditional
	case language.ChineseISO_SG,
		language.ChineseISO_MY:
		return language.ChineseSimple
	case language.ChineseISO_HK,
		language.ChineseISO_MO:
		return language.ChineseTraditional
	}

	return language.Unknown
}

// IsSupportISOString 是否是受支持的语言，中、英、日、韩
// 1. 支持 ISO 639-1、639-2/B、639-2/T、639-3
// 2. 支持中文的多种变种编码
func IsSupportISOString(isoString string) bool {

	lowerString := strings.ToLower(isoString)

	switch lowerString {
	case language.ISO_639_1_Chinese, language.ISO_639_1_English, language.ISO_639_1_Korean, language.ISO_639_1_Japanese:
		// 639-1
		return true
	}
	switch lowerString {
	case language.ISO_639_2B_Chinese, language.ISO_639_2B_English, language.ISO_639_2B_Japanese, language.ISO_639_2B_Korean:
		// 639-2/B
		return true
	}
	switch lowerString {
	case language.ISO_639_2T_Chinese, language.ISO_639_2T_English, language.ISO_639_2T_Japanese, language.ISO_639_2T_Korean:
		// 639-2/T
		return true
	}
	switch lowerString {
	case language.ISO_639_3_Chinese, language.ISO_639_3_English, language.ISO_639_3_Korean, language.ISO_639_3_Japanese:
		// 639-3
		return true
	}
	switch lowerString {
	case language.ChineseISO_Hans,
		language.ChineseISO_Hant,
		language.ChineseISO_CN,
		language.ChineseISO_TW,
		language.ChineseISO_SG,
		language.ChineseISO_MY,
		language.ChineseISO_HK,
		language.ChineseISO_MO:
		// 中文编码变种
		return true
	}

	return false
}

// IsSupportISOChineseString 是否是受支持的语言，中
// 1. 支持 ISO 639-1、639-2/B、639-2/T、639-3
// 2. 支持中文的多种变种编码
func IsSupportISOChineseString(isoString string) bool {

	lowerString := strings.ToLower(isoString)

	switch lowerString {
	case language.ISO_639_1_Chinese:
		// 639-1
		return true
	}
	switch lowerString {
	case language.ISO_639_2B_Chinese:
		// 639-2/B
		return true
	}
	switch lowerString {
	case language.ISO_639_2T_Chinese:
		// 639-2/T
		return true
	}
	switch lowerString {
	case language.ISO_639_3_Chinese:
		// 639-3
		return true
	}
	switch lowerString {
	case language.ChineseISO_Hans,
		language.ChineseISO_Hant,
		language.ChineseISO_CN,
		language.ChineseISO_TW,
		language.ChineseISO_SG,
		language.ChineseISO_MY,
		language.ChineseISO_HK,
		language.ChineseISO_MO:
		// 中文编码变种
		return true
	}

	return false
}

// ISOSupportRegexRule 获取 ISO 匹配的 regex 表达式
func ISOSupportRegexRule() string {

	if isoISORegString != "" {
		return isoISORegString
	}

	isoISORegString = language.RegISORuleFront

	isoISORegString += addISORegSubLangString(language.ChineseISO_Hans)
	isoISORegString += addISORegSubLangString(language.ChineseISO_Hant)
	isoISORegString += addISORegSubLangString(language.ChineseISO_CN)
	isoISORegString += addISORegSubLangString(language.ChineseISO_TW)
	isoISORegString += addISORegSubLangString(language.ChineseISO_SG)
	isoISORegString += addISORegSubLangString(language.ChineseISO_MY)
	isoISORegString += addISORegSubLangString(language.ChineseISO_HK)
	isoISORegString += addISORegSubLangString(language.ChineseISO_MO)

	isoISORegString += addISORegSubLangString(language.ISO_639_1_Chinese)
	isoISORegString += addISORegSubLangString(language.ISO_639_1_English)
	isoISORegString += addISORegSubLangString(language.ISO_639_1_Korean)
	isoISORegString += addISORegSubLangString(language.ISO_639_1_Japanese)

	isoISORegString += addISORegSubLangString(language.ISO_639_2T_Chinese)
	isoISORegString += addISORegSubLangString(language.ISO_639_2T_English)
	isoISORegString += addISORegSubLangString(language.ISO_639_2T_Korean)
	isoISORegString += addISORegSubLangString(language.ISO_639_2T_Japanese)

	isoISORegString += addISORegSubLangString(language.ISO_639_2B_Chinese)

	isoISORegString += language.RegISORuleEnd

	return isoISORegString
}

func addISORegSubLangString(inLang string) string {
	return fmt.Sprintf(`\b%s\b|`, inLang)
}

var isoISORegString = ""
