package normal

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	language2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/sirupsen/logrus"
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
	return common.FormatterNameString_Normal
}

func (f Formatter) GetFormatterFormatterName() int {
	return int(common.Normal)
}

// IsMatchThisFormat 是否满足当前实现接口的字幕命名格式 - 是否符合规则、fileNameWithOutExt string, subExt string, subLang types.MyLanguage, extraSubPreName string
func (f Formatter) IsMatchThisFormat(subName string) (bool, string, string, language2.MyLanguage, string) {
	/*
		Emby 的命名规则比较特殊，而且本程序就是做中文字幕下载的，所以，下面的正则表达式比较特殊
		见本程序内 internal/types/language/ISOLanguage.go 这里的支持 ISO 规范和中文编码变种
		见文档、讨论：
		https://emby.media/community/index.php?/topic/94504-current-chinese-subtitle-filter-not-so-accurate-and-hope-improve-like-this/
		https://en.wikipedia.org/wiki/Chinese_Wikipedia#Automatic_conversion_between_traditional_and_simplified_Chinese_characters
		https://stackoverflow.com/questions/18902072/what-standard-do-language-codes-of-the-form-zh-hans-belong-to
	*/
	//subName = strings.ToLower(subName)

	// get basename to avoid relative path like "../../.." cause issue for regexp
	// CANT just get Base, as autoDetectAndChange expect a full path
	subNameBase := filepath.Base(subName)
	subNameDir := filepath.Dir(subName)
	var re = regexp.MustCompile(language.ISOSupportRegexRule())
	matched := re.FindAllStringSubmatch(subNameBase, -1)
	/*
		详细看测试用例
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
		return false, "", "", language2.Unknown, ""
	}
	var subLang language2.MyLanguage
	var extraSubPreName string

	// replace only applys to basename
	fileNameWithOutExt := strings.ReplaceAll(subNameBase, matched[0][0], "")
	subExt := matched[0][2]
	var subLangStr = matched[0][1]
	extraSubPreName = ""
	// 这里有一个点，是直接从 zh zho ch 去转换成中文语言就行了，还是要做字幕的语言识别
	// 目前倾向于这里用后面的逻辑
	subLang = language.ISOString2SupportLang(subLangStr)
	// 这里可能是拿到的是文件的全路径，那么就可以读取文件内容去判断文件的语言
	if pkg.IsFile(subName) == true {
		bok, fileInfo, err := f.subParser.DetermineFileTypeFromFile(subName)
		if err != nil || bok == false {
			// add original Dir to fileNameWithOutExt to ensure file can be reached
			return true, filepath.Join(subNameDir, fileNameWithOutExt), subExt, subLang, extraSubPreName
		}
		subLang = fileInfo.Lang
	}
	// add original Dir to fileNameWithOutExt to ensure file can be reached
	return true, filepath.Join(subNameDir, fileNameWithOutExt), subExt, subLang, extraSubPreName
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

	subNewName := fileNameWithOutExt + "." + language2.ISO_639_1_Chinese + nowSubExt
	subNewNameWithDefault := fileNameWithOutExt + "." + language2.ISO_639_1_Chinese + subparser.Sub_Ext_Mark_Default + nowSubExt
	subNewNameWithForced := fileNameWithOutExt + "." + language2.ISO_639_1_Chinese + subparser.Sub_Ext_Mark_Forced + nowSubExt

	return subNewName, subNewNameWithDefault, subNewNameWithForced
}
