package mark_system

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/srt"
	"github.com/sirupsen/logrus"
)

// MarkingSystem 评价系统，解决字幕排序优先级问题
type MarkingSystem struct {
	log *logrus.Logger
	subSiteSequence []string			// 网站的优先级，从高到低
	SubTypePriority int					// 字幕格式的优先级
	subParserHub *model.SubParserHub
}

func NewMarkingSystem(subSiteSequence []string, subTypePriority int) *MarkingSystem {
	mk := MarkingSystem{subSiteSequence: subSiteSequence,
		log: model.GetLogger(),
		SubTypePriority: subTypePriority,
		subParserHub: model.NewSubParserHub(ass.NewParser(), srt.NewParser())}
	return &mk
}

// SelectOneSubFile 选择最优的一个字幕文件
func (m MarkingSystem) SelectOneSubFile(organizeSubFiles []string) *common.SubParserFileInfo {
	var finalSubFile *common.SubParserFileInfo
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
				finalSubFile = model.SelectChineseBestBilingualSubtitle(infos, m.SubTypePriority)
			} else if i == 1 {
				finalSubFile = model.SelectChineseBestSubtitle(infos, m.SubTypePriority)
			} else if i == 2 {
				finalSubFile = model.SelectChineseBestBilingualSubtitle(infos, 0)
			} else if i == 3 {
				finalSubFile = model.SelectChineseBestSubtitle(infos, 0)
			}
			if finalSubFile != nil {
				return finalSubFile
			}
		}
	}
	return nil
}

// SelectEachSiteTop1SubFile 每个网站最优的文件
func (m MarkingSystem) SelectEachSiteTop1SubFile(organizeSubFiles []string) ([]string, []common.SubParserFileInfo) {
	// 每个文件都带有出处 [subhd]
	var finalSubFile *common.SubParserFileInfo
	var outSiteName = make([]string, 0)
	var outSubParserFileInfos = make([]common.SubParserFileInfo, 0)
	subInfoDict := m.parseSubFileInfo(organizeSubFiles)
	// 这里需要循环四轮：
	// 第一轮，双语、字幕类型自定义，优先
	// 第二轮，单语言（中文）、字幕类型自定义，优先
	// 第三轮，双语、字幕类型0，优先
	// 第四轮，单语言（中文）、字幕类型0，优先
	for i := 0; i < 4; i++ {
		for siteName, infos := range subInfoDict {
			// 每个网站保存一个
			if i == 0 {
				finalSubFile = model.SelectChineseBestBilingualSubtitle(infos, m.SubTypePriority)
			} else if i == 1 {
				finalSubFile = model.SelectChineseBestSubtitle(infos, m.SubTypePriority)
			} else if i == 2 {
				finalSubFile = model.SelectChineseBestBilingualSubtitle(infos, 0)
			} else if i == 3 {
				finalSubFile = model.SelectChineseBestSubtitle(infos, 0)
			}
			if finalSubFile != nil {
				outSiteName = append(outSiteName, siteName)
				outSubParserFileInfos = append(outSubParserFileInfos, *finalSubFile)
			}
		}
	}

	return outSiteName, outSubParserFileInfos
}

// parseSubFileInfo 从文件解析字幕信息
func (m MarkingSystem) parseSubFileInfo(organizeSubFiles []string) map[string][]common.SubParserFileInfo {
	// 一个网站可能就算取了 Top1 字幕，也可能是返回一个压缩包，然后解压完就是多个字幕，所以
	var subInfoDict = make(map[string][]common.SubParserFileInfo)
	// 拿到现有的字幕列表，开始抉择
	// 先判断当前字幕是什么语言（如果是简体，还需要考虑，判断这个字幕是简体还是繁体）
	for _, oneSubFileFullPath := range organizeSubFiles {
		subFileInfo, err := m.subParserHub.DetermineFileTypeFromFile(oneSubFileFullPath)
		if err != nil {
			m.log.Error("DetermineFileTypeFromFile", oneSubFileFullPath, err)
			continue
		}
		if subFileInfo == nil {
			// 说明这个字幕无法解析
			m.log.Warnln("MarkingSystem.parseSubFileInfo", oneSubFileFullPath, "DetermineFileTypeFromFile is nill")
			continue
		}

		_, ok := subInfoDict[subFileInfo.FromWhereSite]
		if ok == false {
			// 新建
			subInfoDict[subFileInfo.FromWhereSite] = make([]common.SubParserFileInfo, 0)
		}
		// 添加
		subInfoDict[subFileInfo.FromWhereSite] = append(subInfoDict[subFileInfo.FromWhereSite], *subFileInfo)
	}
	return subInfoDict
}