package srt

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_parser"
)

type Parser struct {

}

func (a Parser) DetermineFileType(filePath string) (common.Language, *sub_parser.SubFileInfo, error) {
	panic("implement me")
}

const regString = `(\d+)\n([\d:,]+)\s+-{2}\>\s+([\d:,]+)\n([\s\S]*?(\n{2}|$))`
