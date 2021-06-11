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
	subSiteSequence []string	// 网站的优先级，从高到低
}

func NewMarkingSystem(subSiteSequence []string) *MarkingSystem {
	mk := MarkingSystem{subSiteSequence: subSiteSequence,
		log: model.GetLogger()}
	return &mk
}

func (m MarkingSystem) SelectOneSubFile(organizeSubFiles []string) *common.SubParserFileInfo {
	var finalSubFile common.SubParserFileInfo
	// TODO 这里先处理 Top1 的字幕，后续再考虑怎么觉得 Top N 选择哪一个，很可能选择每个网站 Top 1就行了，具体的过滤逻辑在其内部实现
	// 一个网站可能就算取了 Top1 字幕，也可能是返回一个压缩包，然后解压完就是多个字幕，所以
	var subInfoDict = make(map[string][]common.SubParserFileInfo)
	// 拿到现有的字幕列表，开始抉择
	// 先判断当前字幕是什么语言（如果是简体，还需要考虑，判断这个字幕是简体还是繁体）
	subParserHub := model.NewSubParserHub(ass.NewParser(), srt.NewParser())
	for _, oneSubFileFullPath := range organizeSubFiles {
		subFileInfo, err := subParserHub.DetermineFileTypeFromFile(oneSubFileFullPath)
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
	// 优先级别暂定 subSiteSequence: zimuku -> subhd -> xunlei -> shooter
	for _, subSite := range m.subSiteSequence {
		value, ok := subInfoDict[subSite]
		if ok == true {
			for _, info := range value {
				// 找到了中文字幕
				if model.HasChineseLang(info.Lang) == true {
					finalSubFile = info
					return &finalSubFile
				}
			}
		}
	}
	return nil
}