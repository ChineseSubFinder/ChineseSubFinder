package ifaces

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
)

type ISubParser interface {
	GetParserName() string

	DetermineFileTypeFromFile(filePath string) (bool, *subparser.FileInfo, error)

	DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (bool, *subparser.FileInfo, error)
}
