package regex_things

import "regexp"

const (
	// 匹配 ass 文件中的 Style 变量
	regString4Style = `(?m)^Style:\s*(\w+),`
)

// ReMatchBrace 匹配花括号中的内容
var ReMatchBrace = regexp.MustCompile(`(?m)((?i){[^}]*})`)

var ReMatchBracket = regexp.MustCompile(`(?m)((?i)\[[^]]*\])`)

var ReCutDoubleLanguage = regexp.MustCompile(`(?i)(.*)\\N(.*)`)

// RegMatchSpString 替换特殊字符
//var RegMatchSpString = regexp.MustCompile(`(?i)[^\w\s]`)
var RegMatchSpString = regexp.MustCompile(`(?m)[\p{P}|\p{Z}}}|\p{S}\s|\t|\v]`)

// 字幕文件对话的每一行
// regStringASS = `Dialogue: [^,.]*[0-9]*,([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),[^,.]*,[^,.]*,[0-9]*,[0-9]*,[0-9]*,[^,.]*,(.*)`
const regStringASS = `Dialogue: [^,.]*[0-9]*,([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),([^,.]*),[^,.]*,[0-9]*,[0-9]*,[0-9]*,[^,.]*,(.*)`
const regStringSRT = `(\d+)\n([\d:,]+)\s+-{2}\>\s+([\d:,]+)\n([\s\S]*?(\n{1,2}|$))`
const regStringSRT2 = `(\d+)\n([\d:.]+)\s+-{2}\>\s+([\d:.]+)\n([\s\S]*?(\n{1,2}|$))`

const regStringSRTime = `([\d:,]+)\s+-{2}\>\s+([\d:,]+)`
const regStringSRTime2 = `([\d:.]+)\s+-{2}\>\s+([\d:.]+)`

// 匹配 srt 的字幕特效，需要移除这些
var ReMatchSrtSubtitleEffects = regexp.MustCompile(`(?m)([1-9]\d*\.?\d*)|(0\.\d*[1-9])`)

var ReMatchDialogueASS = regexp.MustCompile(regStringASS)
var ReMatchDialogueSRT = regexp.MustCompile(regStringSRT)
var ReMatchDialogueSRT2 = regexp.MustCompile(regStringSRT2)
var ReMatchDialogueTimeSRT = regexp.MustCompile(regStringSRTime)
var ReMatchDialogueTimeSRT2 = regexp.MustCompile(regStringSRTime2)

// RegOneSeasonSubFolderNameMatch 每个视频文件夹下的缓存文件夹名称，一个季度的
var RegOneSeasonSubFolderNameMatch = regexp.MustCompile(`(?m)^Sub_S\dE0`)

const regStringMathLogOneLine = `(?m)^(\[)(.*)(\])\: (\d{4}-\d{1,2}-\d{1,2} \d{1,2}:\d{1,2}:\d{1,2}) - (.+)`

const regMatchIP = `(?m)((25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d))).){3}(25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d)))`

var ReMatchIP = regexp.MustCompile(regMatchIP)

// 匹配目前日志记录的格式的一行
var ReMathLogOneLine = regexp.MustCompile(regStringMathLogOneLine)
