package same_as_video_name

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	language2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type Formatter struct {
	log       *logrus.Logger
	subParser *sub_parser_hub.SubParserHub
}

func NewFormatter(log *logrus.Logger) *Formatter {
	return &Formatter{log: log, subParser: sub_parser_hub.NewSubParserHub(log, ass.NewParser(log), srt.NewParser(log))}
}

// GetFormatterName 当前的 Formatter 是那个
func (f Formatter) GetFormatterName() string {
	return common.FormatterNameString_SampleAsVideoName
}

func (f Formatter) GetFormatterFormatterName() int {
	return int(common.SameAsVideoName)
}

// IsMatchThisFormat 是否满足当前实现接口的字幕命名格式 - 是否符合规则、fileNameWithOutExt string, subExt string, subLang types.MyLanguage, extraSubPreName string
func (f Formatter) IsMatchThisFormat(subName string) (bool, string, string, language2.MyLanguage, string) {
	/*
		这里要判断的是跟视频文件名称一样的字幕文件命名格式:
		The Boss Baby Family Business (2021) WEBDL-1080p.mp4
		对应字幕：
		The Boss Baby Family Business (2021) WEBDL-1080p.ass
	*/
	subNameBase := filepath.Base(subName)
	subNameDir := filepath.Dir(subName)
	// 这个情况下，字幕只可能有一个 . 符号存在，如果没有或者有多个，则认为不属于
	if strings.Contains(subNameBase, ".") == false {
		return false, "", "", language2.Unknown, ""
	}
	if strings.Count(subNameBase, ".") > 1 {
		return false, "", "", language2.Unknown, ""
	}
	// 获取文件的后缀名
	subExt := filepath.Ext(subNameBase)
	fileNameWithOutExt := strings.ReplaceAll(subNameBase, subExt, "")

	if pkg.IsFile(subName) == true {
		bok, fInfo, err := f.subParser.DetermineFileTypeFromFile(subName)
		if err != nil {
			return false, "", "", language2.Unknown, ""
		}
		if bok == true {
			return true, filepath.Join(subNameDir, fileNameWithOutExt), subExt, fInfo.Lang, ""
		}
	}

	return true, filepath.Join(subNameDir, fileNameWithOutExt), subExt, language2.Unknown, ""
}

// GenerateMixSubName 通过视频和字幕信息，生成当前实现接口的字幕命名格式。extraSubPreName 一般是填写字幕网站，不填写则留空 - 新名称、新名称带有 default 标记，新名称带有 forced 标记
func (f Formatter) GenerateMixSubName(videoFileName, subExt string, subLang language2.MyLanguage, extraSubPreName string) (string, string, string) {
	/*
		这里会生成类似的文件名 xxxx.zh
	*/
	videoFileNameWithOutExt := strings.ReplaceAll(filepath.Base(videoFileName),
		filepath.Ext(videoFileName), "")
	return f.GenerateMixSubNameBase(videoFileNameWithOutExt, subExt, subLang, extraSubPreName)
}

func (f Formatter) GenerateMixSubNameBase(fileNameWithOutExt, subExt string, subLang language2.MyLanguage, extraSubPreName string) (string, string, string) {
	// 这里传入字幕后缀名的时候，可能会带有 default 或者 forced 字段，需要剔除
	nowSubExt := strings.ReplaceAll(subExt, subparser.Sub_Ext_Mark_Default, "")
	nowSubExt = strings.ReplaceAll(nowSubExt, subparser.Sub_Ext_Mark_Forced, "")

	subNewName := fileNameWithOutExt + nowSubExt

	return subNewName, subNewName, subNewName
}
