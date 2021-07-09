package model

import (
	"github.com/abadojack/whatlanggo"
	"github.com/allanpk716/ChineseSubFinder/charset"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/axgle/mahonia"
	"github.com/go-creed/sat"
	chardet2 "github.com/nzlov/chardet"
	//"github.com/qiniu/iconv"
	"github.com/saintfish/chardet"
	"strings"
)

// LangConverter 语言转换器
func LangConverter(subLang string) common.Language {
	/*
		xunlei:未知语言、简体&英语、繁体&英语、简体、繁体、英语
	*/
	if strings.Contains(subLang, common.MatchLangDouble) {
		// 双语 - 简英
		return common.ChineseSimpleEnglish
	} else if strings.Contains(subLang, common.MatchLangChs) {
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

// IsBilingualSubtitle 是否是双语字幕
func IsBilingualSubtitle(lan common.Language) bool {
	switch lan {
	case common.ChineseSimpleEnglish,
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

// Lang2EmbyName 从语言转换到 Emby 能够识别的字幕命名
func Lang2EmbyName(lan common.Language) string {
	switch lan {
	case common.Unknow:                     				// 未知语言
		return common.Emby_unknow
	case common.ChineseSimple:                              // 简体中文
		return common.Emby_chs
	case common.ChineseTraditional:                         // 繁体中文
		return common.Emby_cht
	case common.ChineseSimpleEnglish:                       // 简英双语字幕
		return common.Emby_chs_en
	case common.ChineseTraditionalEnglish:                  // 繁英双语字幕
		return common.Emby_cht_en
	case common.English:                                    // 英文
		return common.Emby_en
	case common.Japanese:                                   // 日语
		return common.Emby_jp
	case common.ChineseSimpleJapanese:                      // 简日双语字幕
		return common.Emby_chs_jp
	case common.ChineseTraditionalJapanese:                 // 繁日双语字幕
		return common.Emby_cht_jp
	case common.Korean:                                     // 韩语
		return common.Emby_kr
	case common.ChineseSimpleKorean:                        // 简韩双语字幕
		return common.Emby_chs_kr
	case common.ChineseTraditionalKorean:                   // 繁韩双语字幕
		return common.Emby_cht_kr
	default:
		return common.Emby_unknow
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
func DetectSubLangAndStatistics(lines []string, langDict map[int]int, chLines *[]string) {

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
		}
	}
}

// SubLangStatistics2SubLangType 由分析的信息转换为具体是什么字幕的语言类型
func SubLangStatistics2SubLangType(countLineFeed, AllLines float32, langDict map[int]int, chLines []string) common.Language {
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
		if float32(isChsCount) / float32(len(chLines)) > 0.8 {
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
			return chIsChsOrCht(common.ChineseSimpleEnglish, isNoOrChsOrCht)
		} else if hasChinese && hasJapanese {
			// 简体 日文
			return chIsChsOrCht(common.ChineseSimpleJapanese, isNoOrChsOrCht)
		} else if hasChinese && hasKorean {
			// 简体 韩文
			return chIsChsOrCht(common.ChineseSimpleKorean, isNoOrChsOrCht)
		} else if hasChinese {
			return chIsChsOrCht(common.ChineseSimple, isNoOrChsOrCht)
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
				return chIsChsOrCht(common.ChineseSimple, isNoOrChsOrCht)
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

// 跟中文相关的再使用，其他的无需传入
func chIsChsOrCht(language common.Language, isNoOrChsOrCht int) common.Language {
	// 输出原来的
	if isNoOrChsOrCht == 0 || isNoOrChsOrCht == 1 {
		return language
	}
	switch language {
	case common.ChineseSimpleEnglish:
		// 简体	英文
		return common.ChineseTraditionalEnglish
	case common.ChineseSimpleJapanese:
		// 简体 日文
		return common.ChineseTraditionalJapanese
	case common.ChineseSimpleKorean:
		// 简体 韩文
		return common.ChineseTraditionalKorean
	case common.ChineseSimple:
		// 简体
		return common.ChineseTraditional
	default:
		return language
	}
}

// IsChineseSimpleOrTraditional 暂时弃用，在 SubLangStatistics2SubLangType 检测语言，通过 unicode 做到。 从字幕的文件名称中尝试确认是简体还是繁体，不需要判断双语问题，有额外的解析器完成。只可能出现 ChineseSimple ChineseTraditional Unknow 三种情况
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

// ConvertToString 将字符串从原始编码转换到目标编码，需要配合字符串检测编码库使用 chardet.NewTextDetector()
func ConvertToString(src string, srcCode string, tagCode string) string {
	defer func() {
		if err := recover(); err != nil {
			GetLogger().Errorln("ConvertToString panic:", err)
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
		detectBest := chardet2.Mostlike(inBytes)
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
	chDict = sat.DefaultDict()
	detector = chardet.NewTextDetector()
)