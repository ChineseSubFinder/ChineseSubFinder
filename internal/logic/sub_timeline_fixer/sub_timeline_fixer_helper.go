package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	formatterEmby "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/normal"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/sub_timeline_fiexer"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type SubTimelineFixerHelper struct {
	embyHelper       *emby_helper.EmbyHelper
	EmbyConfig       emby.EmbyConfig
	FixerConfig      sub_timeline_fiexer.SubTimelineFixerConfig
	subParserHub     *sub_parser_hub.SubParserHub
	subTimelineFixer *sub_timeline_fixer.SubTimelineFixer
	formatter        map[string]ifaces.ISubFormatter
	threads          int
	timeOut          time.Duration
}

func NewSubTimelineFixerHelper(embyConfig emby.EmbyConfig, subTimelineFixerConfig sub_timeline_fiexer.SubTimelineFixerConfig) *SubTimelineFixerHelper {
	sub := SubTimelineFixerHelper{
		EmbyConfig:       embyConfig,
		FixerConfig:      subTimelineFixerConfig,
		embyHelper:       emby_helper.NewEmbyHelper(embyConfig),
		subParserHub:     sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser()),
		subTimelineFixer: sub_timeline_fixer.NewSubTimelineFixer(subTimelineFixerConfig),
		formatter:        make(map[string]ifaces.ISubFormatter),
		threads:          6,
		timeOut:          60 * time.Second,
	}
	// TODO 如果字幕格式新增了实现，这里也需要添加对应的实例
	// 初始化支持的 formatter
	// normal
	sub.formatter = make(map[string]ifaces.ISubFormatter)
	normalM := normal.NewFormatter()
	sub.formatter[normalM.GetFormatterName()] = normalM
	// emby
	embyM := formatterEmby.NewFormatter()
	sub.formatter[embyM.GetFormatterName()] = embyM

	return &sub
}

func (s SubTimelineFixerHelper) FixRecentlyItemsSubTimeline(movieRootDir, seriesRootDir string) error {

	// 首先得开启，不然就直接跳过不执行
	if s.EmbyConfig.FixTimeLine == false {
		log_helper.GetLogger().Debugf("EmbyConfig.FixTimeLine = false, Skip")
		return nil
	}

	movieList, seriesList, err := s.embyHelper.GetRecentlyAddVideoList(movieRootDir, seriesRootDir)
	if err != nil {
		return err
	}

	// 输出调试信息
	log_helper.GetLogger().Debugln("FixRecentlyItemsSubTimeline - DebugInfo - movieList Start")
	for s, value := range movieList {
		log_helper.GetLogger().Debugln(s, value)
	}
	log_helper.GetLogger().Debugln("FixRecentlyItemsSubTimeline - DebugInfo - movieList End")

	log_helper.GetLogger().Debugln("FixRecentlyItemsSubTimeline - DebugInfo - seriesList Start")
	for s, _ := range seriesList {
		log_helper.GetLogger().Debugln(s)
	}
	log_helper.GetLogger().Debugln("FixRecentlyItemsSubTimeline - DebugInfo - seriesList End")

	log_helper.GetLogger().Debugln("Start movieList fix Timeline")
	// 先做电影的字幕校正、然后才是连续剧的
	for _, info := range movieList {
		// path.Dir 在 Windows 有梗，所以换个方式获取路径
		videoRootPath := filepath.Dir(info.VideoFileFullPath)
		err = s.fixOneVideoSub(info.VideoInfo.Id, videoRootPath)
		if err != nil {
			return err
		}
	}
	log_helper.GetLogger().Debugln("End movieList fix Timeline")

	log_helper.GetLogger().Debugln("Start seriesList fix Timeline")
	for _, infos := range seriesList {
		for _, info := range infos {
			// path.Dir 在 Windows 有梗，所以换个方式获取路径
			videoRootPath := filepath.Dir(info.VideoFileFullPath)
			err = s.fixOneVideoSub(info.VideoInfo.Id, videoRootPath)
			if err != nil {
				return err
			}
		}
	}
	log_helper.GetLogger().Debugln("End seriesList fix Timeline")

	// 强制调用，测试 CGO=1 编译问题
	log_helper.GetLogger().Debugln("VAD Mode", vad.Mode)

	return nil
}

func (s SubTimelineFixerHelper) fixOneVideoSub(videoId string, videoRootPath string) error {
	log_helper.GetLogger().Debugln("fixOneVideoSub VideoROotPath:", videoRootPath)
	// internalEngSub 默认第一个是 srt 然后第二个是 ass，就不要去遍历了
	found, internalEngSub, containChineseSubFile, err := s.embyHelper.GetInternalEngSubAndExChineseEnglishSub(videoId)
	if err != nil {
		return err
	}

	if found == false {
		log_helper.GetLogger().Debugln("GetInternalEngSubAndExChineseEnglishSub - found == false")
		return nil
	}

	log_helper.GetLogger().Debugln("internalEngSub:", len(internalEngSub), "containChineseSubFile:", len(containChineseSubFile))
	// 需要先把原有的外置字幕带有 -fix 的删除，然后再做修正
	// 不然如果调整了条件，之前修复的本次其实就不修正了，那么就会“残留”下来，误以为是本次配置的信息导致的
	for _, exSubInfo := range containChineseSubFile {
		// 没有编辑的就跳过
		if strings.Contains(exSubInfo.FileName, sub_timeline_fixer.FixMask) == false {
			continue
		}

		subFileNeedRemove := filepath.Join(videoRootPath, exSubInfo.FileName)

		if videoRootPath == "" {
			log_helper.GetLogger().Debugln("videoRootPath == \"\", Skip Remove:", subFileNeedRemove)
			continue
		}

		log_helper.GetLogger().Debugln("Remove fixed sub:", subFileNeedRemove)
		err = os.Remove(subFileNeedRemove)
		if err != nil {
			return err
		}
	}

	// 从外置双语（中英）字幕中找对对应的内置 srt 字幕进行匹配比较
	for _, exSubInfo := range containChineseSubFile {
		inSelectSubIndex := 1
		if exSubInfo.Ext == common.SubExtSRT {
			inSelectSubIndex = 0
		}
		// 修正过的字幕有标记，将不会再次修复
		if strings.Contains(exSubInfo.FileName, sub_timeline_fixer.FixMask) == true {
			continue
		}

		log_helper.GetLogger().Debugln("fixSubTimeline start")
		bFound, subFixInfos, subNewName, err := s.fixSubTimeline(internalEngSub[inSelectSubIndex], exSubInfo)
		if err != nil {
			return err
		}
		if bFound == false {
			log_helper.GetLogger().Debugln("fixSubTimeline bFound == false", exSubInfo.FileName)
			continue
		}
		// 调试的时候用
		if videoRootPath == "" {
			log_helper.GetLogger().Debugln("videoRootPath == \"\", Skip fix sub:", exSubInfo.FileName)
			continue
		}
		for _, info := range subFixInfos {
			// 写入 fix 后的字幕文件覆盖之前的字幕文件
			desFixedSubFullName := filepath.Join(videoRootPath, subNewName)
			err = s.saveSubFile(desFixedSubFullName, info.FixContent)
			if err != nil {
				return err
			}
			log_helper.GetLogger().Infoln("Sub Timeline fixed:", desFixedSubFullName)
		}
	}

	return nil
}

// fixSubTimeline 修复时间轴，containChineseSubFile 这里可能是，只要是带有中文的都算，简体、繁体、简英、繁英，需要后续额外的判断
func (s SubTimelineFixerHelper) fixSubTimeline(enSubFile emby.SubInfo, containChineseSubFile emby.SubInfo) (bool, []sub_timeline_fixer.SubFixInfo, string, error) {
	fixedSubName := ""
	log_helper.GetLogger().Debugln("fixSubTimeline - DetermineFileTypeFromBytes", enSubFile.FileName)
	bFind, infoBase, err := s.subParserHub.DetermineFileTypeFromBytes(enSubFile.Content, enSubFile.Ext)
	if err != nil {
		return false, nil, fixedSubName, err
	}
	if bFind == false {
		return false, nil, fixedSubName, nil
	}
	infoBase.Name = enSubFile.FileName
	/*
		这里发现一个梗，内置的英文字幕导出的时候，有可能需要合并多个 Dialogue，见
		internal/pkg/sub_helper/sub_helper.go 中 MergeMultiDialogue4EngSubtitle 的实现
	*/
	sub_helper.MergeMultiDialogue4EngSubtitle(infoBase)

	log_helper.GetLogger().Debugln("fixSubTimeline - DetermineFileTypeFromBytes", containChineseSubFile.FileName)
	bFind, infoSrc, err := s.subParserHub.DetermineFileTypeFromBytes(containChineseSubFile.Content, containChineseSubFile.Ext)
	if err != nil {
		return false, nil, fixedSubName, err
	}
	if bFind == false {
		return false, nil, fixedSubName, nil
	}
	infoSrc.Name = containChineseSubFile.FileName
	/*
		这里发现一个梗，内置的英文字幕导出的时候，有可能需要合并多个 Dialogue，见
		internal/pkg/sub_helper/sub_helper.go 中 MergeMultiDialogue4EngSubtitle 的实现
	*/
	sub_helper.MergeMultiDialogue4EngSubtitle(infoSrc)

	infoBaseNameWithOutExt := strings.Replace(infoBase.Name, path.Ext(infoBase.Name), "", -1)
	//infoSrcNameWithOutExt := strings.Replace(infoSrc.Name, path.Ext(infoSrc.Name), "", -1)

	// 把原始的文件缓存下来，新建缓存的文件夹
	subFixCacheRootPath, err := my_util.GetRootSubFixCacheFolder()
	if err != nil {
		return false, nil, fixedSubName, err
	}
	cacheTmpPath := filepath.Join(subFixCacheRootPath, infoBaseNameWithOutExt)
	if my_util.IsDir(cacheTmpPath) == false {
		err = os.MkdirAll(cacheTmpPath, os.ModePerm)
		if err != nil {
			return false, nil, fixedSubName, err
		}
	}
	// 写入内置字幕、外置字幕原始文件
	err = s.saveSubFile(filepath.Join(cacheTmpPath, infoBaseNameWithOutExt+".chinese(inside)"+infoBase.Ext), infoBase.Content)
	if err != nil {
		return false, nil, fixedSubName, err
	}
	err = s.saveSubFile(filepath.Join(cacheTmpPath, infoSrc.Name), infoSrc.Content)
	if err != nil {
		return false, nil, fixedSubName, err
	}
	bok, offsetTime, sd, err := s.subTimelineFixer.GetOffsetTimeV1(infoBase, infoSrc, filepath.Join(cacheTmpPath, infoSrc.Name+"-bar.html"), filepath.Join(cacheTmpPath, infoSrc.Name+".log"))
	if offsetTime != 0 {
		log_helper.GetLogger().Infoln(infoSrc.Name, "offset time is", fmt.Sprintf("%f", offsetTime), "s")
	}
	// 超过 SD 阈值了
	if sd > s.FixerConfig.V1_MaxStartTimeDiffSD {
		log_helper.GetLogger().Infoln(infoSrc.Name, "Start Time Diff SD, skip", fmt.Sprintf("%f", sd))
		return false, nil, fixedSubName, nil
	} else {
		log_helper.GetLogger().Infoln(infoSrc.Name, "Start Time Diff SD", fmt.Sprintf("%f", sd))
	}

	if err != nil || bok == false {
		return false, nil, fixedSubName, err
	}

	// 偏移很小就无视了
	if offsetTime < s.FixerConfig.V1_MinOffset && offsetTime > -s.FixerConfig.V1_MinOffset {
		log_helper.GetLogger().Infoln(infoSrc.Name, fmt.Sprintf("Min Offset Config is %f, skip ", s.FixerConfig.V1_MinOffset), fmt.Sprintf("now is %f", offsetTime))
		return false, nil, fixedSubName, nil
	}
	// 写入校准时间轴后的字幕
	var subFixInfos = make([]sub_timeline_fixer.SubFixInfo, 0)
	for _, formatter := range s.formatter {
		// 符合已知的字幕命名格式，不符合就跳过，都跳过也行，就不做任何操作而已
		bMatch, fileNameWithOutExt, subExt, subLang, extraSubName := formatter.IsMatchThisFormat(infoSrc.Name)
		if bMatch == false {
			log_helper.GetLogger().Debugln(fmt.Sprintf("%s IsMatchThisFormat == false, Skip, %s", formatter.GetFormatterName(), infoSrc.Name))
			continue
		}
		// 是否包含 default 关键词，暂时无需判断 forced
		hasDefault := false
		if strings.Contains(strings.ToLower(infoSrc.Name), subparser.Sub_Ext_Mark_Default) == true {
			hasDefault = true
		}
		// 生成对应字幕命名格式的，字幕命名。这里注意，normal 的时候， extraSubName+"-fix" 是无效的，不会被设置，也就是直接覆盖之前的字幕了。
		subNewName, subNewNameDefault, _ := formatter.GenerateMixSubNameBase(fileNameWithOutExt, subExt, subLang, extraSubName+sub_timeline_fixer.FixMask)

		desFixSubFileFullPath := ""
		if hasDefault == true {
			fixedSubName = subNewNameDefault
			desFixSubFileFullPath = filepath.Join(cacheTmpPath, subNewNameDefault)

		} else {
			fixedSubName = subNewName
			desFixSubFileFullPath = filepath.Join(cacheTmpPath, subNewName)
		}
		fixContent, err := s.subTimelineFixer.FixSubTimeline(infoSrc, offsetTime, desFixSubFileFullPath)
		if err != nil {
			return false, nil, fixedSubName, err
		}
		subFixInfos = append(subFixInfos, *sub_timeline_fixer.NewSubFixInfo(infoSrc.Name, fixContent))
	}

	return true, subFixInfos, fixedSubName, nil
}

func (s SubTimelineFixerHelper) saveSubFile(desSaveSubFileFullPath string, content string) error {
	dstFile, err := os.Create(desSaveSubFileFullPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()
	_, err = dstFile.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
