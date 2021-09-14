package emby

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
	return "emby formatter"
}

// IsMatchThisFormat 是否满足当前实现接口的字幕命名格式 - 是否符合规则、subExt string, subLang types.Language, extraSubPreName string
func (f Formatter) IsMatchThisFormat(subName string) (bool, string, types.Language, string) {
	/*
		Emby 的命名规则比较特殊，而且本程序就是做中文字幕下载的，所以，下面的正则表达式比较特殊
	*/
	var re = regexp.MustCompile(`(?m).chinese\((\S+)\)(\.\S+)`)
	matched := re.FindAllStringSubmatch(subName, -1)
	/*
		[0][0]	.chinese(简英,subhd).ass
		[0][1]	简英,subhd or 简英
		[0][2]	.ass
	*/
	if len(matched) < 1 || len(matched[0]) < 3 {
		return false, "", types.ChineseSimple, ""
	}
	var subLang types.Language
	var subLangStr string
	var extraSubPreName string
	subExt := matched[0][2]
	midString := matched[0][1]
	if strings.Contains(midString, ",") == true {
		tmps := strings.Split(midString, ",")
		if len(tmps) < 2 {
			return false, "", types.ChineseSimple, ""
		}
		subLangStr = tmps[0]
		extraSubPreName = tmps[1]
	} else {
		subLangStr = midString
		extraSubPreName = ""
	}
	subLang = language.ChineseString2Lang(subLangStr)

	return true, subExt, subLang, extraSubPreName
}

// GenerateMixSubName 通过视频和字幕信息，生成当前实现接口的字幕命名格式。extraSubPreName 一般是填写字幕网站，不填写则留空 - 新名称、新名称带有 default 标记，新名称带有 forced 标记
func (f Formatter) GenerateMixSubName(videoFileName, subExt string, subLang types.Language, extraSubPreName string) (string, string, string) {
	/*
		这里会生成类似的文件名 xxxx.chinese(中英,shooter)
	*/
	videoFileNameWithOutExt := strings.ReplaceAll(filepath.Base(videoFileName),
		filepath.Ext(videoFileName), "")
	note := ""
	// extraSubPreName 那个字幕网站下载的
	if extraSubPreName != "" {
		note = "," + extraSubPreName
	}
	const defaultString = ".default"
	const forcedString = ".forced"
	const chineseString = ".chinese"

	subNewName := videoFileNameWithOutExt + chineseString + "(" + language.Lang2ChineseString(subLang) + note + ")" + subExt
	subNewNameWithDefault := videoFileNameWithOutExt + chineseString + "(" + language.Lang2ChineseString(subLang) + note + ")" + defaultString + subExt
	subNewNameWithForced := videoFileNameWithOutExt + chineseString + "(" + language.Lang2ChineseString(subLang) + note + ")" + forcedString + subExt

	return subNewName, subNewNameWithDefault, subNewNameWithForced
}
