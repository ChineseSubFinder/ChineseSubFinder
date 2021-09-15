package normal

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path/filepath"
	"regexp"
	"strings"
)

type Formatter struct {
	subParser *sub_helper.SubParserHub
}

func NewFormatter() *Formatter {
	return &Formatter{subParser: sub_helper.NewSubParserHub(ass.NewParser(), srt.NewParser())}
}

// GetFormatterName 当前的 Formatter 是那个
func (f Formatter) GetFormatterName() string {
	return common.FormatterNameString_Normal
}

func (f Formatter) GetFormatterFormatterName() int {
	return int(common.Normal)
}

// IsMatchThisFormat 是否满足当前实现接口的字幕命名格式 - 是否符合规则、fileNameWithOutExt string, subExt string, subLang types.Language, extraSubPreName string
func (f Formatter) IsMatchThisFormat(subName string) (bool, string, string, types.Language, string) {
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
		return false, "", "", types.Unknow, ""
	}
	var subLang types.Language
	var extraSubPreName string
	fileNameWithOutExt := strings.ReplaceAll(subName, matched[0][0], "")
	subExt := matched[0][2]
	//var subLangStr = matched[0][1]
	extraSubPreName = ""
	// 这里有一个点，是直接从 zh zho ch 去转换成中文语言就行了，还是要做字幕的语言识别
	// 目前倾向于这里用后面的逻辑
	//subLang = language.ChineseISOString2Lang(subLangStr)
	file, err := f.subParser.DetermineFileTypeFromFile(subName)
	if err != nil {
		return false, "", "", 0, ""
	}
	subLang = file.Lang
	return true, fileNameWithOutExt, subExt, subLang, extraSubPreName
}

// GenerateMixSubName 通过视频和字幕信息，生成当前实现接口的字幕命名格式。extraSubPreName 一般是填写字幕网站，不填写则留空 - 新名称、新名称带有 default 标记，新名称带有 forced 标记
func (f Formatter) GenerateMixSubName(videoFileName, subExt string, subLang types.Language, extraSubPreName string) (string, string, string) {
	/*
		这里会生成类似的文件名 xxxx.zh
	*/
	videoFileNameWithOutExt := strings.ReplaceAll(filepath.Base(videoFileName),
		filepath.Ext(videoFileName), "")
	return f.GenerateMixSubNameBase(videoFileNameWithOutExt, subExt, subLang, extraSubPreName)
}

func (f Formatter) GenerateMixSubNameBase(fileNameWithOutExt, subExt string, subLang types.Language, extraSubPreName string) (string, string, string) {

	subNewName := fileNameWithOutExt + "." + types.ChineseAbbr_639_1 + subExt
	subNewNameWithDefault := fileNameWithOutExt + "." + types.ChineseAbbr_639_1 + types.Sub_Ext_Mark_Default + subExt
	subNewNameWithForced := fileNameWithOutExt + "." + types.ChineseAbbr_639_1 + types.Sub_Ext_Mark_Forced + subExt

	return subNewName, subNewNameWithDefault, subNewNameWithForced
}
