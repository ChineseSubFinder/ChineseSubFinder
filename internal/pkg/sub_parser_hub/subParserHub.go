package sub_parser_hub

import (
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/filter"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	languageConst "github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SubParserHub struct {
	log    *logrus.Logger
	Parser []ifaces.ISubParser
}

// NewSubParserHub 处理的字幕文件需要符合 [siteName]_ 的前缀描述，是本程序专用的
func NewSubParserHub(log *logrus.Logger, parser ifaces.ISubParser, _parser ...ifaces.ISubParser) *SubParserHub {
	s := SubParserHub{}
	s.log = log
	s.Parser = make([]ifaces.ISubParser, 0)
	s.Parser = append(s.Parser, parser)
	if len(_parser) > 0 {
		for _, one := range _parser {
			s.Parser = append(s.Parser, one)
		}
	}
	return &s
}

// DetermineFileTypeFromFile 确定字幕文件的类型，是双语字幕或者某一种语言等等信息，如果返回 nil ，那么就说明都没有字幕的格式匹配上
func (p SubParserHub) DetermineFileTypeFromFile(filePath string) (bool, *subparser.FileInfo, error) {
	for _, parser := range p.Parser {
		bFind, subFileInfo, err := parser.DetermineFileTypeFromFile(filePath)
		if err != nil {
			return false, nil, err
		}
		if bFind == false {
			continue
		}
		// 正常至少应该匹配一个吧，不然就是最外层继续返回 nil 出去了
		// 简体和繁体字幕的判断，通过文件名来做到的，基本就算个补判而已
		//newLang := IsChineseSimpleOrTraditional(filePath, subFileInfo.Lang)
		subFileInfo.Name = filepath.Base(filePath)
		//subFileInfo.Lang = newLang
		subFileInfo.FileFullPath = filePath
		subFileInfo.FromWhereSite = p.getFromWhereSite(filePath)
		return true, subFileInfo, nil
	}
	// 如果返回 nil ，那么就说明都没有字幕的格式匹配上
	return false, nil, nil
}

// DetermineFileTypeFromBytes 确定字幕文件的类型，是双语字幕或者某一种语言等等信息，如果返回 nil ，那么就说明都没有字幕的格式匹配上
// 如果要做字幕的时间轴匹配，很可能需要一个功能 sub_helper.MergeMultiDialogue4EngSubtitle，但是仅仅是合并了 English 字幕时间轴
func (p SubParserHub) DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (bool, *subparser.FileInfo, error) {

	for _, parser := range p.Parser {
		bFind, subFileInfo, err := parser.DetermineFileTypeFromBytes(inBytes, nowExt)
		if err != nil {
			return false, nil, err
		}
		if bFind == false {
			continue
		}
		return true, subFileInfo, nil
	}
	// 如果返回 nil ，那么就说明都没有字幕的格式匹配上
	return false, nil, nil
}

// IsSubHasChinese 字幕文件是否包含中文
func (p SubParserHub) IsSubHasChinese(fileInfo *subparser.FileInfo) bool {

	// 增加判断已存在的字幕是否有中文
	if language.HasChineseLang(fileInfo.Lang) == false {
		if p.log != nil {
			p.log.Warnln("IsSubHasChinese.HasChineseLang", fileInfo.FileFullPath, "not chinese sub, is ", fileInfo.Lang.String())
		}
		return false
	}

	return true
}

// getFromWhereSite 从文件名找出是从那个网站下载的。这里的文件名的前缀是下载时候标记好的，比较特殊
func (p SubParserHub) getFromWhereSite(filePath string) string {
	fileName := filepath.Base(filePath)
	var re = regexp.MustCompile(`^\[(\w+)\]_`)
	matched := re.FindStringSubmatch(fileName)
	if matched == nil || len(matched) < 1 {
		return ""
	}
	return matched[1]
}

// IsSubTypeWanted 这里匹配的字幕的格式，不包含 Ext 的 . 小数点，注意，仅仅是包含关系
func IsSubTypeWanted(subName string) bool {
	nowLowerName := strings.ToLower(subName)
	if strings.Contains(nowLowerName, common.SubTypeASS) ||
		strings.Contains(nowLowerName, common.SubTypeSSA) ||
		strings.Contains(nowLowerName, common.SubTypeSRT) {
		return true
	}

	return false
}

// IsSubExtWanted 输入的字幕文件名，判断后缀名是否符合期望的字幕后缀名列表
func IsSubExtWanted(subName string) bool {
	inExt := filepath.Ext(subName)
	switch strings.ToLower(inExt) {
	case common.SubExtSSA, common.SubExtASS, common.SubExtSRT:
		return true
	default:
		return false
	}
}

// IsEmbySubCodecWanted 从 Emby api 拿到字幕的 sub 类型 string (Codec) 是否是符合本程序要求的
func IsEmbySubCodecWanted(inSubCodec string) bool {

	tmpString := strings.ToLower(inSubCodec)
	if tmpString == common.SubTypeSRT ||
		tmpString == common.SubTypeASS ||
		tmpString == common.SubTypeSSA {
		return true
	}

	return false
}

// IsEmbySubChineseLangStringWanted 是否是 Emby 自己解析出来的中文语言类型
func IsEmbySubChineseLangStringWanted(inLangString string) bool {

	isWanted := false

	tmpString := strings.ToLower(inLangString)
	nextString := tmpString
	spStrings := strings.Split(tmpString, "[")
	if len(spStrings) > 1 {
		// 去除 chi[xunlie] 类似的标记
		nextString = spStrings[0]
	} else {
		// 去除 chinese（简英,zimuku）
		spStrings = strings.Split(tmpString, "(")
		if len(spStrings) > 1 {
			nextString = spStrings[0]
		}
	}

	// 先判断 ISO 标准的和变种的支持列表，仅仅是中文的
	if language.IsSupportISOChineseString(nextString) {
		// fmt.Println("###: ERROR")
		isWanted = true
	}

	// 再判断之前支持的列表
	switch nextString {
	case languageConst.Emby_chinese_chs,
		languageConst.Emby_chinese_cht,
		languageConst.Emby_chinese_chi:
		// chi chs cht
		isWanted = true
	case replaceLangString(languageConst.Emby_chinese):
		// chinese，这个比较特殊，是本程序定义的 chinese 的字段，再 Emby API 下特殊的字幕命名字段
		isWanted = true
	}

	return isWanted
}

// SearchMatchedSubFile 搜索符合后缀名的字幕文件
func SearchMatchedSubFile(log *logrus.Logger, dir string) ([]string, error) {

	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 内层的错误就无视了
			oneList, _ := SearchMatchedSubFile(log, fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if IsSubExtWanted(curFile.Name()) == false {
				continue
			} else {

				// 跳过不符合的文件，比如 MAC OS 下可能有缓存文件，见 #138
				if filter.SkipFileInfo(log, curFile) == true {
					continue
				}

				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

func replaceLangString(inString string) string {
	tmpString := strings.ToLower(inString)
	one := strings.ReplaceAll(tmpString, ".", "")
	two := strings.ReplaceAll(one, "_", "")
	return two
}
