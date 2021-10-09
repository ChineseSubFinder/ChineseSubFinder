package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/pkg/emby_api"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"os"
	"path"
	"strings"
	"time"
)

type SubTimelineFixerHelper struct {
	embyApi      *embyHelper.EmbyApi
	embyHelper   *emby_helper.EmbyHelper
	EmbyConfig   emby.EmbyConfig
	subParserHub *sub_parser_hub.SubParserHub
	subFormatter ifaces.ISubFormatter //	字幕格式化命名的实现
	threads      int
	timeOut      time.Duration
}

func NewSubTimelineFixerHelper(embyConfig emby.EmbyConfig, inSubFormatter ifaces.ISubFormatter) *SubTimelineFixerHelper {
	sub := SubTimelineFixerHelper{
		EmbyConfig:   embyConfig,
		embyHelper:   emby_helper.NewEmbyHelper(embyConfig),
		embyApi:      embyHelper.NewEmbyApi(embyConfig),
		subParserHub: sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser()),
		subFormatter: inSubFormatter,
		threads:      6,
		timeOut:      60 * time.Second,
	}
	return &sub
}

func (s SubTimelineFixerHelper) FixRecentlyItemsSubTimeline() error {

	items, err := s.embyApi.GetRecentlyItems()
	if err != nil {
		return err
	}
	for _, item := range items.Items {
		err = s.fixOneVideoSub(item.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s SubTimelineFixerHelper) fixOneVideoSub(videoId string) error {
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

		bFound, err := s.fixSubTimeline(internalEngSub[inSelectSubIndex], exSubInfo)
		if err != nil {
			return err
		}
		if bFound == false {
			continue
		}
	}

	return nil
}

func (s SubTimelineFixerHelper) fixSubTimeline(enSubFile emby.SubInfo, ch_enSubFile emby.SubInfo) (bool, error) {

	bFind, infoBase, err := s.subParserHub.DetermineFileTypeFromBytes(enSubFile.Content, enSubFile.Ext)
	if err != nil {
		return false, err
	}
	if bFind == false {
		return false, nil
	}
	infoBase.Name = enSubFile.FileName
	bFind, infoSrc, err := s.subParserHub.DetermineFileTypeFromBytes(ch_enSubFile.Content, ch_enSubFile.Ext)
	if err != nil {
		return false, err
	}
	if bFind == false {
		return false, nil
	}
	infoSrc.Name = ch_enSubFile.FileName

	infoBaseNameWithOutExt := strings.Replace(infoBase.Name, path.Ext(infoBase.Name), "", -1)
	infoSrcNameWithOutExt := strings.Replace(infoSrc.Name, path.Ext(infoSrc.Name), "", -1)

	// 把原始的文件缓存下来，新建缓存的文件夹
	cacheTmpPath := path.Join(tmpFolder, infoBaseNameWithOutExt)
	if pkg.IsDir(cacheTmpPath) == false {
		err = os.MkdirAll(cacheTmpPath, os.ModePerm)
		if err != nil {
			return false, err
		}
	}
	offsetTime, err := sub_timeline_fixer.GetOffsetTime(infoBase, infoSrc, path.Join(cacheTmpPath, infoSrcNameWithOutExt+"-bar.html"))
	if err != nil {
		return false, err
	}
	// 偏移很小就无视了
	if offsetTime < 0.2 && offsetTime > -0.2 {
		_ = os.RemoveAll(cacheTmpPath)
		return false, nil
	}
	// 写入内置字幕、外置字幕原始文件
	err = s.saveOrgSubFile(path.Join(cacheTmpPath, infoBaseNameWithOutExt+".chinese(inside)"+infoBase.Ext), infoBase.Content)
	if err != nil {
		return false, err
	}
	err = s.saveOrgSubFile(path.Join(cacheTmpPath, infoSrc.Name), infoSrc.Content)
	if err != nil {
		return false, err
	}
	// 写入校准时间轴后的字幕
	bMatch, fileNameWithOutExt, subExt, subLang, extraSubName := s.subFormatter.IsMatchThisFormat(infoSrc.Name)
	if bMatch == false {
		return false, nil
	}
	subNewName, _, _ := s.subFormatter.GenerateMixSubNameBase(fileNameWithOutExt, subExt, subLang, extraSubName+"-fix")
	desFixSubFileFullPath := path.Join(cacheTmpPath, subNewName)
	err = sub_timeline_fixer.FixSubTimeline(infoSrc, offsetTime, desFixSubFileFullPath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s SubTimelineFixerHelper) saveOrgSubFile(desSaveSubFileFullPath string, content string) error {
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

const tmpFolder = "tmpSubFix"
