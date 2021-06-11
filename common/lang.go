package common

import (
	"github.com/abadojack/whatlanggo"
	"strings"
)

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

// HasChineseLang 是否包含中文
func HasChineseLang(lan Language) bool {
	switch lan {
	case ChineseSimple,
	ChineseTraditional,

	ChineseSimpleEnglish,
	ChineseTraditionalEnglish,

	ChineseSimpleJapanese,
	ChineseTraditionalJapanese,

	ChineseSimpleKorean,
	ChineseTraditionalKorean:
		return true
	default:
		return false
	}
}

// GetLangOptions 语言识别的 Options Whitelist
func GetLangOptions() whatlanggo.Options {
	return whatlanggo.Options{
		Whitelist: map[whatlanggo.Lang]bool{
			whatlanggo.Cmn: true,	// 中文	11
			whatlanggo.Eng: true,	// 英文	15
			whatlanggo.Jpn: true,	// 日文	32
			whatlanggo.Kor: true,	// 韩文	37
		},
	}
}
// IsWhiteListLang 是否是白名单语言
func IsWhiteListLang(lang whatlanggo.Lang) bool {
	switch lang {
	// 中文 英文 日文 韩文
	case whatlanggo.Cmn, whatlanggo.Eng,whatlanggo.Jpn,whatlanggo.Kor:
		return true
	default:
		return false
	}
}

// DetectSubLangAndStatistics 检测语言然后统计
func DetectSubLangAndStatistics(lines []string, langDict map[int]int) {
	for _, line := range lines {
		info := whatlanggo.DetectWithOptions(line, GetLangOptions())
		tmpLang := -1
		if IsWhiteListLang(info.Lang) == true {
			tmpLang = (int)(info.Lang)
		}
		// 这一种语言的 key 是否存在，不存在则新建，存在再数值 +1
		value, ok := langDict[tmpLang]
		if ok == true {
			// 累加
			value++
			langDict[tmpLang] = value
		} else {
			langDict[tmpLang] = 1
		}
	}
}

// SubLangStatistics2SubLangType 由分析的信息转换为具体是什么字幕的语言类型
func SubLangStatistics2SubLangType(countLineFeed, AllLines float32, langDict map[int]int) Language {
	const basePer = 0.8
	// 是否是双语？
	isDouble := false
	perLines := countLineFeed / AllLines
	// 第二行字幕出现的概率大于 80% 应该稳了吧，不然还能三语？
	if perLines > basePer {
		isDouble = true
	}
	// TODO 现在是没有很好的办法去识别是简体还是繁体中文的，所以···
	// 中文
	countChinese, hasChinese := langDict[int(whatlanggo.Cmn)]
	// 英文
	countEnglish, hasEnglish := langDict[int(whatlanggo.Eng)]
	// 日文
	countJapanese, hasJapanese := langDict[int(whatlanggo.Jpn)]
	// 韩文
	countKorean, hasKorean := langDict[int(whatlanggo.Kor)]

	// 优先判断双语
	if isDouble == true {
		// 首先得在外面统计就知道是双语
		if hasChinese && hasEnglish {
			// 简体	英文
			return ChineseSimpleEnglish
		} else if hasChinese && hasJapanese {
			// 简体 日文
			return ChineseSimpleJapanese
		} else if hasChinese && hasKorean {
			// 简体 韩文
			return ChineseSimpleKorean
		} else if hasChinese {
			return ChineseSimple
		} else if hasEnglish {
			return English
		} else if hasJapanese {
			return Japanese
		} else if hasKorean {
			return Korean
		} else {
			return Unknow
		}
	} else {
		// 如果比例达不到，那么就是单语言，所以最多的那个就是当前的语言
		// 这里的字典是有可能出现
		if hasChinese {
			// 那么起码要占比 80% 对吧
			perLines = float32(countChinese) / AllLines
			if perLines > basePer {
				return ChineseSimple
			}
		}
		if hasEnglish {
			// 那么起码要占比 80% 对吧
			perLines = float32(countEnglish) / AllLines
			if perLines > basePer {
				return English
			}
		}
		if hasJapanese {
			// 那么起码要占比 80% 对吧
			perLines = float32(countJapanese) / AllLines
			if perLines > basePer {
				return Japanese
			}
		}
		if hasKorean {
			// 那么起码要占比 80% 对吧
			perLines = float32(countKorean) / AllLines
			if perLines > basePer {
				return Korean
			}
		}

		return Unknow
	}

}

// IsChineseSimpleOrTraditional 从字幕的文件名称中尝试确认是简体还是繁体，不需要判断双语问题，有额外的解析器完成。只可能出现 ChineseSimple ChineseTraditional Unknow 三种情况
func IsChineseSimpleOrTraditional(inputFileName string, orgLang Language) Language {

	if strings.Contains(inputFileName, SubNameKeywordChineseSimple) || strings.Contains(inputFileName, MatchLangChs) {
		// 简体中文关键词的匹配
		return orgLang
	} else if strings.Contains(inputFileName, SubNameKeywordTraditional) || strings.Contains(inputFileName, MatchLangCht) {
		// 繁体中文关键词的匹配
		if orgLang == ChineseSimple {
			// 简体 -> 繁体
			return ChineseTraditional
		} else if orgLang == ChineseSimpleEnglish {
			// 简体英文 -> 繁体英文
			return ChineseTraditionalEnglish
		} else if orgLang == ChineseSimpleJapanese {
			// 简体日文 -> 繁体日文
			return ChineseTraditionalJapanese
		} else if orgLang == ChineseSimpleKorean {
			// 简体韩文 -> 繁体韩文
			return ChineseTraditionalKorean
		}
		// 进来了都不是，那么就返回原来的语言
		return orgLang
	} else {
		// 都没有匹配上，返回原来识别出来的类型即可
		return orgLang
	}
}

const (
	SubNameKeywordChineseSimple = "chs"
	SubNameKeywordTraditional	 = "cht"
)

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
	default:
		return MathLangChnUnknow
	}
}
