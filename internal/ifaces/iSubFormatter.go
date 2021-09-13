package ifaces

import "github.com/allanpk716/ChineseSubFinder/internal/types"

type ISubFormatter interface {
	IsMatchThisFormat(subName string) bool

	GenerateMixSubName(videoFileName, subExt string, subLang types.Language, extraSubPreName string) (string, string, string)
}
