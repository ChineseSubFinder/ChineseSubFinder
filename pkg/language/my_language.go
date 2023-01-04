package language

import (
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
)

// LangConverter4Sub_Supplier 语言转换器，给字幕的提供者实例解析使用（xunlei、zimuku等）
// 支持的字符串语言见 internal/types/language/my_language.go
func LangConverter4Sub_Supplier(subLang string) language.MyLanguage {
	/*
		xunlei:未知语言、简体&英语、繁体&英语、简体、繁体、英语
	*/
	if strings.Contains(subLang, language.MatchLangDouble) {
		// 双语 - 简英
		return language.ChineseSimpleEnglish
	} else if strings.Contains(subLang, language.MatchLangChs) {
		// 优先简体
		if strings.Contains(subLang, language.MatchLangEn) {
			// 简英
			return language.ChineseSimpleEnglish
		} else if strings.Contains(subLang, language.MatchLangJp) {
			// 简日
			return language.ChineseSimpleJapanese
		} else if strings.Contains(subLang, language.MatchLangKr) {
			// 简韩
			return language.ChineseSimpleKorean
		}
		// 默认简体中文
		return language.ChineseSimple
	} else if strings.Contains(subLang, language.MatchLangCht) {
		// 然后是繁体
		if strings.Contains(subLang, language.MatchLangEn) {
			// 繁英
			return language.ChineseTraditionalEnglish
		} else if strings.Contains(subLang, language.MatchLangJp) {
			// 繁日
			return language.ChineseTraditionalJapanese
		} else if strings.Contains(subLang, language.MatchLangKr) {
			// 繁韩
			return language.ChineseTraditionalKorean
		}
		// 默认繁体中文
		return language.ChineseTraditional
	} else if strings.Contains(subLang, language.MatchLangEn) {
		// 英文
		return language.English
	} else if strings.Contains(subLang, language.MatchLangJp) {
		// 日文
		return language.Japanese
	} else if strings.Contains(subLang, language.MatchLangKr) {
		// 韩文
		return language.Korean
	} else {
		// 都没有，则标记未知
		return language.Unknown
	}
}

// HasChineseLang 是否包含中文
func HasChineseLang(lan language.MyLanguage) bool {
	switch lan {
	case language.ChineseSimple,
		language.ChineseTraditional,

		language.ChineseSimpleEnglish,
		language.ChineseTraditionalEnglish,

		language.ChineseSimpleJapanese,
		language.ChineseTraditionalJapanese,

		language.ChineseSimpleKorean,
		language.ChineseTraditionalKorean:
		return true
	default:
		return false
	}
}

// IsBilingualSubtitle 是否是双语字幕
func IsBilingualSubtitle(lan language.MyLanguage) bool {
	switch lan {
	case language.ChineseSimpleEnglish,
		language.ChineseTraditionalEnglish,

		language.ChineseSimpleJapanese,
		language.ChineseTraditionalJapanese,

		language.ChineseSimpleKorean,
		language.ChineseTraditionalKorean:
		return true
	default:
		return false
	}
}

// Lang2ChineseString 将 types.MyLanguage 转换为中文描述：简、繁、简英
// 支持的字符串语言见 internal/types/language/my_language.go
func Lang2ChineseString(lan language.MyLanguage) string {
	switch lan {
	case language.Unknown:
		// 未知语言
		return language.MathLangChnUnknown
	case language.ChineseSimple:
		// 简体中文
		return language.MatchLangChs
	case language.ChineseTraditional:
		// 繁体中文
		return language.MatchLangCht
	case language.ChineseSimpleEnglish:
		// 简英双语字幕
		return language.MatchLangChsEn
	case language.ChineseTraditionalEnglish:
		// 繁英双语字幕
		return language.MatchLangChtEn
	case language.English:
		// 英文
		return language.MatchLangEn
	case language.Japanese:
		// 日语
		return language.MatchLangJp
	case language.ChineseSimpleJapanese:
		// 简日双语字幕
		return language.MatchLangChsJp
	case language.ChineseTraditionalJapanese:
		// 繁日双语字幕
		return language.MatchLangChtJp
	case language.Korean:
		// 韩语
		return language.MatchLangKr
	case language.ChineseSimpleKorean:
		// 简韩双语字幕
		return language.MatchLangChsKr
	case language.ChineseTraditionalKorean:
		// 繁韩双语字幕
		return language.MatchLangChtKr
	default:
		return language.MathLangChnUnknown
	}
}

// ChineseString2Lang 将 中文描述：简、繁、简英 转换为 types.MyLanguage
// 支持的字符串语言见 internal/types/language/my_language.go
func ChineseString2Lang(chineseStr string) language.MyLanguage {
	switch chineseStr {
	case language.MathLangChnUnknown:
		// 未知语言
		return language.Unknown
	case language.MatchLangChs:
		// 简体中文
		return language.ChineseSimple
	case language.MatchLangCht:
		// 繁体中文
		return language.ChineseTraditional
	case language.MatchLangChsEn:
		// 简英双语字幕
		return language.ChineseSimpleEnglish
	case language.MatchLangChtEn:
		// 繁英双语字幕
		return language.ChineseTraditionalEnglish
	case language.MatchLangEn:
		// 英文
		return language.English
	case language.MatchLangJp:
		// 日语
		return language.Japanese
	case language.MatchLangChsJp:
		// 简日双语字幕
		return language.ChineseSimpleJapanese
	case language.MatchLangChtJp:
		// 繁日双语字幕
		return language.ChineseTraditionalJapanese
	case language.MatchLangKr:
		// 韩语
		return language.Korean
	case language.MatchLangChsKr:
		// 简韩双语字幕
		return language.ChineseSimpleKorean
	case language.MatchLangChtKr:
		// 繁韩双语字幕
		return language.ChineseTraditionalKorean
	default:
		return language.Unknown
	}
}
