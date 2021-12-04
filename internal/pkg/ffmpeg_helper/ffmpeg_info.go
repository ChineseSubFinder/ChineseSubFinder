package ffmpeg_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"path/filepath"
	"strings"
)

type FFMPEGInfo struct {
	VideoFullPath    string                // 视频文件的路径
	AudioInfoList    []AudioInfo           // 内置音频列表
	SubtitleInfoList []SubtitleInfo        // 内置字幕列表
	ExternalSubInfos []*subparser.FileInfo // 外置字幕列表
}

func NewFFMPEGInfo(videoFullPath string) *FFMPEGInfo {
	return &FFMPEGInfo{
		VideoFullPath:    videoFullPath,
		AudioInfoList:    make([]AudioInfo, 0),
		SubtitleInfoList: make([]SubtitleInfo, 0),
		ExternalSubInfos: make([]*subparser.FileInfo, 0),
	}
}

// GetCacheFolderFPath 获取缓存文件夹的绝对路径，存储在通用的 SubFixCacheFolder 中
// csf-cache/当前的视频文件名(不带后缀)
func (f *FFMPEGInfo) GetCacheFolderFPath() (string, error) {
	noExtVideoName := strings.ReplaceAll(filepath.Base(f.VideoFullPath), filepath.Ext(f.VideoFullPath), "")
	return my_folder.GetSubFixCacheFolderByName(noExtVideoName)
}

// IsExported 是否已经导出过，如果没有导出或者导出不完整为 false
func (f *FFMPEGInfo) IsExported() bool {

	nowCacheFolder, err := f.GetCacheFolderFPath()
	if err != nil {
		log_helper.GetLogger().Errorln("FFMPEGInfo.IsExported.GetCacheFolderFPath", f.VideoFullPath, err.Error())
		return false
	}
	// 首先存储的缓存目录要存在
	if my_util.IsDir(nowCacheFolder) == false {
		return false
	}
	// 字幕都要导出了
	for index, subtitleInfo := range f.SubtitleInfoList {

		subSrtFPath := filepath.Join(nowCacheFolder, subtitleInfo.GetName()+common.SubExtSRT)
		if my_util.IsFile(subSrtFPath) == false {
			return false
		} else {
			f.SubtitleInfoList[index].FullPath = subSrtFPath
		}
		subASSFPath := filepath.Join(nowCacheFolder, subtitleInfo.GetName()+common.SubExtASS)
		if my_util.IsFile(subASSFPath) == false {
			return false
		} else {
			f.SubtitleInfoList[index].FullPath = subASSFPath
		}

	}
	//TODO 音频可以后面再导出，按需。因为优先级最高的还是用字幕修复字幕，然后才是音频修复字幕
	// 音频是否导出了
	for index, audioInfo := range f.AudioInfoList {
		audioFPath := filepath.Join(nowCacheFolder, audioInfo.GetName()+extPCM)
		if my_util.IsFile(audioFPath) == false {
			return false
		} else {
			f.AudioInfoList[index].FullPath = audioFPath
		}
	}

	return true
}

// GetExternalSubInfos 获取外置的字幕信息
func (f *FFMPEGInfo) GetExternalSubInfos(subParserHub *sub_parser_hub.SubParserHub) error {
	subFiles, err := sub_helper.SearchMatchedSubFileByOneVideo(f.VideoFullPath)
	if err != nil {
		return err
	}
	for _, subFile := range subFiles {
		bok, subInfo, err := subParserHub.DetermineFileTypeFromFile(subFile)
		if err != nil {
			return err
		}
		if bok == false {
			continue
		}
		f.ExternalSubInfos = append(f.ExternalSubInfos, subInfo)
	}

	return nil
}
