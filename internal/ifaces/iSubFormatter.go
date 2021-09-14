package ifaces

import "github.com/allanpk716/ChineseSubFinder/internal/types"

type ISubFormatter interface {
	// GetFormatterName 当前的 Formatter 是那个
	GetFormatterName() string
	// IsMatchThisFormat 是否满足当前实现接口的字幕命名格式 - 是否符合规则、subExt string, subLang types.Language, extraSubPreName string
	IsMatchThisFormat(subName string) (bool, string, types.Language, string)
	// GenerateMixSubName 通过视频和字幕信息，生成当前实现接口的字幕命名格式。extraSubPreName 一般是填写字幕网站，不填写则留空 - 新名称、新名称带有 default 标记，新名称带有 forced 标记
	GenerateMixSubName(videoFileName, subExt string, subLang types.Language, extraSubPreName string) (string, string, string)
}
