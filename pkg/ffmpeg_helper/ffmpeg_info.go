package ffmpeg_helper

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/sirupsen/logrus"
)

type FFMPEGInfo struct {
	log              *logrus.Logger
	VideoFullPath    string                // 视频文件的路径
	Duration         float64               // 视频的时长
	AudioInfoList    []AudioInfo           // 内置音频列表
	SubtitleInfoList []SubtitleInfo        // 内置字幕列表
	ExternalSubInfos []*subparser.FileInfo // 外置字幕列表
}

func NewFFMPEGInfo(log *logrus.Logger, videoFullPath string) *FFMPEGInfo {
	return &FFMPEGInfo{
		log:              log,
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
	return pkg.GetSubFixCacheFolderByName(noExtVideoName)
}

// IsExported 是否已经导出过，如果没有导出或者导出不完整为 false
func (f *FFMPEGInfo) IsExported(exportType ExportType) bool {

	bProcessDone := false
	nowCacheFolder, err := f.GetCacheFolderFPath()
	if err != nil {
		f.log.Errorln("FFMPEGInfo.IsExported.GetCacheFolderFPath", f.VideoFullPath, err.Error())
		return false
	}
	tmpNowExportedMaskFile := filepath.Join(nowCacheFolder, exportedMakeFileName)

	defer func() {
		// 函数执行完毕，再进行 check，是否需要删除 exportedMakeFileName 这个文件
		if bProcessDone == false {
			// 失败就需要删除这个 exportedMakeFileName 文件
			if pkg.IsFile(tmpNowExportedMaskFile) == true {
				_ = os.Remove(tmpNowExportedMaskFile)
			}
		}
	}()

	// 首先存储的缓存目录要存在
	if pkg.IsDir(nowCacheFolder) == false {
		return bProcessDone
	}

	if pkg.IsFile(tmpNowExportedMaskFile) == false {
		return bProcessDone
	}

	switch exportType {
	case Audio:
		// 音频是否导出了
		done := f.isAudioExported(nowCacheFolder)
		if done == false {
			return bProcessDone
		}
		break
	case Subtitle:
		// 字幕都要导出了
		done := f.isSubExported(nowCacheFolder)
		if done == false {
			return bProcessDone
		}
	case SubtitleAndAudio:
		// 音频是否导出了
		done := f.isAudioExported(nowCacheFolder)
		if done == false {
			return bProcessDone
		}
		// 字幕都要导出了
		done = f.isSubExported(nowCacheFolder)
		if done == false {
			return bProcessDone
		}
	default:
		return bProcessDone
	}

	bProcessDone = true

	return bProcessDone
}

func (f *FFMPEGInfo) CreateExportedMask() error {
	maskFileFPath, err := f.getExportedMaskFileFPath()
	if err != nil {
		return err
	}
	if pkg.IsFile(maskFileFPath) == false {
		create, err := os.Create(maskFileFPath)
		if err != nil {
			return err
		}
		defer create.Close()
	}

	return nil
}

func (f *FFMPEGInfo) getExportedMaskFileFPath() (string, error) {
	nowCacheFolder, err := f.GetCacheFolderFPath()
	if err != nil {
		return "", err
	}

	tmpNowExportedMaskFile := filepath.Join(nowCacheFolder, exportedMakeFileName)

	return tmpNowExportedMaskFile, nil
}

// isAudioExported 只需要确认导出了一个音频即可，同时在导出的时候也需要确定只导出一个，且识别出来多个音频，这里会调整到只有一个
func (f *FFMPEGInfo) isAudioExported(nowCacheFolder string) bool {

	newAudioInfos := make([]AudioInfo, 0)
	for index, audioInfo := range f.AudioInfoList {

		audioFPath := filepath.Join(nowCacheFolder, audioInfo.GetName()+extPCM)
		if pkg.IsFile(audioFPath) == true {

			f.AudioInfoList[index].FullPath = audioFPath

			tmpOneAudioInfo := NewAudioInfo(
				f.AudioInfoList[index].Index,
				f.AudioInfoList[index].CodecName,
				f.AudioInfoList[index].CodecType,
				f.AudioInfoList[index].timeBase,
				f.AudioInfoList[index].startTime,
				f.AudioInfoList[index].language,
			)
			tmpOneAudioInfo.FullPath = audioFPath
			tmpOneAudioInfo.Duration = f.AudioInfoList[index].Duration
			newAudioInfos = append(newAudioInfos, *tmpOneAudioInfo)
			// 替换
			f.AudioInfoList = newAudioInfos
			return true
		}
	}

	return false
}

func (f *FFMPEGInfo) isSubExported(nowCacheFolder string) bool {
	for index, subtitleInfo := range f.SubtitleInfoList {

		subSrtFPath := filepath.Join(nowCacheFolder, subtitleInfo.GetName()+common.SubExtSRT)
		if pkg.IsFile(subSrtFPath) == false {
			return false
		} else {
			f.SubtitleInfoList[index].FullPath = subSrtFPath
		}
		subASSFPath := filepath.Join(nowCacheFolder, subtitleInfo.GetName()+common.SubExtASS)
		if pkg.IsFile(subASSFPath) == false {
			return false
		} else {
			f.SubtitleInfoList[index].FullPath = subASSFPath
		}
	}
	return true
}

// GetExternalSubInfos 获取外置的字幕信息
func (f *FFMPEGInfo) GetExternalSubInfos(subParserHub *sub_parser_hub.SubParserHub) error {
	subFiles, err := sub_helper.SearchMatchedSubFileByOneVideo(f.log, f.VideoFullPath)
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

// 导出成功才生成这个文件
const exportedMakeFileName = "Exported"
