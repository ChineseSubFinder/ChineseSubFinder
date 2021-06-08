package common

import "strings"

// LangConverter 语言转换器
func LangConverter(subLang string) Language {
	/*
		xunlei:未知语言、简体&英语、繁体&英语、简体、繁体、英语
	*/
	if strings.Contains(subLang, MatchLangChs) {
		// 优先简体
		if strings.Contains(subLang, MatchLangEn) {
			// 简英
			return ChineseSimpleEnglish
		} else if strings.Contains(subLang, MatchLangJp) {
			// 简日
			return ChineseSimpleJapanese
		} else if strings.Contains(subLang, MatchLangKr) {
			// 简韩
			return ChineseSimpleKorean
		}
		// 默认简体中文
		return ChineseSimple
	} else if strings.Contains(subLang, MatchLangCht) {
		// 然后是繁体
		if strings.Contains(subLang, MatchLangEn) {
			// 繁英
			return ChineseTraditionalEnglish
		} else if strings.Contains(subLang, MatchLangJp) {
			// 繁日
			return ChineseTraditionalJapanese
		} else if strings.Contains(subLang, MatchLangKr) {
			// 繁韩
			return ChineseTraditionalKorean
		}
		// 默认繁体中文
		return ChineseTraditional
	} else if strings.Contains(subLang, MatchLangEn) {
		// 英文
		return English
	} else if strings.Contains(subLang, MatchLangJp) {
		// 日文
		return Japanese
	} else if strings.Contains(subLang, MatchLangKr) {
		// 韩文
		return Korean
	} else {
		// 都没有，则标记未知
		return Unknow
	}
}

// Language 语言类型，注意，这里默认还是查找的是中文字幕，只不过下载的时候可能附带了其他的
type Language int
const (
	Unknow	Language = iota				// 未知语言
	ChineseSimple    					// 简体中文
	ChineseTraditional					// 繁体中文
	ChineseSimpleEnglish				// 简英双语字幕
	ChineseTraditionalEnglish			// 繁英双语字幕
	English								// 英文
	Japanese							// 日语
	ChineseSimpleJapanese				// 简日双语字幕
	ChineseTraditionalJapanese			// 繁日双语字幕
	Korean								// 韩语
	ChineseSimpleKorean					// 简韩双语字幕
	ChineseTraditionalKorean			// 繁韩双语字幕
)

const (
	MathLangChnUnknow = "未知语言"
	MatchLangChs      = "简"
	MatchLangCht      = "繁"
	MatchLangChsEn    = "简英"
	MatchLangChtEn    = "繁英"
	MatchLangEn       = "英"
	MatchLangJp       = "日"
	MatchLangChsJp    = "简日"
	MatchLangChtJp    = "繁日"
	MatchLangKr       = "韩"
	MatchLangChsKr    = "简韩"
	MatchLangChtKr    = "繁韩"
)

func (l Language) String() string {
	switch l {
	case ChineseSimple:
		return MatchLangChs
	case ChineseTraditional:
		return MatchLangCht
	case ChineseSimpleEnglish:
		return MatchLangChsEn
	case ChineseTraditionalEnglish:
		return MatchLangChtEn
	case English:
		return MatchLangEn
	case Japanese:
		return MatchLangJp
	case ChineseSimpleJapanese:
		return MatchLangChsJp
	case ChineseTraditionalJapanese:
		return MatchLangChtJp
	case Korean:
		return MatchLangKr
	case ChineseSimpleKorean:
		return MatchLangChsKr
	case ChineseTraditionalKorean:
		return MatchLangChtKr
	}

	return MathLangChnUnknow
}
