package sub_helper

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
)

// DialogueMerger 合并分散的对白，目标是搞定英文字幕
type DialogueMerger struct {
	dialogueMap   map[string]*subparser.OneDialogueEx
	dialogueIndex []string
	lastStartTime string
}

func NewDialogueMerger() *DialogueMerger {
	return &DialogueMerger{
		dialogueMap:   make(map[string]*subparser.OneDialogueEx, 0),
		dialogueIndex: make([]string, 0),
		lastStartTime: "",
	}
}

func (d *DialogueMerger) Add(inDialogueEx subparser.OneDialogueEx) bool {

	// 第一个首字母是否是大写
	isUpper := isFirstLetterIsEngUpper(inDialogueEx.EnLine)
	isLower := isFirstLetterIsEngLower(inDialogueEx.EnLine)
	if isUpper == true {
		// 大写就新增
		d.dialogueMap[inDialogueEx.StartTime] = &inDialogueEx
		d.lastStartTime = inDialogueEx.StartTime
		d.dialogueIndex = append(d.dialogueIndex, inDialogueEx.StartTime)
		return true
	} else if isLower == true {
		// 小写就跟上一条的大写进行匹配，看是否能够附加到后面
		if d.lastStartTime == "" {
			return false
		}
		// 这里除了拼接 EnLine，还需要把 offsetEndTime 更新
		d.dialogueMap[d.lastStartTime].EnLine += " " + inDialogueEx.EnLine
		d.dialogueMap[d.lastStartTime].EndTime = inDialogueEx.EndTime
		//d.lastStartTime = ""
		return true
	} else {
		// 其他情况也新增
		d.dialogueMap[inDialogueEx.StartTime] = &inDialogueEx
		d.dialogueIndex = append(d.dialogueIndex, inDialogueEx.StartTime)
	}

	return false
}

func (d *DialogueMerger) Clear() {
	d.lastStartTime = ""
}

func (d *DialogueMerger) Get() []subparser.OneDialogueEx {
	var outDialogueExList = make([]subparser.OneDialogueEx, 0)
	for _, startString := range d.dialogueIndex {
		outDialogueExList = append(outDialogueExList, *d.dialogueMap[startString])
	}

	return outDialogueExList
}

// isFirstLetterIsEngUpper 字符开头的是英文大写的字幕
func isFirstLetterIsEngUpper(instring string) bool {

	if len(instring) <= 0 {
		return false
	}

	if 64 < instring[0] && instring[0] < 91 {
		return true
	}

	return false
}

// isFirstLetterIsEngLower 字符开头的是英文小写的字幕
func isFirstLetterIsEngLower(instring string) bool {

	if len(instring) <= 0 {
		return false
	}

	if 96 < instring[0] && instring[0] < 123 {
		return true
	}

	return false
}
