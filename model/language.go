package model

import (
	"github.com/abadojack/whatlanggo"
	"github.com/allanpk716/ChineseSubFinder/common"
	"strings"
)

// LangConverter 语言转换器
func LangConverter(subLang string) common.Language {
	/*
		xunlei:未知语言、简体&英语、繁体&英语、简体、繁体、英语
	*/
	if strings.Contains(subLang, common.MatchLangChs) {
		// 优先简体
		if strings.Contains(subLang, common.MatchLangEn) {
			// 简英
			return common.ChineseSimpleEnglish
		} else if strings.Contains(subLang, common.MatchLangJp) {
			// 简日
			return common.ChineseSimpleJapanese
		} else if strings.Contains(subLang, common.MatchLangKr) {
			// 简韩
			return common.ChineseSimpleKorean
		}
		// 默认简体中文
		return common.ChineseSimple
	} else if strings.Contains(subLang, common.MatchLangCht) {
		// 然后是繁体
		if strings.Contains(subLang, common.MatchLangEn) {
			// 繁英
			return common.ChineseTraditionalEnglish
		} else if strings.Contains(subLang, common.MatchLangJp) {
			// 繁日
			return common.ChineseTraditionalJapanese
		} else if strings.Contains(subLang, common.MatchLangKr) {
			// 繁韩
			return common.ChineseTraditionalKorean
		}
		// 默认繁体中文
		return common.ChineseTraditional
	} else if strings.Contains(subLang, common.MatchLangEn) {
		// 英文
		return common.English
	} else if strings.Contains(subLang, common.MatchLangJp) {
		// 日文
		return common.Japanese
	} else if strings.Contains(subLang, common.MatchLangKr) {
		// 韩文
		return common.Korean
	} else {
		// 都没有，则标记未知
		return common.Unknow
	}
}

// HasChineseLang 是否包含中文
func HasChineseLang(lan common.Language) bool {
	switch lan {
	case common.ChineseSimple,
		common.ChineseTraditional,

		common.ChineseSimpleEnglish,
		common.ChineseTraditionalEnglish,

		common.ChineseSimpleJapanese,
		common.ChineseTraditionalJapanese,

		common.ChineseSimpleKorean,
		common.ChineseTraditionalKorean:
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
func SubLangStatistics2SubLangType(countLineFeed, AllLines float32, langDict map[int]int) common.Language {
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
			return common.ChineseSimpleEnglish
		} else if hasChinese && hasJapanese {
			// 简体 日文
			return common.ChineseSimpleJapanese
		} else if hasChinese && hasKorean {
			// 简体 韩文
			return common.ChineseSimpleKorean
		} else if hasChinese {
			return common.ChineseSimple
		} else if hasEnglish {
			return common.English
		} else if hasJapanese {
			return common.Japanese
		} else if hasKorean {
			return common.Korean
		} else {
			return common.Unknow
		}
	} else {
		// 如果比例达不到，那么就是单语言，所以最多的那个就是当前的语言
		// 这里的字典是有可能出现
		if hasChinese {
			// 那么起码要占比 80% 对吧
			perLines = float32(countChinese) / AllLines
			if perLines > basePer {
				return common.ChineseSimple
			}
		}
		if hasEnglish {
			// 那么起码要占比 80% 对吧
			perLines = float32(countEnglish) / AllLines
			if perLines > basePer {
				return common.English
			}
		}
		if hasJapanese {
			// 那么起码要占比 80% 对吧
			perLines = float32(countJapanese) / AllLines
			if perLines > basePer {
				return common.Japanese
			}
		}
		if hasKorean {
			// 那么起码要占比 80% 对吧
			perLines = float32(countKorean) / AllLines
			if perLines > basePer {
				return common.Korean
			}
		}

		return common.Unknow
	}

}

// IsChineseSimpleOrTraditional 从字幕的文件名称中尝试确认是简体还是繁体，不需要判断双语问题，有额外的解析器完成。只可能出现 ChineseSimple ChineseTraditional Unknow 三种情况
func IsChineseSimpleOrTraditional(inputFileName string, orgLang common.Language) common.Language {

	if strings.Contains(inputFileName, common.SubNameKeywordChineseSimple) || strings.Contains(inputFileName, common.MatchLangChs) {
		// 简体中文关键词的匹配
		return orgLang
	} else if strings.Contains(inputFileName, common.SubNameKeywordTraditional) || strings.Contains(inputFileName, common.MatchLangCht) {
		// 繁体中文关键词的匹配
		if orgLang == common.ChineseSimple {
			// 简体 -> 繁体
			return common.ChineseTraditional
		} else if orgLang == common.ChineseSimpleEnglish {
			// 简体英文 -> 繁体英文
			return common.ChineseTraditionalEnglish
		} else if orgLang == common.ChineseSimpleJapanese {
			// 简体日文 -> 繁体日文
			return common.ChineseTraditionalJapanese
		} else if orgLang == common.ChineseSimpleKorean {
			// 简体韩文 -> 繁体韩文
			return common.ChineseTraditionalKorean
		}
		// 进来了都不是，那么就返回原来的语言
		return orgLang
	} else {
		// 都没有匹配上，返回原来识别出来的类型即可
		return orgLang
	}
}


