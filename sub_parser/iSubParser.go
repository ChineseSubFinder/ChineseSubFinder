package sub_parser

import "github.com/allanpk716/ChineseSubFinder/common"

type ISubParser interface {
	DetermineFileType(filePath string) (common.Language, *SubFileInfo, error)
}