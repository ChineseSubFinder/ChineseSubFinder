package srt

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/regex_things"
	"github.com/sirupsen/logrus"
)

type Parser struct {
	log *logrus.Logger
}

func NewParser(log *logrus.Logger) *Parser {
	return &Parser{log: log}
}

func (p Parser) GetParserName() string {
	return "srt"
}

/*
	DetermineFileTypeFromFile 确定字幕文件的类型，是双语字幕或者某一种语言等等信息
	当 error 是 common.DetermineFileTypeFromFileExtNotFitSRT
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

	subFileInfo := subparser.FileInfo{}
	subFileInfo.Content = string(inBytes)
	subFileInfo.Ext = nowExt
	subFileInfo.Dialogues = make([]subparser.OneDialogue, 0)
	subFileInfo.DialoguesFilter = make([]subparser.OneDialogue, 0)

	orgDialogues := p.parseContent(inBytes)
	if len(orgDialogues) <= 0 {
		if p.log != nil {
			p.log.Debugln("DetermineFileTypeFromBytes can't found DialoguesFilter, Skip")
		}
		return false, nil, nil
	}
	subFileInfo.Dialogues = orgDialogues
	// 这里需要统计一共有几个 \N，以及这个数量在整体行数中的比例，这样就知道是不是双语字幕了
	countLineFeed := 0
	for _, oneDialogue := range orgDialogues {

		if len(oneDialogue.Lines) == 0 || pkg.ReplaceSpecString(oneDialogue.Lines[0], "") == "" {
			continue
		}
		ol := oneDialogue
		for i, line := range oneDialogue.Lines {
			fixedLine := line

			// 剔除 {\fn微软雅黑\fs14}C'mon, Rick. We're -- We're almost there. {} 这一段
			fixedLine = regex_things.ReMatchBrace.ReplaceAllString(line, "")
			fixedLine = regex_things.ReMatchBracket.ReplaceAllString(fixedLine, "")
			fixedLine = strings.ReplaceAll(fixedLine, `\N`, "")
			if pkg.ReplaceSpecString(fixedLine, "") == "" {
				ol.Lines[i] = ""
				break
			} else {
				if i == 1 {
					// 这样说明有两行字幕，也就是双语啦
					countLineFeed++
				}
				ol.Lines[i] = fixedLine
			}
		}
		if ol.Lines[0] == "" {
			continue
		}
		subFileInfo.DialoguesFilter = append(subFileInfo.DialoguesFilter, ol)
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
	emptyLines := 0
	for _, dialogue := range subFileInfo.DialoguesFilter {
		emptyLines += language.DetectSubLangAndStatistics(dialogue, langDict, &usefulDialogueExs, &chLines, &otherLines)
	}
	// 从统计出来的字典，找出 Top 1 或者 2 的出来，然后计算出是什么语言的字幕
	detectLang := language.SubLangStatistics2SubLangType(float32(countLineFeed), float32(len(subFileInfo.DialoguesFilter)-emptyLines), langDict, chLines)
	subFileInfo.Lang = detectLang
	subFileInfo.Data = inBytes
	subFileInfo.DialoguesFilterEx = usefulDialogueExs
	subFileInfo.CHLines = chLines
	subFileInfo.OtherLines = otherLines
	return true, &subFileInfo, nil
}

func (p Parser) parseContent(inBytes []byte) []subparser.OneDialogue {

	allString := string(inBytes)
	// 注意，需要替换掉 \r 不然正则表达式会有问题
	allString = strings.ReplaceAll(allString, "\r", "")

	lines := strings.Split(allString, "\n")
	// 需要把每一行如果是多余的特殊剔除掉
	// 这里的目标是后续的匹配更加容易，但是，后续也得注意
	// 因为这个样的操作，那么匹配对白内容的时候，可能是不存在的，只要是 index 和 时间匹配上了，就应该算一句话，只要在 dialogue 上是没得问题的
	// 而 dialogueFilter 中则可以把这样没有内容的排除，但是实际时间轴匹配的时候还是用 dialogue 而不是 dialogueFilter
	filterLines := make([]string, 0)
	for _, line := range lines {
		// 如果当前的这一句话，为空，或者进过正则表达式剔除特殊字符后为空，则跳过
		if pkg.ReplaceSpecString(line, "") == "" {
			continue
		}
		filterLines = append(filterLines, line)
	}

	dialogues := make([]subparser.OneDialogue, 0)
	/*
		这里可以确定，srt 格式，开始一定是第一句话，那么首先就需要找到，第一行，一定是数字的，从这里开始算起
		1. 先将 content 进行 \r 的替换为空
		2. 将 content 进行 \n 来分割
		3. 将分割的数组进行筛选，把空行剔除掉
		4. 然后使用循环，用下面的 steps 进行解析一句对白
		steps:
				0	找对白的 ID
				1	找时间轴
				2	找对白内容，可能有多行，停止的方式，一个是向后能找到 0以及2 或者 是最后一行
	*/
	steps := 0
	nowDialogue := subparser.NewOneDialogue()
	newOneDialogueFun := func() {
		// 重新新建一个缓存对白，从新开始
		steps = 0
		nowDialogue = subparser.NewOneDialogue()
	}
	// 使用过滤后的列表
	for i, line := range filterLines {

		if steps == 0 {
			// 匹配对白的索引
			line = pkg.ReplaceSpecString(line, "")
			dialogueIndex, err := strconv.Atoi(line)
			if err != nil {
				newOneDialogueFun()
				continue
			}
			nowDialogue.Index = dialogueIndex
			// 继续
			steps = 1
			continue
		}

		if steps == 1 {
			// 匹配时间
			matched := regex_things.ReMatchDialogueTimeSRT.FindAllStringSubmatch(line, -1)
			if matched == nil || len(matched) < 1 || matched[0][0] != line {
				matched = regex_things.ReMatchDialogueTimeSRT2.FindAllStringSubmatch(line, -1)
				if matched == nil || len(matched) < 1 || matched[0][0] != line {
					newOneDialogueFun()
					continue
				}
			}
			nowDialogue.StartTime = matched[0][1]
			nowDialogue.EndTime = matched[0][2]

			// 是否到结尾
			if i+1 > len(filterLines)-1 {
				// 是尾部
				// 那么这一个对白就需要 add 到总列表中了
				dialogues = append(dialogues, nowDialogue)
				newOneDialogueFun()
				continue
			}
			// 如上面提到的，因为把特殊字符的行去除了，那么一个对话，如果只有 index 和 时间，也是需要添加进去的
			if p.needMatchNextContentLine(filterLines, i+1) == true {
				// 是，那么也认为当前这个对话完成了，需要 add 到总列表中了
				dialogues = append(dialogues, nowDialogue)
				newOneDialogueFun()
				continue
			}
			// 非上述特殊情况，继续
			steps = 2
			continue
		}

		if steps == 2 {
			// 在上述情况排除后，才继续
			// 匹配内容

			if len(regex_things.ReMatchSrtSubtitleEffects.FindAllString(line, -1)) > 5 {
				continue
			}

			nowDialogue.Lines = append(nowDialogue.Lines, line)
			// 是否到结尾
			if i+1 > len(filterLines)-1 {
				// 是尾部
				// 那么这一个对白就需要 add 到总列表中了
				dialogues = append(dialogues, nowDialogue)
				newOneDialogueFun()
				continue
			}

			// 不是尾部，那么就需要往后看两句话，是否是下一个对白的头部（index 和 时间）
			if p.needMatchNextContentLine(filterLines, i+1) == true {
				// 是，那么也认为当前这个对话完成了，需要 add 到总列表中了
				dialogues = append(dialogues, nowDialogue)
				newOneDialogueFun()
				continue
			} else {
				// 如果还不是，那么就可能是这个对白有多行，有可能是同一种语言的多行，也可能是多语言的多行
				// 那么 step 应该不变继续是 2
				continue
			}
		}
	}

	return dialogues
}

// needMatchNextContentLine 是否需要继续匹配下一句话作为一个对白的对话内容
func (p Parser) needMatchNextContentLine(lines []string, index int) bool {

	if index+1 > len(lines)-1 {
		return false
	}

	// 匹配到对白的 Index
	_, err := strconv.Atoi(lines[index])
	if err != nil {
		return false
	}
	// 匹配到字幕的时间
	matched := regex_things.ReMatchDialogueTimeSRT.FindAllStringSubmatch(lines[index+1], -1)
	if matched == nil || len(matched) < 1 {
		matched = regex_things.ReMatchDialogueTimeSRT2.FindAllStringSubmatch(lines[index+1], -1)
		if matched == nil || len(matched) < 1 {
			return false
		}
	}

	return true
}
