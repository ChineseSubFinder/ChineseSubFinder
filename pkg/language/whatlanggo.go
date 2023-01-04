package language

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
	"github.com/abadojack/whatlanggo"
)

// WhichChineseType 是简体中文（1）还是繁体中文（2），如果都不是，那么是 0
func WhichChineseType(inputString string) int {

	info := whatlanggo.DetectWithOptions(inputString, GetLangOptions())
	// 是否是中文（简体、繁体）
	if info.Lang == whatlanggo.Cmn {
		// 判断是简体还是繁体
		if ChDict.IsChs(inputString, 0.9) == true {
			return 1
		} else {
			return 2
		}
	}
	return 0
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
func DetectSubLangAndStatistics(oneDialogue subparser.OneDialogue, langDict map[int]int,
	usefulDialogueEx *[]subparser.OneDialogueEx, chLines *[]string, otherLines *[]string) int {

	var oneDialogueEx subparser.OneDialogueEx
	oneDialogueEx.StartTime = oneDialogue.StartTime
	oneDialogueEx.EndTime = oneDialogue.EndTime

	emptyLine := 0

	for _, line := range oneDialogue.Lines {

		if line == "" {
			emptyLine++
			continue
		}

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
		// 这里可能是一个 dialogue 里面有两句话，而且两句话都是一个类型的语言，所以其实需要的是合并
		switch info.Lang {
		case whatlanggo.Cmn:
			oneDialogueEx.ChLine += line + " "
		case whatlanggo.Eng:
			oneDialogueEx.EnLine += line + " "
		case whatlanggo.Kor:
			oneDialogueEx.KrLine += line + " "
		case whatlanggo.Jpn:
			oneDialogueEx.JpLine += line + " "
		}
	}

	*usefulDialogueEx = append(*usefulDialogueEx, oneDialogueEx)

	return emptyLine
}

// SubLangStatistics2SubLangType 由分析的信息转换为具体是什么字幕的语言类型
func SubLangStatistics2SubLangType(countLineFeed, AllLines float32, langDict map[int]int, chLines []string) language.MyLanguage {
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
			// 判断是简体还是繁体
			if ChDict.IsChs(line, 0.9) == true {
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
			return chIsChsOrCht(language.ChineseSimpleEnglish, isNoOrChsOrCht)
		} else if hasChinese && hasJapanese {
			// 简体 日文
			return chIsChsOrCht(language.ChineseSimpleJapanese, isNoOrChsOrCht)
		} else if hasChinese && hasKorean {
			// 简体 韩文
			return chIsChsOrCht(language.ChineseSimpleKorean, isNoOrChsOrCht)
		} else if hasChinese {
			return chIsChsOrCht(language.ChineseSimple, isNoOrChsOrCht)
		} else if hasEnglish {
			return language.English
		} else if hasJapanese {
			return language.Japanese
		} else if hasKorean {
			return language.Korean
		} else {
			return language.Unknown
		}
	} else {
		// 如果比例达不到，那么就是单语言，所以最多的那个就是当前的语言
		// 这里的字典是有可能出现
		/*
			这里的 AllLines 需要考虑一点，字幕内有很多特殊的背景声音旁白
			那么会再上面提出的时候，直接把这一句话设置为空，那么这里所有的对白数量应该减去这些被检测为空的对白
		*/
		if hasChinese {
			// 那么起码要占比 80% 对吧
			perLines = float32(countChinese) / AllLines
			if perLines > basePer {
				return chIsChsOrCht(language.ChineseSimple, isNoOrChsOrCht)
			}
		}
		if hasEnglish {
			// 那么起码要占比 80% 对吧
			perLines = float32(countEnglish) / AllLines
			if perLines > basePer {
				return language.English
			}
		}
		if hasJapanese {
			// 那么起码要占比 80% 对吧
			perLines = float32(countJapanese) / AllLines
			if perLines > basePer {
				return language.Japanese
			}
		}
		if hasKorean {
			// 那么起码要占比 80% 对吧
			perLines = float32(countKorean) / AllLines
			if perLines > basePer {
				return language.Korean
			}
		}

		return language.Unknown
	}

}

// 跟中文相关的再使用，其他的无需传入
func chIsChsOrCht(inLanguage language.MyLanguage, isNoOrChsOrCht int) language.MyLanguage {
	// 输出原来的
	if isNoOrChsOrCht == 0 || isNoOrChsOrCht == 1 {
		return inLanguage
	}
	switch inLanguage {
	case language.ChineseSimpleEnglish:
		// 简体	英文
		return language.ChineseTraditionalEnglish
	case language.ChineseSimpleJapanese:
		// 简体 日文
		return language.ChineseTraditionalJapanese
	case language.ChineseSimpleKorean:
		// 简体 韩文
		return language.ChineseTraditionalKorean
	case language.ChineseSimple:
		// 简体
		return language.ChineseTraditional
	default:
		return inLanguage
	}
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
