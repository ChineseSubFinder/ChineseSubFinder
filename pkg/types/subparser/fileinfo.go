package subparser

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
)

type FileInfo struct {
	PrefixDialogueString string              // 在 Dialogue: 这个关键词之前的字符串，ass 中的字体以及其他信息的描述
	Content              string              // 字幕的内容
	FromWhereSite        string              // 从那个网站下载的
	Name                 string              // 字幕的名称，注意，这里需要额外的赋值，不会自动检测
	Ext                  string              // 字幕的后缀名
	Lang                 language.MyLanguage // 识别出来的语言
	FileFullPath         string              // 字幕文件的全路径
	Data                 []byte              // 字幕的二进制文件内容
	Dialogues            []OneDialogue       // 整个字幕文件的所有对话，如果是做时间轴匹配，就使用原始的
	DialoguesFilter      []OneDialogue       // 整个字幕文件的所有对话，过滤掉特殊字符的对白
	DialoguesFilterEx    []OneDialogueEx     // 整个字幕文件的所有对话，过滤掉特殊字符的对白，这里会把一句话中支持的 中、英、韩、日 四国语言给分离出来
	CHLines              []string            // 抽取出所有的中文对话
	OtherLines           []string            // 抽取出所有的第二语言对话，可能是英文、韩文、日文
}

// SaveTranslated 保存字幕文件，注意，这里是用于翻译后的字幕文件
func (f *FileInfo) SaveTranslated(desSubFileFPath string) error {

	allString := ""
	allString += f.PrefixDialogueString + "\n"
	for _, oneDialogue := range f.Dialogues {

		if len(oneDialogue.Lines) < 1 {
			continue
		}
		oneDialogueString := "Dialogue: 0," + oneDialogue.StartTime + "," + oneDialogue.EndTime + ",Default,,0,0,0,," + oneDialogue.Lines[0]
		allString += oneDialogueString + "\n"
	}

	return pkg.WriteFile(desSubFileFPath, []byte(allString))
}

// GetSourceTranslateString 获取翻以前的字符串，会移除 \N 这样的信息，替换为空格
func (f *FileInfo) GetSourceTranslateString() string {
	sourceString := ""

	// 去除每一句中的 \N
	for index, oneDialogue := range f.Dialogues {

		f.Dialogues[index].Lines[0] = strings.ReplaceAll(oneDialogue.Lines[0], `\N`, " ")

		sourceString += f.Dialogues[index].Lines[0] + "\n"
	}

	return sourceString
}

func (f *FileInfo) SetTranslatedStrings(translatedString string) error {

	// 分行
	lines := strings.Split(translatedString, "\n")
	linesWithOutEmpty := make([]string, 0)
	// 移除空行
	for _, line := range lines {
		if len(line) < 1 {
			continue
		}
		linesWithOutEmpty = append(linesWithOutEmpty, line)
	}
	// 比较两个数组是否长度一致
	if len(f.Dialogues) != len(linesWithOutEmpty) {
		return fmt.Errorf("dialogue line not the same，org：%d，translated：%d", len(f.Dialogues), len(linesWithOutEmpty))
	}
	// 对每一句话进行赋值
	for index := range f.Dialogues {
		f.Dialogues[index].Lines = []string{linesWithOutEmpty[index]}
	}

	return nil
}

// SortDialogues 排序对话，时间递减
func (f *FileInfo) SortDialogues() {
	sort.Sort(OneDialogueByStartTime(f.Dialogues))
	sort.Sort(OneDialogueByStartTime(f.DialoguesFilter))
	sort.Sort(OneDialogueByStartTimeEx(f.DialoguesFilterEx))
}

// GetTimeFormat 获取时间轴的格式化格式
func (f FileInfo) GetTimeFormat() string {
	if f.Ext == common.SubExtASS || f.Ext == common.SubExtSSA {
		return common.TimeFormatPoint2
	} else {
		return common.TimeFormatPoint3
	}
}

// GetDialogueExContent 获取当前字幕文件语言对应索引的对白内容
// 凡是带有 Eng 的返回 Eng，其他的就与对应语言相关
func (f FileInfo) GetDialogueExContent(index int) string {

	switch f.Lang {
	case language.ChineseSimple, language.ChineseTraditional,
		language.ChineseSimpleJapanese, language.ChineseSimpleKorean,
		language.ChineseTraditionalJapanese, language.ChineseTraditionalKorean:
		// 带有中文的，但是又不是中英的
		return f.DialoguesFilterEx[index].ChLine
	case language.English, language.ChineseSimpleEnglish, language.ChineseTraditionalEnglish:
		return f.DialoguesFilterEx[index].EnLine
	case language.Japanese:
		return f.DialoguesFilterEx[index].JpLine
	case language.Korean:
		return f.DialoguesFilterEx[index].KrLine
	default:
		return f.DialoguesFilterEx[index].EnLine
	}
}

// ChangeDialoguesTimeByFramerateRatio 根据帧数比率调整时间轴 对应 ffsubsync -- SubtitleScaler
func (f *FileInfo) ChangeDialoguesTimeByFramerateRatio(framerateRatio float64) error {

	timeFormat := f.GetTimeFormat()
	f.changeOneDialoguesFramerateRatio(f.Dialogues, framerateRatio, timeFormat)
	f.changeOneDialoguesFramerateRatio(f.DialoguesFilter, framerateRatio, timeFormat)
	f.changeOneDialogueExsFramerateRatio(f.DialoguesFilterEx, framerateRatio, timeFormat)

	return nil
}

func (f *FileInfo) changeOneDialoguesFramerateRatio(oneDialogues []OneDialogue, framerateRatio float64, timeFormat string) {
	for i := 0; i < len(oneDialogues); i++ {

		timeStart := oneDialogues[i].GetStartTime()
		timeEnd := oneDialogues[i].GetEndTime()
		timeStartNumber := pkg.Time2SecondNumber(timeStart)
		timeEndNumber := pkg.Time2SecondNumber(timeEnd)

		scaleTimeStart := pkg.TimeNumber2Time(timeStartNumber * framerateRatio)
		scaleTimeEnd := pkg.TimeNumber2Time(timeEndNumber * framerateRatio)

		oneDialogues[i].StartTime = pkg.Time2SubTimeString(scaleTimeStart, timeFormat)
		oneDialogues[i].EndTime = pkg.Time2SubTimeString(scaleTimeEnd, timeFormat)
	}
}

func (f *FileInfo) changeOneDialogueExsFramerateRatio(oneDialogues []OneDialogueEx, framerateRatio float64, timeFormat string) {
	for i := 0; i < len(oneDialogues); i++ {

		timeStart := oneDialogues[i].GetStartTime()
		timeEnd := oneDialogues[i].GetEndTime()
		timeStartNumber := pkg.Time2SecondNumber(timeStart)
		timeEndNumber := pkg.Time2SecondNumber(timeEnd)

		scaleTimeStart := pkg.TimeNumber2Time(timeStartNumber * framerateRatio)
		scaleTimeEnd := pkg.TimeNumber2Time(timeEndNumber * framerateRatio)

		oneDialogues[i].StartTime = pkg.Time2SubTimeString(scaleTimeStart, timeFormat)
		oneDialogues[i].EndTime = pkg.Time2SubTimeString(scaleTimeEnd, timeFormat)
	}
}

// GetStartTime 获取的是从 Dialogues 得到的
func (f FileInfo) GetStartTime() time.Time {
	startTime := math.MaxFloat64
	for i := 0; i < len(f.Dialogues); i++ {
		// 找到最小的开始时间
		tmpNowStartTimeNumber := pkg.Time2SecondNumber(f.Dialogues[i].GetStartTime())
		startTime = math.Min(startTime, tmpNowStartTimeNumber)
	}
	return pkg.TimeNumber2Time(startTime)
}

// GetEndTime 获取的是从 Dialogues 得到的
func (f FileInfo) GetEndTime() time.Time {
	endTime := -math.MaxFloat64
	for i := 0; i < len(f.Dialogues); i++ {
		// 找到最大的结束时间
		tmpNowEndTimeNumber := pkg.Time2SecondNumber(f.Dialogues[i].GetEndTime())
		endTime = math.Max(endTime, tmpNowEndTimeNumber)
	}
	return pkg.TimeNumber2Time(endTime)
}

// GetNumFrames 获取这个字幕的时间 Frame 数量
func (f FileInfo) GetNumFrames() int {

	return int(math.Abs((pkg.Time2SecondNumber(f.GetEndTime()) - pkg.Time2SecondNumber(f.GetStartTime())) * 100))
}

func (f FileInfo) GetFileSha256() string {
	return fmt.Sprintf("%x", sha256.Sum256(f.Data))
}

// OneDialogue 一句对话
type OneDialogue struct {
	Index     int      // 对白的索引
	StartTime string   // 开始时间
	EndTime   string   // 结束时间
	StyleName string   // StyleName
	Lines     []string // 台词
}

func NewOneDialogue() OneDialogue {
	return OneDialogue{
		Lines: make([]string, 0),
	}
}

func (o OneDialogue) GetStartTime() time.Time {
	srcTimeStartNow, err := pkg.ParseTime(o.StartTime)
	if err != nil {
		return time.Time{}
	}
	return srcTimeStartNow
}

func (o OneDialogue) GetEndTime() time.Time {
	srcTimeEndNow, err := pkg.ParseTime(o.EndTime)
	if err != nil {
		return time.Time{}
	}
	return srcTimeEndNow
}

type OneDialogueByStartTime []OneDialogue

func (d OneDialogueByStartTime) Len() int {
	return len(d)
}

func (d OneDialogueByStartTime) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d OneDialogueByStartTime) Less(i, j int) bool {

	subStartTimeI, err := pkg.ParseTime(d[i].StartTime)
	if err != nil {
		return false
	}
	subStartTimeJ, err := pkg.ParseTime(d[j].StartTime)
	if err != nil {
		return false
	}
	return pkg.Time2SecondNumber(subStartTimeI) < pkg.Time2SecondNumber(subStartTimeJ)
}

// OneDialogueEx 一句对话，这里会把一句话中支持的 中、英、韩、日 四国语言给分离出来
type OneDialogueEx struct {
	StartTime string // 开始时间
	EndTime   string // 结束时间
	ChLine    string
	EnLine    string
	KrLine    string
	JpLine    string
}

func (o OneDialogueEx) GetStartTime() time.Time {
	srcTimeStartNow, err := pkg.ParseTime(o.StartTime)
	if err != nil {
		return time.Time{}
	}
	return srcTimeStartNow
}

func (o OneDialogueEx) GetEndTime() time.Time {
	srcTimeEndNow, err := pkg.ParseTime(o.EndTime)
	if err != nil {
		return time.Time{}
	}
	return srcTimeEndNow
}

type OneDialogueByStartTimeEx []OneDialogueEx

func (d OneDialogueByStartTimeEx) Len() int {
	return len(d)
}

func (d OneDialogueByStartTimeEx) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d OneDialogueByStartTimeEx) Less(i, j int) bool {

	subStartTimeI, err := pkg.ParseTime(d[i].StartTime)
	if err != nil {
		return false
	}
	subStartTimeJ, err := pkg.ParseTime(d[j].StartTime)
	if err != nil {
		return false
	}
	return pkg.Time2SecondNumber(subStartTimeI) < pkg.Time2SecondNumber(subStartTimeJ)
}

const (
	Sub_Ext_Mark_Default = ".default" // 指定这个字幕是默认的
	Sub_Ext_Mark_Forced  = ".forced"  // 指定这个字幕是强制的
)
