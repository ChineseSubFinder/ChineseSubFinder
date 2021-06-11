package common

import (
	"path/filepath"
	"regexp"
)

type SubParserHub struct {
	Parser []ISubParser
}

// NewSubParserHub 处理的字幕文件需要符合 [siteName]_ 的前缀描述，是本程序专用的
func NewSubParserHub(parser ISubParser, _parser ...ISubParser) *SubParserHub {
	s := SubParserHub{}
	s.Parser = make([]ISubParser, 0)
	s.Parser = append(s.Parser, parser)
	if len(_parser) > 0 {
		for _, one := range _parser {
			s.Parser = append(s.Parser, one)
		}
	}
	return &s
}

// DetermineFileTypeFromFile 确定字幕文件的类型，是双语字幕或者某一种语言等等信息，如果返回 nil ，那么就说明都没有字幕的格式匹配上
func (p SubParserHub) DetermineFileTypeFromFile(filePath string) (*SubParserFileInfo, error){
	for _, parser := range p.Parser {
		subFileInfo, err := parser.DetermineFileTypeFromFile(filePath)
		if err != nil {
			return nil, err
		}
		// 文件的格式不匹配解析器就是 nil
		if subFileInfo == nil {
			continue
		} else {
			// 正常至少应该匹配一个吧，不然就是最外层继续返回 nil 出去了
			// 简体和繁体字幕的判断，通过文件名来做到的，基本就算个补判而已
			newLang := IsChineseSimpleOrTraditional(filePath, subFileInfo.Lang)
			subFileInfo.Name = filepath.Base(filePath)
			subFileInfo.Lang = newLang
			subFileInfo.FileFullPath = filePath
			subFileInfo.FromWhereSite = p.getFromWhereSite(filePath)
			return subFileInfo, nil
		}
	}
	// 如果返回 nil ，那么就说明都没有字幕的格式匹配上
	return nil, nil
}
// getFromWhereSite 从文件名找出是从那个网站下载的。这里的文件名的前缀是下载时候标记好的，比较特殊
func (p SubParserHub) getFromWhereSite(filePath string) string {
	fileName := filepath.Base(filePath)
	var re = regexp.MustCompile(`^\[(\w+)\]_`)
	matched := re.FindStringSubmatch(fileName)
	if len(matched) < 1 {
		return ""
	}
	return matched[1]
}
