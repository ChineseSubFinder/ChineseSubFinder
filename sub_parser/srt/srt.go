package srt

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type Parser struct {

}

func NewParser() *Parser {
	return &Parser{}
}

// DetermineFileTypeFromFile 确定字幕文件的类型，是双语字幕或者某一种语言等等信息
func (p Parser) DetermineFileTypeFromFile(filePath string) (*common.SubFileInfo, error) {
	nowExt := filepath.Ext(filePath)
	if strings.ToLower(nowExt) != common.SubExtSRT {
		return nil ,nil
	}
	fBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil ,err
	}

	return p.DetermineFileTypeFromBytes(fBytes, nowExt)
}

// DetermineFileTypeFromBytes 确定字幕文件的类型，是双语字幕或者某一种语言等等信息
func (p Parser) DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (*common.SubFileInfo, error){

	allString := string(inBytes)
	// 注意，需要替换掉 \r 不然正则表达式会有问题
	allString = strings.ReplaceAll(allString, "\r", "")
	re := regexp.MustCompile(regString)
	// 找到 start end text
	matched := re.FindAllStringSubmatch(allString, -1)
	if len(matched) < 1 {
		return nil ,nil
	}
	subFileInfo := common.SubFileInfo{}
	subFileInfo.Ext = nowExt
	subFileInfo.Dialogues = make([]common.OneDialogue, 0)
	// 这里需要统计一共有几个 \N，以及这个数量在整体行数中的比例，这样就知道是不是双语字幕了
	countLineFeed := 0
	for _, oneDial := range matched {
		startTime := oneDial[2]
		endTime := oneDial[3]
		nowText := oneDial[4]
		odl := common.OneDialogue{
			StartTime: startTime,
			EndTime: endTime,
		}
		odl.Lines = make([]string, 0)
		nowText = strings.TrimRight(nowText, "\n")
		texts := strings.Split(nowText, "\n")
		for i, text := range texts {
			if i == 1 {
				// 这样说明有两行字幕，也就是双语啦
				countLineFeed++
			}
			odl.Lines = append(odl.Lines, text)
		}
		subFileInfo.Dialogues = append(subFileInfo.Dialogues, odl)
	}
	// 再分析
	// 需要判断每一个 Line 是啥语言，[语言的code]次数
	var langDict map[int]int
	langDict = make(map[int]int)
	for _, dialogue := range subFileInfo.Dialogues {
		common.DetectSubLangAndStatistics(dialogue.Lines, langDict)
	}
	// 从统计出来的字典，找出 Top 1 或者 2 的出来，然后计算出是什么语言的字幕
	detectLang := common.SubLangStatistics2SubLangType(float32(countLineFeed), float32(len(matched)), langDict)
	subFileInfo.Lang = detectLang
	subFileInfo.Data = inBytes
	return &subFileInfo, nil
}

const regString = `(\d+)\n([\d:,]+)\s+-{2}\>\s+([\d:,]+)\n([\s\S]*?(\n{2}|$))`
