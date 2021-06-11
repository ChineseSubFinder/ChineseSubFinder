package InterFace

import "github.com/allanpk716/ChineseSubFinder/sub_parser"

type ISubParser interface {

	DetermineFileTypeFromFile(filePath string) (*sub_parser.SubFileInfo, error)

	DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (*sub_parser.SubFileInfo, error)
}