package srt

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_parser"
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

func (a Parser) DetermineFileType(filePath string) (*sub_parser.SubFileInfo, error) {
	nowExt := filepath.Ext(filePath)
	if strings.ToLower(nowExt) != common.SubExtSRT {
		return nil ,nil
	}
	fBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil ,err
	}

	allString := string(fBytes)
	// 注意，需要替换掉 \r 不然正则表达式会有问题
	allString = strings.ReplaceAll(allString, "\r", "")
	re := regexp.MustCompile(regString)
	// 找到 start end text
	matched := re.FindAllStringSubmatch(allString, -1)
	if len(matched) < 1 {
		return nil ,nil
	}
	subFileInfo := sub_parser.SubFileInfo{}
	subFileInfo.Ext = nowExt
	subFileInfo.Dialogues = make([]sub_parser.OneDialogue, 0)
	// 这里需要统计一共有几个 \N，以及这个数量在整体行数中的比例，这样就知道是不是双语字幕了
	countLineFeed := 0

	println(countLineFeed)
	return nil, nil
}

//const regString = `(?i)(\d+)\s+([\d:,]+)\s+-{2}\>\s+([\d:,]+)\s+([\s\S]*?(\s{2}|$))`
const regString = `(\d+)\n([\d:,]+)\s+-{2}\>\s+([\d:,]+)\n([\s\S]*?(\n{2}|$))`
