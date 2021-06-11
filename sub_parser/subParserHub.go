package sub_parser

import "github.com/allanpk716/ChineseSubFinder/common"

type SubParserHub struct {
	Parser []ISubParser
}

func NewSubParserHub(parser ISubParser, _inparser ... ISubParser) *SubParserHub {
	s := SubParserHub{}
	s.Parser = make([]ISubParser, 0)
	s.Parser = append(s.Parser, parser)
	if len(_inparser) > 0 {
		for _, one := range _inparser {
			s.Parser = append(s.Parser, one)
		}
	}
	return &s
}

// DetermineFileTypeFromFile 确定字幕文件的类型，是双语字幕或者某一种语言等等信息，如果返回 nil ，那么就说明都没有字幕的格式匹配上
func (p SubParserHub) DetermineFileTypeFromFile(filePath string) (*SubFileInfo, error){
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
			return subFileInfo, nil
		}
	}
	// 如果返回 nil ，那么就说明都没有字幕的格式匹配上
	return nil, nil
}

type SubFileInfo struct {
	Name	string			// 字幕的名称，注意，这里需要额外的赋值，不会自动检测
	Ext		string			// 字幕的后缀名
	Lang common.Language	// 识别出来的语言
	Dialogues []OneDialogue	// 整个字幕文件的所有对话
}

// OneDialogue 一句对话
type OneDialogue struct {
	StartTime string		// 开始时间
	EndTime string			// 结束时间
	Lines	[]string		// 台词
}