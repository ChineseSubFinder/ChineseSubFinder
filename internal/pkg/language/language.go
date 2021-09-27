package language

import (
	"github.com/abadojack/whatlanggo"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/charset"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/axgle/mahonia"
	"github.com/go-creed/sat"
	nzlov "github.com/nzlov/chardet"
	"github.com/saintfish/chardet"
	"strings"
)

// LangConverter 语言转换器
func LangConverter(subLang string) types.Language {
	/*
		xunlei:未知语言、简体&英语、繁体&英语、简体、繁体、英语
	*/
	if strings.Contains(subLang, types.MatchLangDouble) {
		// 双语 - 简英
		return types.ChineseSimpleEnglish
	} else if strings.Contains(subLang, types.MatchLangChs) {
		// 优先简体
		if strings.Contains(subLang, types.MatchLangEn) {
			// 简英
			return types.ChineseSimpleEnglish
		} else if strings.Contains(subLang, types.MatchLangJp) {
			// 简日
			return types.ChineseSimpleJapanese
		} else if strings.Contains(subLang, types.MatchLangKr) {
			// 简韩
			return types.ChineseSimpleKorean
		}
		// 默认简体中文
		return types.ChineseSimple
	} else if strings.Contains(subLang, types.MatchLangCht) {
		// 然后是繁体
		if strings.Contains(subLang, types.MatchLangEn) {
			// 繁英
			return types.ChineseTraditionalEnglish
		} else if strings.Contains(subLang, types.MatchLangJp) {
			// 繁日
			return types.ChineseTraditionalJapanese
		} else if strings.Contains(subLang, types.MatchLangKr) {
			// 繁韩
			return types.ChineseTraditionalKorean
		}
		// 默认繁体中文
		return types.ChineseTraditional
	} else if strings.Contains(subLang, types.MatchLangEn) {
		// 英文
		return types.English
	} else if strings.Contains(subLang, types.MatchLangJp) {
		// 日文
		return types.Japanese
	} else if strings.Contains(subLang, types.MatchLangKr) {
		// 韩文
		return types.Korean
	} else {
		// 都没有，则标记未知
		return types.Unknow
	}
}

// HasChineseLang 是否包含中文
func HasChineseLang(lan types.Language) bool {
	switch lan {
	case types.ChineseSimple,
		types.ChineseTraditional,

		types.ChineseSimpleEnglish,
		types.ChineseTraditionalEnglish,

		types.ChineseSimpleJapanese,
		types.ChineseTraditionalJapanese,

		types.ChineseSimpleKorean,
		types.ChineseTraditionalKorean:
		return true
	default:
		return false
	}
}

// IsBilingualSubtitle 是否是双语字幕
func IsBilingualSubtitle(lan types.Language) bool {
	switch lan {
	case types.ChineseSimpleEnglish,
		types.ChineseTraditionalEnglish,

		types.ChineseSimpleJapanese,
		types.ChineseTraditionalJapanese,

		types.ChineseSimpleKorean,
		types.ChineseTraditionalKorean:
		return true
	default:
		return false
	}
}

// Lang2EmbyNameOld 弃用。从语言转换到 Emby 能够识别的字幕命名
func Lang2EmbyNameOld(lan types.Language) string {
	switch lan {
	case types.Unknow: // 未知语言
		return types.Emby_unknow
	case types.ChineseSimple: // 简体中文
		return types.Emby_chs
	case types.ChineseTraditional: // 繁体中文
		return types.Emby_cht
	case types.ChineseSimpleEnglish: // 简英双语字幕
		return types.Emby_chs_en
	case types.ChineseTraditionalEnglish: // 繁英双语字幕
		return types.Emby_cht_en
	case types.English: // 英文
		return types.Emby_en
	case types.Japanese: // 日语
		return types.Emby_jp
	case types.ChineseSimpleJapanese: // 简日双语字幕
		return types.Emby_chs_jp
	case types.ChineseTraditionalJapanese: // 繁日双语字幕
		return types.Emby_cht_jp
	case types.Korean: // 韩语
		return types.Emby_kr
	case types.ChineseSimpleKorean: // 简韩双语字幕
		return types.Emby_chs_kr
	case types.ChineseTraditionalKorean: // 繁韩双语字幕
		return types.Emby_cht_kr
	default:
		return types.Emby_unknow
	}
}

// Lang2ChineseString 将 types.Language 转换为中文描述：简、繁、简英
func Lang2ChineseString(lan types.Language) string {
	switch lan {
	case types.Unknow: // 未知语言
		return types.MathLangChnUnknow
	case types.ChineseSimple: // 简体中文
		return types.MatchLangChs
	case types.ChineseTraditional: // 繁体中文
		return types.MatchLangCht
	case types.ChineseSimpleEnglish: // 简英双语字幕
		return types.MatchLangChsEn
	case types.ChineseTraditionalEnglish: // 繁英双语字幕
		return types.MatchLangChtEn
	case types.English: // 英文
		return types.MatchLangEn
	case types.Japanese: // 日语
		return types.MatchLangJp
	case types.ChineseSimpleJapanese: // 简日双语字幕
		return types.MatchLangChsJp
	case types.ChineseTraditionalJapanese: // 繁日双语字幕
		return types.MatchLangChtJp
	case types.Korean: // 韩语
		return types.MatchLangKr
	case types.ChineseSimpleKorean: // 简韩双语字幕
		return types.MatchLangChsKr
	case types.ChineseTraditionalKorean: // 繁韩双语字幕
		return types.MatchLangChtKr
	default:
		return types.MathLangChnUnknow
	}
}

// ChineseISOString2Lang 将 中文描述：zh、zho、chi 转换为 types.Language
func ChineseISOString2Lang(chineseStr string) types.Language {

	switch chineseStr {
	case types.ChineseAbbr_639_1, types.ChineseAbbr_639_2T, types.ChineseAbbr_639_2B:
		return types.ChineseSimple
	default:
		return types.Unknow
	}
}

// ChineseString2Lang 将 中文描述：简、繁、简英 转换为 types.Language
func ChineseString2Lang(chineseStr string) types.Language {
	switch chineseStr {
	case types.MathLangChnUnknow: // 未知语言
		return types.Unknow
	case types.MatchLangChs: // 简体中文
		return types.ChineseSimple
	case types.MatchLangCht: // 繁体中文
		return types.ChineseTraditional
	case types.MatchLangChsEn: // 简英双语字幕
		return types.ChineseSimpleEnglish
	case types.MatchLangChtEn: // 繁英双语字幕
		return types.ChineseTraditionalEnglish
	case types.MatchLangEn: // 英文
		return types.English
	case types.MatchLangJp: // 日语
		return types.Japanese
	case types.MatchLangChsJp: // 简日双语字幕
		return types.ChineseSimpleJapanese
	case types.MatchLangChtJp: // 繁日双语字幕
		return types.ChineseTraditionalJapanese
	case types.MatchLangKr: // 韩语
		return types.Korean
	case types.MatchLangChsKr: // 简韩双语字幕
		return types.ChineseSimpleKorean
	case types.MatchLangChtKr: // 繁韩双语字幕
		return types.ChineseTraditionalKorean
	default:
		return types.Unknow
	}
}

// GetLangOptions 语言识别的 Options Whitelist
func GetLangOptions() whatlanggo.Options {
	return whatlanggo.Options{
		Whitelist: map[whatlanggo.Lang]bool{
			whatlanggo.Cmn: true, // 中文	11
			whatlanggo.Eng: true, // 英文	15
			whatlanggo.Jpn: true, // 日文	32
			whatlanggo.Kor: true, // 韩文	37
		},
	}
}

// IsWhiteListLang 是否是白名单语言
func IsWhiteListLang(lang whatlanggo.Lang) bool {
	switch lang {
	// 中文 英文 日文 韩文
	case whatlanggo.Cmn, whatlanggo.Eng, whatlanggo.Jpn, whatlanggo.Kor:
		return true
	default:
		return false
	}
}

// DetectSubLangAndStatistics 检测语言然后统计
func DetectSubLangAndStatistics(lines []string, langDict map[int]int, chLines *[]string, otherLines *[]string) {

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
		// 统计中文有多少行
		if info.Lang == whatlanggo.Cmn {
			*chLines = append(*chLines, line)
		} else {
			*otherLines = append(*otherLines, line)
		}
	}
}

// SubLangStatistics2SubLangType 由分析的信息转换为具体是什么字幕的语言类型
func SubLangStatistics2SubLangType(countLineFeed, AllLines float32, langDict map[int]int, chLines []string) types.Language {
	const basePer = 0.8
	// 是否是双语？
	isDouble := false
	perLines := countLineFeed / AllLines
	// 第二行字幕出现的概率大于 80% 应该稳了吧，不然还能三语？
	if perLines > basePer {
		isDouble = true
	}
	// 中文(包含了 chs 以及 cht，这一级是无法区分的，需要额外的简体和繁体区分方法)
	countChinese, hasChinese := langDict[int(whatlanggo.Cmn)]
	// 英文
	countEnglish, hasEnglish := langDict[int(whatlanggo.Eng)]
	// 日文
	countJapanese, hasJapanese := langDict[int(whatlanggo.Jpn)]
	// 韩文
	countKorean, hasKorean := langDict[int(whatlanggo.Kor)]
	// 0 - No , 1 - Chs, 2 - Cht
	isNoOrChsOrCht := 0
	isChsCount := 0
	if hasChinese {
		for _, line := range chLines {
			if chDict.IsChs(line, 0.9) == true {
				isChsCount++
			}
		}
		// 简体句子的占比超过 80%
		if float32(isChsCount)/float32(len(chLines)) > 0.8 {
			isNoOrChsOrCht = 1
		} else {
			isNoOrChsOrCht = 2
		}
	}

	// 这里有一种情况，就是双语的字幕不是在一个时间轴上的，而是分成两个时间轴的
	// 那么之前的 isDouble 判断就失效了，需要补判一次
	if isDouble == false {
		if hasChinese && hasEnglish {
			isDouble = isDoubleLang(countChinese, countEnglish)
		}
		if hasChinese && hasJapanese {
			isDouble = isDoubleLang(countChinese, countJapanese)
		}
		if hasChinese && hasKorean {
			isDouble = isDoubleLang(countChinese, countKorean)
		}
	}

	// 优先判断双语
	if isDouble == true {
		// 首先得在外面统计就知道是双语
		if hasChinese && hasEnglish {
			// 简体	英文
			return chIsChsOrCht(types.ChineseSimpleEnglish, isNoOrChsOrCht)
		} else if hasChinese && hasJapanese {
			// 简体 日文
			return chIsChsOrCht(types.ChineseSimpleJapanese, isNoOrChsOrCht)
		} else if hasChinese && hasKorean {
			// 简体 韩文
			return chIsChsOrCht(types.ChineseSimpleKorean, isNoOrChsOrCht)
		} else if hasChinese {
			return chIsChsOrCht(types.ChineseSimple, isNoOrChsOrCht)
		} else if hasEnglish {
			return types.English
		} else if hasJapanese {
			return types.Japanese
		} else if hasKorean {
			return types.Korean
		} else {
			return types.Unknow
		}
	} else {
		// 如果比例达不到，那么就是单语言，所以最多的那个就是当前的语言
		// 这里的字典是有可能出现
		if hasChinese {
			// 那么起码要占比 80% 对吧
			perLines = float32(countChinese) / AllLines
			if perLines > basePer {
				return chIsChsOrCht(types.ChineseSimple, isNoOrChsOrCht)
			}
		}
		if hasEnglish {
			// 那么起码要占比 80% 对吧
			perLines = float32(countEnglish) / AllLines
			if perLines > basePer {
				return types.English
			}
		}
		if hasJapanese {
			// 那么起码要占比 80% 对吧
			perLines = float32(countJapanese) / AllLines
			if perLines > basePer {
				return types.Japanese
			}
		}
		if hasKorean {
			// 那么起码要占比 80% 对吧
			perLines = float32(countKorean) / AllLines
			if perLines > basePer {
				return types.Korean
			}
		}

		return types.Unknow
	}

}

// 跟中文相关的再使用，其他的无需传入
func chIsChsOrCht(language types.Language, isNoOrChsOrCht int) types.Language {
	// 输出原来的
	if isNoOrChsOrCht == 0 || isNoOrChsOrCht == 1 {
		return language
	}
	switch language {
	case types.ChineseSimpleEnglish:
		// 简体	英文
		return types.ChineseTraditionalEnglish
	case types.ChineseSimpleJapanese:
		// 简体 日文
		return types.ChineseTraditionalJapanese
	case types.ChineseSimpleKorean:
		// 简体 韩文
		return types.ChineseTraditionalKorean
	case types.ChineseSimple:
		// 简体
		return types.ChineseTraditional
	default:
		return language
	}
}

// IsChineseSimpleOrTraditional 暂时弃用，在 SubLangStatistics2SubLangType 检测语言，通过 unicode 做到。 从字幕的文件名称中尝试确认是简体还是繁体，不需要判断双语问题，有额外的解析器完成。只可能出现 ChineseSimple ChineseTraditional Unknow 三种情况
func IsChineseSimpleOrTraditional(inputFileName string, orgLang types.Language) types.Language {
	if strings.Contains(inputFileName, types.SubNameKeywordChineseSimple) || strings.Contains(inputFileName, types.MatchLangChs) {
		// 简体中文关键词的匹配
		return orgLang
	} else if strings.Contains(inputFileName, types.SubNameKeywordTraditional) || strings.Contains(inputFileName, types.MatchLangCht) {
		// 繁体中文关键词的匹配
		if orgLang == types.ChineseSimple {
			// 简体 -> 繁体
			return types.ChineseTraditional
		} else if orgLang == types.ChineseSimpleEnglish {
			// 简体英文 -> 繁体英文
			return types.ChineseTraditionalEnglish
		} else if orgLang == types.ChineseSimpleJapanese {
			// 简体日文 -> 繁体日文
			return types.ChineseTraditionalJapanese
		} else if orgLang == types.ChineseSimpleKorean {
			// 简体韩文 -> 繁体韩文
			return types.ChineseTraditionalKorean
		}
		// 进来了都不是，那么就返回原来的语言
		return orgLang
	} else {
		// 都没有匹配上，返回原来识别出来的类型即可
		return orgLang
	}
}

// ConvertToString 将字符串从原始编码转换到目标编码，需要配合字符串检测编码库使用 chardet.NewTextDetector()
func ConvertToString(src string, srcCode string, tagCode string) string {
	defer func() {
		if err := recover(); err != nil {
			log_helper.GetLogger().Errorln("ConvertToString panic:", err)
		}
	}()
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

// 感谢: https://blog.csdn.net/gaoluhua/article/details/109128154，解决了编码问题

// ChangeFileCoding2UTF8 自动检测文件的编码，然后转换到 UTF-8
func ChangeFileCoding2UTF8(inBytes []byte) ([]byte, error) {
	best, err := detector.DetectBest(inBytes)
	utf8String := ""
	if err != nil {
		return nil, err
	}
	if best.Confidence < 90 {
		detectBest := nzlov.Mostlike(inBytes)
		utf8String, err = charset.ToUTF8(charset.Charset(detectBest), string(inBytes))
	} else {
		utf8String, err = charset.ToUTF8(charset.Charset(best.Charset), string(inBytes))
	}
	if err != nil {
		return nil, err
	}
	if utf8String == "" {
		return inBytes, nil
	}
	return []byte(utf8String), nil
}

func isDoubleLang(count0, count1 int) bool {
	if count0 >= count1 {
		f := float32(count0) / float32(count1)
		if f >= 1 && f <= 1.4 {
			return true
		} else {
			return false
		}
	} else {
		f := float32(count1) / float32(count0)
		if f >= 1 && f <= 1.4 {
			return true
		} else {
			return false
		}
	}
}

var (
	chDict   = sat.DefaultDict()
	detector = chardet.NewTextDetector()
)
