package _interface

import "github.com/allanpk716/ChineseSubFinder/common"

type ISubParser interface {

	DetermineFileTypeFromFile(filePath string) (*common.SubParserFileInfo, error)

	DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (*common.SubParserFileInfo, error)
}