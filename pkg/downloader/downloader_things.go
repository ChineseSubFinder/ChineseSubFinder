package downloader

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/series"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	subcommon "github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
)

// oneVideoSelectBestSub 一个视频，选择最佳的一个字幕（也可以保存所有网站第一个最佳字幕）
func (d *Downloader) oneVideoSelectBestSub(oneVideoFullPath string, organizeSubFiles []string) error {

	// 如果没有则直接跳过
	if organizeSubFiles == nil || len(organizeSubFiles) < 1 {
		return common.AllSiteDownloadSubNotFound
	}

	var err error
	// 得到目标视频文件的文件名
	videoFileName := filepath.Base(oneVideoFullPath)
	// -------------------------------------------------
	// 调试缓存，把下载好的字幕写到对应的视频目录下，方便调试
	if settings.Get().AdvancedSettings.DebugMode == true {

		err = pkg.CopyFiles2DebugFolder([]string{videoFileName}, organizeSubFiles)
		if err != nil {
			// 这个错误可以忍
			d.log.Errorln("copySubFile2DesFolder", err)
		}
	}
	// -------------------------------------------------
	/*
		这里需要额外考虑一点，有可能当前目录已经有一个 .Default .Forced 标记的字幕了
		那么下载字幕丢进来的时候就需要提前把这个字幕找出来，去除整个 .Default .Forced  标记
		然后进行正常的下载，存储和替换字幕，最后将本次操作的第一次标记为 .Default
	*/
	// 不管是不是保存多个字幕，都要先扫描本地的字幕，进行 .Default .Forced 去除
	// 这个视频的所有字幕，去除 .default .Forced 标记
	err = sub_helper.SearchVideoMatchSubFileAndRemoveExtMark(d.log, oneVideoFullPath)
	if err != nil {
		// 找个错误可以忍
		d.log.Errorln("SearchVideoMatchSubFileAndRemoveExtMark,", oneVideoFullPath, err)
	}
	if settings.Get().AdvancedSettings.SaveMultiSub == false {
		// 选择最优的一个字幕
		var finalSubFile *subparser.FileInfo
		finalSubFile = d.mk.SelectOneSubFile(organizeSubFiles)
		if finalSubFile == nil {
			outString := fmt.Sprintln("Found", len(organizeSubFiles), " subtitles but not one fit:", oneVideoFullPath)
			d.log.Warnln(outString)
			return errors.New(outString)
		}
		/*
			这里还有一个梗，Emby、jellyfin 支持 default 和 forced 扩展字段
			但是，plex 只支持 forced
			那么就比较麻烦，干脆，normal 的命名格式化实例，就不设置 default 了，forced 不想用，因为可能会跟你手动选择的字幕冲突（下次观看的时候，理论上也可能不会）
		*/
		// 判断配置文件中的字幕命名格式化的选择
		bSetDefault := true
		if d.subNameFormatter == subcommon.Normal {
			bSetDefault = false
		}
		// 找到了，写入文件
		err = d.SaveSubHelper.WriteSubFile2VideoPath(oneVideoFullPath, *finalSubFile, "", bSetDefault, false)
		if err != nil {
			return errors.New(fmt.Sprintf("SaveMultiSub: %v, writeSubFile2VideoPath, Error: %v ", settings.Get().AdvancedSettings.SaveMultiSub, err))
		}
	} else {
		// 每个网站 Top1 的字幕
		siteNames, finalSubFiles := d.mk.SelectEachSiteTop1SubFile(organizeSubFiles)
		if len(siteNames) < 0 {
			outString := fmt.Sprintln("SelectEachSiteTop1SubFile found none sub file")
			d.log.Warnln(outString)
			return errors.New(outString)
		}
		// 多网站 Top 1 字幕保存的时候，第一个设置为 Default 即可
		/*
			由于新功能支持了字幕命名格式的选择，那么如果触发了多个字幕保存的逻辑，如果不调整
			则会遇到，top1 先写入，然后 top2 覆盖 top1 ，以此类推的情况出现
			所以如果开启了 Normal SubNameFormatter 的功能，则要反序写入文件
			如果是 Emby 的字幕命名格式则无需考虑此问题，因为每个网站只会有一个字幕，且字幕命名格式决定了不会重复写入覆盖
		*/
		if d.subNameFormatter == subcommon.Emby {
			for i, file := range finalSubFiles {
				setDefault := false
				if i == 0 {
					setDefault = true
				}
				err = d.SaveSubHelper.WriteSubFile2VideoPath(oneVideoFullPath, file, siteNames[i], setDefault, false)
				if err != nil {
					return errors.New(fmt.Sprintf("SaveMultiSub: %v, writeSubFile2VideoPath, Error: %v ", settings.Get().AdvancedSettings.SaveMultiSub, err))
				}
			}
		} else {
			// 默认这里就是 normal 模式
			// 逆序写入
			/*
				这里还有一个梗，Emby、jellyfin 支持 default 和 forced 扩展字段
				但是，plex 只支持 forced
				那么就比较麻烦，干脆，normal 的命名格式化实例，就不设置 default 了，forced 不想用，因为可能会跟你手动选择的字幕冲突（下次观看的时候，理论上也可能不会）
			*/
			for i := len(finalSubFiles) - 1; i > -1; i-- {
				err = d.SaveSubHelper.WriteSubFile2VideoPath(oneVideoFullPath, finalSubFiles[i], siteNames[i], false, false)
				if err != nil {
					return errors.New(fmt.Sprintf("SaveMultiSub: %v, writeSubFile2VideoPath, Error: %v ", settings.Get().AdvancedSettings.SaveMultiSub, err))
				}
			}
		}
	}
	// -------------------------------------------------

	return nil
}

// saveFullSeasonSub 这里就需要单独存储到连续剧每一季的文件夹的特殊文件夹中。需要跟 DeleteOneSeasonSubCacheFolder 关联起来
func (d *Downloader) saveFullSeasonSub(seriesInfo *series.SeriesInfo, organizeSubFiles map[string][]string) map[string][]string {

	var fullSeasonSubDict = make(map[string][]string)

	for _, season := range seriesInfo.SeasonDict {
		seasonKey := pkg.GetEpisodeKeyName(season, 0)
		subs, ok := organizeSubFiles[seasonKey]
		if ok == false {
			continue
		}
		for _, sub := range subs {
			subFileName := filepath.Base(sub)

			newSeasonSubRootPath, err := pkg.GetDebugFolderByName([]string{
				filepath.Base(seriesInfo.DirPath),
				"Sub_" + seasonKey})
			if err != nil {
				d.log.Errorln("saveFullSeasonSub.GetDebugFolderByName", subFileName, err)
				continue
			}

			newSubFullPath := filepath.Join(newSeasonSubRootPath, subFileName)
			err = pkg.CopyFile(sub, newSubFullPath)
			if err != nil {
				d.log.Errorln("saveFullSeasonSub.CopyFile", subFileName, err)
				continue
			}
			// 从字幕的文件名推断是 哪一季 的 那一集
			_, gusSeason, gusEpisode, err := decode.GetSeasonAndEpisodeFromSubFileName(subFileName)
			if err != nil {
				return nil
			}
			// 把整季的字幕缓存位置也提供出去，如果之前没有下载到的，这里返回出来的可以补上
			seasonEpsKey := pkg.GetEpisodeKeyName(gusSeason, gusEpisode)
			_, ok := fullSeasonSubDict[seasonEpsKey]
			if ok == false {
				// 初始化
				fullSeasonSubDict[seasonEpsKey] = make([]string, 0)
			}
			fullSeasonSubDict[seasonEpsKey] = append(fullSeasonSubDict[seasonEpsKey], sub)
		}
	}

	return fullSeasonSubDict
}
