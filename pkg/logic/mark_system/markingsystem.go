package mark_system

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
	"github.com/sirupsen/logrus"
)

// MarkingSystem 评价系统，解决字幕排序优先级问题
type MarkingSystem struct {
	log             *logrus.Logger
	subSiteSequence []string // 网站的优先级，从高到低
	SubTypePriority int      // 字幕格式的优先级
	subParserHub    *sub_parser_hub.SubParserHub
}

func NewMarkingSystem(log *logrus.Logger, subSiteSequence []string, subTypePriority int) *MarkingSystem {
	mk := MarkingSystem{subSiteSequence: subSiteSequence,
		log:             log,
		SubTypePriority: subTypePriority,
		subParserHub:    sub_parser_hub.NewSubParserHub(log, ass.NewParser(log), srt.NewParser(log))}
	return &mk
}

// SelectOneSubFile 选择最优的一个字幕文件
func (m MarkingSystem) SelectOneSubFile(organizeSubFiles []string) *subparser.FileInfo {
	var finalSubFile *subparser.FileInfo
	subInfoDict := m.parseSubFileInfo(organizeSubFiles)
	// 优先级别暂定 subSiteSequence: zimuku -> subhd -> xunlei -> shooter
	// 这里需要循环四轮：
	// 第一轮，双语、字幕类型自定义，优先
	// 第二轮，单语言（中文）、字幕类型自定义，优先
	// 第三轮，双语、字幕类型0，优先
	// 第四轮，单语言（中文）、字幕类型0，优先
	for i := 0; i < 4; i++ {
		for _, subSite := range m.subSiteSequence {
			infos, ok := subInfoDict[subSite]
			if ok == false {
				continue
			}
			if i == 0 {
				finalSubFile = sub_helper.SelectChineseBestBilingualSubtitle(infos, m.SubTypePriority)
			} else if i == 1 {
				finalSubFile = sub_helper.SelectChineseBestSubtitle(infos, m.SubTypePriority)
			} else if i == 2 {
				finalSubFile = sub_helper.SelectChineseBestBilingualSubtitle(infos, 0)
			} else if i == 3 {
				finalSubFile = sub_helper.SelectChineseBestSubtitle(infos, 0)
			}
			if finalSubFile != nil {
				return finalSubFile
			}
		}
	}
	return nil
}

// SelectEachSiteTop1SubFile 每个网站最优的文件
func (m MarkingSystem) SelectEachSiteTop1SubFile(organizeSubFiles []string) ([]string, []subparser.FileInfo) {
	// 每个文件都带有出处 [subhd]
	var finalSubFile *subparser.FileInfo
	var outSiteName = make([]string, 0)
	var outSubParserFileInfos = make([]subparser.FileInfo, 0)
	subInfoDict := m.parseSubFileInfo(organizeSubFiles)
	// 这里需要循环四轮：
	// 第一轮，双语、字幕类型自定义，优先
	// 第二轮，单语言（中文）、字幕类型自定义，优先
	// 第三轮，双语、字幕类型0，优先
	// 第四轮，单语言（中文）、字幕类型0，优先
	for siteName, infos := range subInfoDict {
		// 每个网站保存一个
		for i := 0; i < 4; i++ {
			if i == 0 {
				finalSubFile = sub_helper.SelectChineseBestBilingualSubtitle(infos, m.SubTypePriority)
			} else if i == 1 {
				finalSubFile = sub_helper.SelectChineseBestSubtitle(infos, m.SubTypePriority)
			} else if i == 2 {
				finalSubFile = sub_helper.SelectChineseBestBilingualSubtitle(infos, 0)
			} else if i == 3 {
				finalSubFile = sub_helper.SelectChineseBestSubtitle(infos, 0)
			}
			if finalSubFile != nil {
				outSiteName = append(outSiteName, siteName)
				outSubParserFileInfos = append(outSubParserFileInfos, *finalSubFile)
				break
			}
		}
	}

	return outSiteName, outSubParserFileInfos
}

// parseSubFileInfo 从文件解析字幕信息
func (m MarkingSystem) parseSubFileInfo(organizeSubFiles []string) map[string][]subparser.FileInfo {
	// 一个网站可能就算取了 Top1 字幕，也可能是返回一个压缩包，然后解压完就是多个字幕，所以
	var subInfoDict = make(map[string][]subparser.FileInfo)
	// 拿到现有的字幕列表，开始抉择
	// 先判断当前字幕是什么语言（如果是简体，还需要考虑，判断这个字幕是简体还是繁体）
	for _, oneSubFileFullPath := range organizeSubFiles {
		bFind, subFileInfo, err := m.subParserHub.DetermineFileTypeFromFile(oneSubFileFullPath)
		if err != nil {
			m.log.Errorln("DetermineFileTypeFromFile", oneSubFileFullPath, err)
			continue
		}
		if bFind == false {
			m.log.Warnln("DetermineFileTypeFromFile", oneSubFileFullPath, "not support SubType")
			continue
		}
		_, ok := subInfoDict[subFileInfo.FromWhereSite]
		if ok == false {
			// 新建
			subInfoDict[subFileInfo.FromWhereSite] = make([]subparser.FileInfo, 0)
		}
		// 添加
		subInfoDict[subFileInfo.FromWhereSite] = append(subInfoDict[subFileInfo.FromWhereSite], *subFileInfo)
	}
	return subInfoDict
}
