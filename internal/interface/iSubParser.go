package _interface

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
)

type ISubParser interface {

	GetParserName() string

	DetermineFileTypeFromFile(filePath string) (*subparser.FileInfo, error)

	DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (*subparser.FileInfo, error)
}