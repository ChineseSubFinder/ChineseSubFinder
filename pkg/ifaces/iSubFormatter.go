package ifaces

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
)

// ISubFormatter 如果新增了更多的字幕命名格式化的实现，需要找所有以下 to do 去增加实现
// TODO 如果字幕格式新增了实现，这里也需要添加对应的实例
type ISubFormatter interface {
	// GetFormatterName 当前的 formatter 是那个
	GetFormatterName() string
	// GetFormatterFormatterName 需要转换为 FormatterName 使用
	GetFormatterFormatterName() int
	// IsMatchThisFormat 是否满足当前实现接口的字幕命名格式 - 是否符合规则、fileNameWithOutExt string, subExt string, subLang types.MyLanguage, extraSubPreName string
	IsMatchThisFormat(subName string) (bool, string, string, language.MyLanguage, string)
	// GenerateMixSubName 通过视频和字幕信息，生成当前实现接口的字幕命名格式。extraSubPreName 一般是填写字幕网站，不填写则留空 - 新名称、新名称带有 default 标记，新名称带有 forced 标记
	GenerateMixSubName(videoFileName, subExt string, subLang language.MyLanguage, extraSubPreName string) (string, string, string)
	// GenerateMixSubNameBase 通过没有后缀名信息的文件名，生成当前实现接口的字幕命名格式。extraSubPreName 一般是填写字幕网站，不填写则留空 - 新名称、新名称带有 default 标记，新名称带有 forced 标记
	GenerateMixSubNameBase(fileNameWithOutExt, subExt string, subLang language.MyLanguage, extraSubPreName string) (string, string, string)
}
