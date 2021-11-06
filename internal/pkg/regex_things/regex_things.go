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
var RegMatchSpString = regexp.MustCompile(`(?i)[^\w\s]`)

// 字幕文件对话的每一行
// regStringASS = `Dialogue: [^,.]*[0-9]*,([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),[^,.]*,[^,.]*,[0-9]*,[0-9]*,[0-9]*,[^,.]*,(.*)`
const regStringASS = `Dialogue: [^,.]*[0-9]*,([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),([^,.]*),[^,.]*,[0-9]*,[0-9]*,[0-9]*,[^,.]*,(.*)`
const regStringSRT = `(\d+)\n([\d:,]+)\s+-{2}\>\s+([\d:,]+)\n([\s\S]*?(\n{2}|$))`

var ReMatchDialogueSRT = regexp.MustCompile(regStringSRT)
var ReMatchDialogueASS = regexp.MustCompile(regStringASS)
