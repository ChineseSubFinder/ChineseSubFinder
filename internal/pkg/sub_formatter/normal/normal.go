package normal

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path/filepath"
	"regexp"
	"strings"
)

type Formatter struct {
}

func NewFormatter() *Formatter {
	return &Formatter{}
}

// GetFormatterName 当前的 Formatter 是那个
func (f Formatter) GetFormatterName() string {
	return "normal formatter"
}

// IsMatchThisFormat 是否满足当前实现接口的字幕命名格式 - 是否符合规则、subExt string, subLang types.Language, extraSubPreName string
func (f Formatter) IsMatchThisFormat(subName string) (bool, string, types.Language, string) {
	/*
		Emby 的命名规则比较特殊，而且本程序就是做中文字幕下载的，所以，下面的正则表达式比较特殊
	*/
	var re = regexp.MustCompile(`(?m)\.(\bzh\b|\bzho\b|\bchi\b)(\.\S+)`)
	matched := re.FindAllStringSubmatch(subName, -1)
	/*
		The Boss Baby Family Business (2021) WEBDL-1080p.zh.ass
		The Boss Baby Family Business (2021) WEBDL-1080p.zh.default.ass
		The Boss Baby Family Business (2021) WEBDL-1080p.zh.forced.ass
		The Boss Baby Family Business (2021) WEBDL-1080p.chi.ass
		The Boss Baby Family Business (2021) WEBDL-1080p.chi.default.ass
		The Boss Baby Family Business (2021) WEBDL-1080p.chi.forced.ass
		The Boss Baby Family Business (2021) WEBDL-1080p.zho.ass
		The Boss Baby Family Business (2021) WEBDL-1080p.zho.default.ass
		The Boss Baby Family Business (2021) WEBDL-1080p.zho.forced.ass

		[0][0]	.zh.ass
		[0][1]	zh
		[0][2]	.ass
	*/
	if matched == nil || len(matched) < 1 || len(matched[0]) < 3 {
		return false, "", types.Unknow, ""
	}
	var subLang types.Language
	var subLangStr string
	var extraSubPreName string
	subExt := matched[0][2]
	subLangStr = matched[0][1]
	extraSubPreName = ""
	subLang = language.ChineseISOString2Lang(subLangStr)

	return true, subExt, subLang, extraSubPreName
}

// GenerateMixSubName 通过视频和字幕信息，生成当前实现接口的字幕命名格式。extraSubPreName 一般是填写字幕网站，不填写则留空 - 新名称、新名称带有 default 标记，新名称带有 forced 标记
func (f Formatter) GenerateMixSubName(videoFileName, subExt string, subLang types.Language, extraSubPreName string) (string, string, string) {
	/*
		这里会生成类似的文件名 xxxx.zh
	*/
	videoFileNameWithOutExt := strings.ReplaceAll(filepath.Base(videoFileName),
		filepath.Ext(videoFileName), "")

	subNewName := videoFileNameWithOutExt + "." + types.ChineseAbbr_639_1 + subExt
	subNewNameWithDefault := videoFileNameWithOutExt + "." + types.ChineseAbbr_639_1 + types.Sub_Ext_Mark_Default + subExt
	subNewNameWithForced := videoFileNameWithOutExt + "." + types.ChineseAbbr_639_1 + types.Sub_Ext_Mark_Forced + subExt

	return subNewName, subNewNameWithDefault, subNewNameWithForced
}
