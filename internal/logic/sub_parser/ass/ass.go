package ass

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/regex_things"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/sirupsen/logrus"
)

type Parser struct {
	log *logrus.Logger
}

func NewParser(log *logrus.Logger) *Parser {
	return &Parser{log: log}
}

func (p Parser) GetParserName() string {
	return "ass"
}

/*
	DetermineFileTypeFromFile 确定字幕文件的类型，是双语字幕或者某一种语言等等信息
	当 error 是 common.DetermineFileTypeFromFileExtNotFitASSorSSA
	需要额外的处理逻辑，比如不用报错，而是跳过后续的逻辑
*/
func (p Parser) DetermineFileTypeFromFile(filePath string) (bool, *subparser.FileInfo, error) {
	nowExt := filepath.Ext(filePath)

	if p.log != nil {
		p.log.Debugln("DetermineFileTypeFromFile", p.GetParserName(), filePath)
	}

	fBytes, err := os.ReadFile(filePath)
	if err != nil {
		return false, nil, err
	}
	inBytes, err := language.ChangeFileCoding2UTF8(fBytes)
	if err != nil {
		return false, nil, err
	}
	return p.DetermineFileTypeFromBytes(inBytes, nowExt)
}

// DetermineFileTypeFromBytes 确定字幕文件的类型，是双语字幕或者某一种语言等等信息
func (p Parser) DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (bool, *subparser.FileInfo, error) {
	allString := string(inBytes)
	// 注意，需要替换掉 \r 不然正则表达式会有问题
	allString = strings.ReplaceAll(allString, "\r", "")
	// 找到 start end text
	matched := regex_things.ReMatchDialogueASS.FindAllStringSubmatch(allString, -1)
	if matched == nil || len(matched) < 1 {
		if p.log != nil {
			p.log.Debugln("DetermineFileTypeFromBytes can't found DialoguesFilter, Skip")
		}
		return false, nil, nil
	}
	subFileInfo := subparser.FileInfo{}
	subFileInfo.Content = string(inBytes)
	subFileInfo.Ext = nowExt
	subFileInfo.Dialogues = make([]subparser.OneDialogue, 0)
	subFileInfo.DialoguesFilter = make([]subparser.OneDialogue, 0)
	// 这里需要统计一共有几个 \N，以及这个数量在整体行数中的比例，这样就知道是不是双语字幕了
	countLineFeed := 0
	// 有意义的对话统计数，排除 Style 类型
	usefullyDialogueCount := 0
	// 先进行字幕 StyleName 的出现次数排序，找到最多的，就是常规字幕的，不是特效的
	var nameMap = make(map[string]int)
	for _, oneLine := range matched {
		nowStyleName := oneLine[3]
		_, ok := nameMap[nowStyleName]
		if ok == false {
			nameMap[nowStyleName] = 1
		} else {
			nameMap[nowStyleName]++
		}
	}
	/*
		现在可能会遇到两种可能出现的双语字幕：
		1.
			一个 Dialogue 中，直接描述两个语言
		2.
			排序的目标是找出 Name 有几种，一般来说都是 Default 一种
			但是目前也会有用这个 Name 来做双语标记的
			比如相同的时间点：一个 Name 是 Chs Subtitle
							一个 Name 是 Eng Subtitle
			那么排序来说，就应该是 Top1、2 两个

		但是之前是为了剔除某一些特效动画，进行排序后只找 Top 1，但是遇到上面 2 的情况
		解析就只读取到一个语言的字幕了

		那么现在的解决方案就是，一开始先进行 Name 的统计。
		然后统计是否有一个相同的时间段，出现了两个 Dialogue，比如：
		0:01:01.00-0:01:11.00 这个时间段，一共有两个 Dialogue 使用了，然后需要统计这种情况占比所有的 Dialogue 的比例
		如果比例很高，那么就认为是情况 2 的双语字幕
		如果没有那么多，或者就没得。就任务是情况 1 的双语字幕，这个也不能说就是双语字幕，只不过走之前的逻辑就够了。
	*/
	mapByValue := sortMapByValue(nameMap)

	// 把所有的对白缓存下来，其实优先是把时间信息缓存，其他信息无所谓
	p.oneLineSubDialogueParser0(matched, &subFileInfo)

	if p.detectOneOrTwoLineDialogue(matched) == true {
		// 情况1
		usefullyDialogueCount, countLineFeed = p.oneLineSubDialogueParser1(matched, mapByValue, &subFileInfo)
	} else {
		// 情况2
		usefullyDialogueCount, countLineFeed = p.oneLineSubDialogueParser2(matched, mapByValue, &subFileInfo)
	}
	// 再分析
	// 需要判断每一个 Line 是啥语言，[语言的code]次数
	var langDict map[int]int
	langDict = make(map[int]int)
	// 抽取出所有的中文对话
	var chLines = make([]string, 0)
	// 抽取出所有的第二语言对话
	var otherLines = make([]string, 0)
	// 抽取出来的对话数组，为了后续用来匹配和修改时间轴
	var usefulDialogueExs = make([]subparser.OneDialogueEx, 0)
	// 在这之前需要把 subFileInfo.DialoguesFilter 的内容填好，Lines 这里如果是单种语言应该就是一个元素，如果是双语就需要拆分成两个元素
	// 这样向后传递就简单了，也统一了
	emptyLines := 0
	for _, dialogue := range subFileInfo.DialoguesFilter {
		emptyLines += language.DetectSubLangAndStatistics(dialogue, langDict, &usefulDialogueExs, &chLines, &otherLines)
	}
	// 从统计出来的字典，找出 Top 1 或者 2 的出来，然后计算出是什么语言的字幕
	detectLang := language.SubLangStatistics2SubLangType(float32(countLineFeed), float32(usefullyDialogueCount-emptyLines), langDict, chLines)
	subFileInfo.Lang = detectLang
	subFileInfo.Data = inBytes
	subFileInfo.DialoguesFilterEx = usefulDialogueExs
	subFileInfo.CHLines = chLines
	subFileInfo.OtherLines = otherLines
	return true, &subFileInfo, nil
}

// oneLineSubDialogueParser0 情况 0 时候的解析器，不过滤，只要是对白都加进去
func (p Parser) oneLineSubDialogueParser0(matched [][]string, subFileInfo *subparser.FileInfo) {

	for _, oneLine := range matched {
		startTime := oneLine[1]
		endTime := oneLine[2]
		nowStyleName := oneLine[3]
		nowText := oneLine[4]
		odl := subparser.OneDialogue{
			StyleName: nowStyleName,
			StartTime: startTime,
			EndTime:   endTime,
			Lines:     []string{nowText},
		}
		subFileInfo.Dialogues = append(subFileInfo.Dialogues, odl)
	}
}

// oneLineSubDialogueParser1 情况 1 时候的解析器
func (p Parser) oneLineSubDialogueParser1(matched [][]string, mapByValue StyleNameInfos, subFileInfo *subparser.FileInfo) (int, int) {

	var countLineFeed = 0
	var usefullyDialogueCount = 0
	// 先读取一次字幕文件
	for _, oneLine := range matched {

		if len(oneLine) < 1 {
			continue
		}

		// 排除特效内容，只统计有意义的对话部分
		if strings.Contains(oneLine[0], mapByValue[0].Name) == false {
			continue
		}
		usefullyDialogueCount++

		startTime := oneLine[1]
		endTime := oneLine[2]
		nowStyleName := oneLine[3]
		nowText := oneLine[4]
		odl := subparser.OneDialogue{
			StyleName: nowStyleName,
			StartTime: startTime,
			EndTime:   endTime,
		}
		odl.Lines = make([]string, 0)
		countLineFeed = p.parseOneDialogueText(nowText, &odl, countLineFeed)

		subFileInfo.DialoguesFilter = append(subFileInfo.DialoguesFilter, odl)
	}
	return usefullyDialogueCount, countLineFeed
}

// oneLineSubDialogueParser2 情况 2 时候的解析器
func (p Parser) oneLineSubDialogueParser2(matched [][]string, mapByValue StyleNameInfos, subFileInfo *subparser.FileInfo) (int, int) {

	var countLineFeed = 0
	var usefullyDialogueCount = 0
	//var timeMap = make(map[string]subparser.OneDialogue, 0)
	// 更换数据结构的原因是为了能够使用顺序，go 内置的 map 不是顺序的，是随机的，会导致后续的逻辑出问题
	var timeMap = treemap.NewWithStringComparator()
	// 先读取一次字幕文件
	for _, oneLine := range matched {

		usefullyDialogueCount++
		// 这里可能会统计到特效的部分，但是这里忽略这个问题，因为目标不是这个
		// 统计 Dialogue 的开始和结束时间
		startTime := oneLine[1]
		endTime := oneLine[2]
		nowStyleName := oneLine[3]
		nowText := oneLine[4]
		mergeTime := startTime + "_" + endTime
		value, ok := timeMap.Get(mergeTime)
		if ok == false {
			// 首次新增
			odl := subparser.OneDialogue{
				StyleName: nowStyleName,
				StartTime: startTime,
				EndTime:   endTime,
			}
			odl.Lines = make([]string, 0)
			countLineFeed = p.parseOneDialogueText(nowText, &odl, countLineFeed)
			timeMap.Put(mergeTime, odl)
		} else {
			// 双语
			odl := value.(subparser.OneDialogue)
			countLineFeed = p.parseOneDialogueText(nowText, &odl, countLineFeed)
			timeMap.Put(mergeTime, odl)
		}
	}

	for _, value := range timeMap.Values() {
		odl := value.(subparser.OneDialogue)
		subFileInfo.DialoguesFilter = append(subFileInfo.DialoguesFilter, odl)
	}

	return usefullyDialogueCount, countLineFeed
}

// parseOneDialogueText 对话的对白内容解析
func (p Parser) parseOneDialogueText(nowText string, odl *subparser.OneDialogue, countLineFeed int) int {
	// nowText 优先移除 \h 这个是替换空格， \h 是让两个词在一行，不换行显示
	nowText = strings.ReplaceAll(nowText, `\h`, " ")
	// nowText 这个需要先把 {} 花括号内的内容给移除
	nowText1 := regex_things.ReMatchBrace.ReplaceAllString(nowText, "")
	nowText1 = regex_things.ReMatchBracket.ReplaceAllString(nowText1, "")
	nowText1 = strings.TrimRight(nowText1, "\r")
	// 然后判断是否有 \N 或者 \n
	// 直接把 \n 替换为 \N 来解析
	nowText1 = strings.ReplaceAll(nowText1, `\n`, `\N`)
	if strings.Contains(nowText1, `\N`) {
		// 有，那么就需要再次切割，一般是双语字幕
		for _, matched2 := range regex_things.ReCutDoubleLanguage.FindAllStringSubmatch(nowText1, -1) {
			if matched2 == nil {
				continue
			}
			for i, s := range matched2 {
				if i == 0 {
					continue
				}
				s = strings.ReplaceAll(s, `\N`, "")
				odl.Lines = append(odl.Lines, s)
			}
		}
		countLineFeed++
	} else {
		// 无，则可以直接添加
		nowText1 = strings.ReplaceAll(nowText1, `\N`, "")
		odl.Lines = append(odl.Lines, nowText1)
	}
	return countLineFeed
}

// detectOneOrTwoLineDialogue 优先检测一次字幕文件，可能存在的双语字幕的情况，是 1 还是 2 ，详细解释看调用此函数前的解释
func (p Parser) detectOneOrTwoLineDialogue(matched [][]string) bool {
	/*
		这里判断的方法粗暴一点，直接判断两个 Dialogue 都是一个时间段的比例是多少，达到了就是情况2，不是就是情况1
	*/
	allDialogue := len(matched)
	twoLine := 0
	var timeMap = make(map[string]int, 0)
	// 先读取一次字幕文件
	for _, oneLine := range matched {
		// 这里可能会统计到特效的部分，但是这里忽略这个问题，因为目标不是这个
		// 统计 Dialogue 的开始和结束时间
		startTime := oneLine[1]
		endTime := oneLine[2]

		mergeTime := startTime + "_" + endTime
		_, ok := timeMap[mergeTime]
		if ok == false {
			timeMap[mergeTime] = 1
		} else {
			timeMap[mergeTime]++
			if timeMap[mergeTime] == 2 {
				twoLine++
			}
		}
	}
	// 目前看到的文件大概再 47% 以上，考虑到更多的“注释”、“特效”,至少有 38% 就够了
	per := float64(twoLine) / float64(allDialogue)
	if per > 0.38 {
		// 使用情况2的字幕分析方式
		return false
	}
	// 使用情况1的字幕分析方式
	return true
}
