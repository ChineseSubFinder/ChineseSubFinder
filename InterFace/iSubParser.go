package InterFace

import (
	"github.com/allanpk716/ChineseSubFinder/common"
)

type ISubParser interface {

	DetermineFileTypeFromFile(filePath string) (*common.SubFileInfo, error)

	DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (*common.SubFileInfo, error)
}