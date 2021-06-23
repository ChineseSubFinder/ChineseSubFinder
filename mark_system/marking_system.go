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
	subParserHub *model.SubParserHub
}
// TODO 在这里添加字幕格式选择的逻辑

func NewMarkingSystem(subSiteSequence []string) *MarkingSystem {
	mk := MarkingSystem{subSiteSequence: subSiteSequence,
		log: model.GetLogger(),
		subParserHub: model.NewSubParserHub(ass.NewParser(), srt.NewParser())}
	return &mk
}

// SelectOneSubFile 选择最优的一个字幕文件
func (m MarkingSystem) SelectOneSubFile(organizeSubFiles []string) *common.SubParserFileInfo {
	var finalSubFile common.SubParserFileInfo
	subInfoDict := m.parseSubFileInfo(organizeSubFiles)
	// 优先级别暂定 subSiteSequence: zimuku -> subhd -> xunlei -> shooter
	for _, subSite := range m.subSiteSequence {
		value, ok := subInfoDict[subSite]
		if ok == false {
			continue
		}
		info := model.FindChineseBestSubtitle(value)
		if info != nil {
			finalSubFile = *info
			return &finalSubFile
		}
	}
	return nil
}

// SelectEachSiteTop1SubFile 每个网站最优的文件
func (m MarkingSystem) SelectEachSiteTop1SubFile(organizeSubFiles []string) ([]string, []common.SubParserFileInfo) {
	// 每个文件都带有出处 [subhd]
	var outSiteName = make([]string, 0)
	var outSubParserFileInfos = make([]common.SubParserFileInfo, 0)
	subInfoDict := m.parseSubFileInfo(organizeSubFiles)
	for siteName, infos := range subInfoDict {
		// 每个网站保存一个
		info := model.FindChineseBestSubtitle(infos)
		if info != nil {
			outSiteName = append(outSiteName, siteName)
			outSubParserFileInfos = append(outSubParserFileInfos, *info)
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
			m.log.Error(err)
			continue
		}
		if subFileInfo == nil {
			// 说明这个字幕无法解析
			m.log.Warnln(oneSubFileFullPath, "DetermineFileTypeFromFile is nill")
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