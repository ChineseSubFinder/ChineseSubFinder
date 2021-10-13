package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	formatterEmby "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/normal"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type SubTimelineFixerHelper struct {
	embyHelper   *emby_helper.EmbyHelper
	EmbyConfig   emby.EmbyConfig
	subParserHub *sub_parser_hub.SubParserHub
	formatter    map[string]ifaces.ISubFormatter
	threads      int
	timeOut      time.Duration
}

func NewSubTimelineFixerHelper(embyConfig emby.EmbyConfig) *SubTimelineFixerHelper {
	sub := SubTimelineFixerHelper{
		EmbyConfig:   embyConfig,
		embyHelper:   emby_helper.NewEmbyHelper(embyConfig),
		subParserHub: sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser()),
		formatter:    make(map[string]ifaces.ISubFormatter),
		threads:      6,
		timeOut:      60 * time.Second,
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
		return nil
	}

	movieList, seriesList, err := s.embyHelper.GetRecentlyAddVideoList(movieRootDir, seriesRootDir)
	if err != nil {
		return err
	}
	// 先做电影的字幕校正、然后才是连续剧的
	for _, info := range movieList {
		// path.Dir 在 Windows 有梗，所以换个方式获取路径
		videoRootPath := filepath.Dir(info.VideoFileFullPath)
		err = s.fixOneVideoSub(info.VideoInfo.Id, videoRootPath)
		if err != nil {
			return err
		}
	}
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

	return nil
}

func (s SubTimelineFixerHelper) fixOneVideoSub(videoId string, videoRootPath string) error {
	// internalEngSub 默认第一个是 srt 然后第二个是 ass，就不要去遍历了
	found, internalEngSub, exCh_EngSub, err := s.embyHelper.GetInternalEngSubAndExChineseEnglishSub(videoId)
	if err != nil {
		return err
	}
	if found == false {
		return nil
	}
	// 从外置双语（中英）字幕中找对对应的内置 srt 字幕进行匹配比较
	for _, exSubInfo := range exCh_EngSub {
		inSelectSubIndex := 1
		if exSubInfo.Ext == common.SubExtSRT {
			inSelectSubIndex = 0
		}

		bFound, subFixInfos, err := s.fixSubTimeline(internalEngSub[inSelectSubIndex], exSubInfo)
		if err != nil {
			return err
		}
		if bFound == false {
			continue
		}
		// 调试的时候用
		if videoRootPath == "" {
			continue
		}
		for _, info := range subFixInfos {
			// 写入 fix 后的字幕文件覆盖之前的字幕文件
			desFixedSubFullName := path.Join(videoRootPath, info.FileName)
			log_helper.GetLogger().Debugln("Sub Timeline fixed:", desFixedSubFullName)
			continue
			//err = s.saveSubFile(desFixedSubFullName, info.FixContent)
			//if err != nil {
			//	return err
			//}
			//log_helper.GetLogger().Debugln("Sub Timeline fixed:", desFixedSubFullName)
		}
	}

	return nil
}

func (s SubTimelineFixerHelper) fixSubTimeline(enSubFile emby.SubInfo, ch_enSubFile emby.SubInfo) (bool, []sub_timeline_fixer.SubFixInfo, error) {

	bFind, infoBase, err := s.subParserHub.DetermineFileTypeFromBytes(enSubFile.Content, enSubFile.Ext)
	if err != nil {
		return false, nil, err
	}
	if bFind == false {
		return false, nil, nil
	}
	infoBase.Name = enSubFile.FileName
	/*
		这里发现一个梗，内置的英文字幕导出的时候，有可能需要合并多个 Dialogue，见
		internal/pkg/sub_helper/sub_helper.go 中 MergeMultiDialogue4EngSubtitle 的实现
	*/
	sub_helper.MergeMultiDialogue4EngSubtitle(infoBase)

	bFind, infoSrc, err := s.subParserHub.DetermineFileTypeFromBytes(ch_enSubFile.Content, ch_enSubFile.Ext)
	if err != nil {
		return false, nil, err
	}
	if bFind == false {
		return false, nil, nil
	}
	infoSrc.Name = ch_enSubFile.FileName

	infoBaseNameWithOutExt := strings.Replace(infoBase.Name, path.Ext(infoBase.Name), "", -1)
	infoSrcNameWithOutExt := strings.Replace(infoSrc.Name, path.Ext(infoSrc.Name), "", -1)

	// 把原始的文件缓存下来，新建缓存的文件夹
	cacheTmpPath := path.Join(tmpSubFixCacheFolder, infoBaseNameWithOutExt)
	if pkg.IsDir(cacheTmpPath) == false {
		err = os.MkdirAll(cacheTmpPath, os.ModePerm)
		if err != nil {
			return false, nil, err
		}
	}
	// 写入内置字幕、外置字幕原始文件
	err = s.saveSubFile(path.Join(cacheTmpPath, infoBaseNameWithOutExt+".chinese(inside)"+infoBase.Ext), infoBase.Content)
	if err != nil {
		return false, nil, err
	}
	err = s.saveSubFile(path.Join(cacheTmpPath, infoSrc.Name), infoSrc.Content)
	if err != nil {
		return false, nil, err
	}
	offsetTime, err := sub_timeline_fixer.GetOffsetTime(infoBase, infoSrc, path.Join(cacheTmpPath, infoSrcNameWithOutExt+"-bar.html"), path.Join(cacheTmpPath, infoSrcNameWithOutExt+".log"))
	if err != nil {
		return false, nil, err
	}
	if offsetTime != 0 {
		log_helper.GetLogger().Debugln(infoSrc.Name, "offset time is", fmt.Sprintf("%f", offsetTime), "s")
	}
	// 偏移很小就无视了
	if offsetTime < 0.2 && offsetTime > -0.2 {
		return false, nil, nil
	}
	// 写入校准时间轴后的字幕
	var subFixInfos = make([]sub_timeline_fixer.SubFixInfo, 0)
	for _, formatter := range s.formatter {
		// 符合已知的字幕命名格式，不符合就跳过，都跳过也行，就不做任何操作而已
		bMatch, fileNameWithOutExt, subExt, subLang, extraSubName := formatter.IsMatchThisFormat(infoSrc.Name)
		if bMatch == false {
			continue
		}
		// 生成对应字幕命名格式的，字幕命名。这里注意，normal 的时候， extraSubName+"-fix" 是无效的，不会被设置，也就是直接覆盖之前的字幕了。
		subNewName, _, _ := formatter.GenerateMixSubNameBase(fileNameWithOutExt, subExt, subLang, extraSubName+"-fix")
		desFixSubFileFullPath := path.Join(cacheTmpPath, subNewName)
		fixContent, err := sub_timeline_fixer.FixSubTimeline(infoSrc, offsetTime, desFixSubFileFullPath)
		if err != nil {
			return false, nil, err
		}
		subFixInfos = append(subFixInfos, *sub_timeline_fixer.NewSubFixInfo(infoSrc.Name, fixContent))
	}

	return true, subFixInfos, nil
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

const tmpSubFixCacheFolder = "SubFixCache"
