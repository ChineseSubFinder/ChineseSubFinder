package subparser

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"time"
)

type FileInfo struct {
	Content       string              // 字幕的内容
	FromWhereSite string              // 从那个网站下载的
	Name          string              // 字幕的名称，注意，这里需要额外的赋值，不会自动检测
	Ext           string              // 字幕的后缀名
	Lang          language.MyLanguage // 识别出来的语言
	FileFullPath  string              // 字幕文件的全路径
	Data          []byte              // 字幕的二进制文件内容
	Dialogues     []OneDialogue       // 整个字幕文件的所有对话
	DialoguesEx   []OneDialogueEx     // 整个字幕文件的所有对话，这里会把一句话中支持的 中、英、韩、日 四国语言给分离出来
	CHLines       []string            // 抽取出所有的中文对话
	OtherLines    []string            // 抽取出所有的第二语言对话，可能是英文、韩文、日文
}

// GetTimeFormat 获取时间轴的格式化格式
func (f FileInfo) GetTimeFormat() string {
	if f.Ext == common.SubExtASS || f.Ext == common.SubExtSSA {
		return common.TimeFormatPoint2
	} else {
		return common.TimeFormatPoint3
	}
}

// ParseTime 解析字幕时间字符串，这里可能小数点后面有 2-4 位
func (f FileInfo) ParseTime(inTime string) (time.Time, error) {

	parseTime, err := time.Parse(common.TimeFormatPoint2, inTime)
	if err != nil {
		parseTime, err = time.Parse(common.TimeFormatPoint3, inTime)
		if err != nil {
			parseTime, err = time.Parse(common.TimeFormatPoint4, inTime)
		}
	}
	return parseTime, err
}

// GetDialogueExContent 获取当前字幕文件语言对应索引的对白内容
// 凡是带有 Eng 的返回 Eng，其他的就与对应语言相关
func (f FileInfo) GetDialogueExContent(index int) string {

	switch f.Lang {
	case language.ChineseSimple, language.ChineseTraditional,
		language.ChineseSimpleJapanese, language.ChineseSimpleKorean,
		language.ChineseTraditionalJapanese, language.ChineseTraditionalKorean:
		// 带有中文的，但是又不是中英的
		return f.DialoguesEx[index].ChLine
	case language.English, language.ChineseSimpleEnglish, language.ChineseTraditionalEnglish:
		return f.DialoguesEx[index].EnLine
	case language.Japanese:
		return f.DialoguesEx[index].JpLine
	case language.Korean:
		return f.DialoguesEx[index].KrLine
	default:
		return f.DialoguesEx[index].EnLine
	}
}

// OneDialogue 一句对话
type OneDialogue struct {
	StartTime string   // 开始时间
	EndTime   string   // 结束时间
	StyleName string   // StyleName
	Lines     []string // 台词
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

const (
	Sub_Ext_Mark_Default = ".default" // 指定这个字幕是默认的
	Sub_Ext_Mark_Forced  = ".forced"  // 指定这个字幕是强制的
)
