package ass

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
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
	if strings.ToLower(nowExt) != common.SubExtASS && strings.ToLower(nowExt) != common.SubExtSSA {
		return false, nil, nil
	}
	fBytes, err := ioutil.ReadFile(filePath)
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
	matched := sub_parser.ReMatchDialogueASS.FindAllStringSubmatch(allString, -1)
	if len(matched) < 1 {
		return false, nil, nil
	}
	subFileInfo := subparser.FileInfo{}
	subFileInfo.Ext = nowExt
	subFileInfo.Dialogues = make([]subparser.OneDialogue, 0)
	// 这里需要统计一共有几个 \N，以及这个数量在整体行数中的比例，这样就知道是不是双语字幕了
	countLineFeed := 0
	// 有意义的对话统计数，排除 Style 类型
	usefullDialogueCount := 0
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
	// 先读取一次字幕文件
	for _, oneLine := range matched {
		// 排除特效内容，只统计有意义的对话部分
		if strings.Contains(oneLine[0], mapByValue[0].Name) == false {
			continue
		}
		usefullDialogueCount++

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
		// nowText 优先移除 \h 这个是替换空格， \h 是让两个词在一行，不换行显示
		nowText = strings.ReplaceAll(nowText, `\h`, " ")
		// nowText 这个需要先把 {} 花括号内的内容给移除
		nowText1 := sub_parser.ReMatchBrace.ReplaceAllString(nowText, "")
		nowText1 = strings.TrimRight(nowText1, "\r")
		// 然后判断是否有 \N 或者 \n
		// 直接把 \n 替换为 \N 来解析
		nowText1 = strings.ReplaceAll(nowText1, `\n`, `\N`)
		if strings.Contains(nowText1, `\N`) {
			// 有，那么就需要再次切割，一般是双语字幕
			for _, matched2 := range sub_parser.ReCutDoubleLanguage.FindAllStringSubmatch(nowText1, -1) {
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

		subFileInfo.Dialogues = append(subFileInfo.Dialogues, odl)
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
	for _, dialogue := range subFileInfo.Dialogues {
		language.DetectSubLangAndStatistics(dialogue, langDict, &usefulDialogueExs, &chLines, &otherLines)
	}
	// 从统计出来的字典，找出 Top 1 或者 2 的出来，然后计算出是什么语言的字幕
	detectLang := language.SubLangStatistics2SubLangType(float32(countLineFeed), float32(usefullDialogueCount), langDict, chLines)
	subFileInfo.Lang = detectLang
	subFileInfo.Data = inBytes
	subFileInfo.DialoguesEx = usefulDialogueExs
	subFileInfo.CHLines = chLines
	subFileInfo.OtherLines = otherLines
	return true, &subFileInfo, nil
}

const (
	// 匹配 ass 文件中的 Style 变量
	regString4Style = `(?m)^Style:\s*(\w+),`
)

type StyleNameInfo struct {
	Name  string
	Count int
}
type StyleNameInfos []StyleNameInfo

func (a StyleNameInfos) Len() int           { return len(a) }
func (a StyleNameInfos) Less(i, j int) bool { return a[i].Count < a[j].Count }
func (a StyleNameInfos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func sortMapByValue(m map[string]int) StyleNameInfos {
	p := make(StyleNameInfos, len(m))
	i := 0
	for k, v := range m {
		p[i] = StyleNameInfo{k, v}
		i++
	}
	sort.Sort(sort.Reverse(p))
	return p
}
