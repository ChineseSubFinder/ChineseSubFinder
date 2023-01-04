package language

import (
	"fmt"
	"strings"

	language2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
)

// ISOString2SupportLang 从语言缩写字符串转换为内部的 MyLanguage 类型
// 1. 支持 ISO 639-1、639-2/B、639-2/T、639-3
// 2. 支持中文的多种变种编码
// https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
func ISOString2SupportLang(isoString string) language2.MyLanguage {

	lowerString := strings.ToLower(isoString)

	// 639-1
	switch lowerString {
	case language2.ISO_639_1_Chinese:
		return language2.ChineseSimple
	case language2.ISO_639_1_English:
		return language2.English
	case language2.ISO_639_1_Korean:
		return language2.Korean
	case language2.ISO_639_1_Japanese:
		return language2.Japanese
	}
	// 639-2/B
	switch lowerString {
	case language2.ISO_639_2B_Chinese:
		return language2.ChineseSimple
	case language2.ISO_639_2B_English:
		return language2.English
	case language2.ISO_639_2B_Japanese:
		return language2.Japanese
	case language2.ISO_639_2B_Korean:
		return language2.Korean
	}
	// 639-2/T
	switch lowerString {
	case language2.ISO_639_2T_Chinese:
		return language2.ChineseSimple
	case language2.ISO_639_2T_English:
		return language2.English
	case language2.ISO_639_2T_Japanese:
		return language2.Japanese
	case language2.ISO_639_2T_Korean:
		return language2.Korean
	}
	// 639-3
	switch lowerString {
	case language2.ISO_639_3_Chinese:
		return language2.ChineseSimple
	case language2.ISO_639_3_English:
		return language2.English
	case language2.ISO_639_3_Korean:
		return language2.Korean
	case language2.ISO_639_3_Japanese:
		return language2.Japanese
	}
	// 中文编码变种
	switch lowerString {
	case language2.ChineseISO_Hans:
		return language2.ChineseSimple
	case language2.ChineseISO_Hant:
		return language2.ChineseTraditional
	case language2.ChineseISO_CN:
		return language2.ChineseSimple
	case language2.ChineseISO_TW:
		return language2.ChineseTraditional
	case language2.ChineseISO_SG,
		language2.ChineseISO_MY:
		return language2.ChineseSimple
	case language2.ChineseISO_HK,
		language2.ChineseISO_MO:
		return language2.ChineseTraditional
	}

	return language2.Unknown
}

// IsSupportISOString 是否是受支持的语言，中、英、日、韩
// 1. 支持 ISO 639-1、639-2/B、639-2/T、639-3
// 2. 支持中文的多种变种编码
func IsSupportISOString(isoString string) bool {

	lowerString := strings.ToLower(isoString)

	switch lowerString {
	case language2.ISO_639_1_Chinese, language2.ISO_639_1_English, language2.ISO_639_1_Korean, language2.ISO_639_1_Japanese:
		// 639-1
		return true
	}
	switch lowerString {
	case language2.ISO_639_2B_Chinese, language2.ISO_639_2B_English, language2.ISO_639_2B_Japanese, language2.ISO_639_2B_Korean:
		// 639-2/B
		return true
	}
	switch lowerString {
	case language2.ISO_639_2T_Chinese, language2.ISO_639_2T_English, language2.ISO_639_2T_Japanese, language2.ISO_639_2T_Korean:
		// 639-2/T
		return true
	}
	switch lowerString {
	case language2.ISO_639_3_Chinese, language2.ISO_639_3_English, language2.ISO_639_3_Korean, language2.ISO_639_3_Japanese:
		// 639-3
		return true
	}
	switch lowerString {
	case language2.ChineseISO_Hans,
		language2.ChineseISO_Hant,
		language2.ChineseISO_CN,
		language2.ChineseISO_TW,
		language2.ChineseISO_SG,
		language2.ChineseISO_MY,
		language2.ChineseISO_HK,
		language2.ChineseISO_MO:
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
	case language2.ISO_639_1_Chinese:
		// 639-1
		return true
	}
	switch lowerString {
	case language2.ISO_639_2B_Chinese:
		// 639-2/B
		return true
	}
	switch lowerString {
	case language2.ISO_639_2T_Chinese:
		// 639-2/T
		return true
	}
	switch lowerString {
	case language2.ISO_639_3_Chinese:
		// 639-3
		return true
	}
	switch lowerString {
	case language2.ChineseISO_Hans,
		language2.ChineseISO_Hant,
		language2.ChineseISO_CN,
		language2.ChineseISO_TW,
		language2.ChineseISO_SG,
		language2.ChineseISO_MY,
		language2.ChineseISO_HK,
		language2.ChineseISO_MO:
		// 中文编码变种
		return true
	}

	return false
}

// MyLang2ISO_639_1_String 内置的语言转换到 ISO_639-1_codes 标准
func MyLang2ISO_639_1_String(myLanguage language2.MyLanguage) string {

	switch myLanguage {
	case language2.ChineseSimple,
		language2.ChineseTraditional,
		language2.ChineseSimpleEnglish,
		language2.ChineseTraditionalEnglish,
		language2.ChineseSimpleJapanese,
		language2.ChineseTraditionalJapanese,
		language2.ChineseSimpleKorean,
		language2.ChineseTraditionalKorean:
		return language2.ISO_639_1_Chinese
	case language2.English:
		return language2.ISO_639_1_English
	case language2.Japanese:
		return language2.ISO_639_1_Japanese
	case language2.Korean:
		return language2.ISO_639_1_Korean
	default:
		return language2.MathLangChnUnknown
	}
}

// MyLang2ChineseISO 中文语言编码变种，见 ISOLanguage.go 文件，这里区分简体、繁体等，如果语言是非中文则这里是空
func MyLang2ChineseISO(myLanguage language2.MyLanguage) string {
	switch myLanguage {
	case language2.ChineseSimple,
		language2.ChineseSimpleEnglish,
		language2.ChineseSimpleJapanese,
		language2.ChineseSimpleKorean:
		return language2.ChineseISO_Hans

	case language2.ChineseTraditional,
		language2.ChineseTraditionalEnglish,
		language2.ChineseTraditionalJapanese,
		language2.ChineseTraditionalKorean:
		return language2.ChineseISO_Hant

	case language2.English, language2.Japanese, language2.Korean:
		return ""
	default:
		return ""
	}
}

// ISOSupportRegexRule 获取 ISO 匹配的 regex 表达式
func ISOSupportRegexRule() string {

	if isoISORegString != "" {
		return isoISORegString
	}

	isoISORegString = language2.RegISORuleFront

	isoISORegString += addISORegSubLangString(language2.ChineseISO_Hans)
	isoISORegString += addISORegSubLangString(language2.ChineseISO_Hant)
	isoISORegString += addISORegSubLangString(language2.ChineseISO_CN)
	isoISORegString += addISORegSubLangString(language2.ChineseISO_TW)
	isoISORegString += addISORegSubLangString(language2.ChineseISO_SG)
	isoISORegString += addISORegSubLangString(language2.ChineseISO_MY)
	isoISORegString += addISORegSubLangString(language2.ChineseISO_HK)
	isoISORegString += addISORegSubLangString(language2.ChineseISO_MO)

	isoISORegString += addISORegSubLangString(language2.ISO_639_1_Chinese)
	isoISORegString += addISORegSubLangString(language2.ISO_639_1_English)
	isoISORegString += addISORegSubLangString(language2.ISO_639_1_Korean)
	isoISORegString += addISORegSubLangString(language2.ISO_639_1_Japanese)

	isoISORegString += addISORegSubLangString(language2.ISO_639_2T_Chinese)
	isoISORegString += addISORegSubLangString(language2.ISO_639_2T_English)
	isoISORegString += addISORegSubLangString(language2.ISO_639_2T_Korean)
	isoISORegString += addISORegSubLangString(language2.ISO_639_2T_Japanese)

	isoISORegString += addISORegSubLangString(language2.ISO_639_2B_Chinese)

	isoISORegString += language2.RegISORuleEnd

	return isoISORegString
}

func addISORegSubLangString(inLang string) string {
	return fmt.Sprintf(`\b%s\b|`, inLang)
}

var isoISORegString = ""
