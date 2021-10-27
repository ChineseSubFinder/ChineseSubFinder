package ffmpeg_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"path/filepath"
	"strings"
)

type FFMPEGInfo struct {
	VideoFullPath    string
	AudioInfoList    []AudioInfo
	SubtitleInfoList []SubtitleInfo
}

func NewFFMPEGInfo(videoFullPath string) *FFMPEGInfo {
	return &FFMPEGInfo{
		VideoFullPath:    videoFullPath,
		AudioInfoList:    make([]AudioInfo, 0),
		SubtitleInfoList: make([]SubtitleInfo, 0),
	}
}

// GetCacheFolderFPath 获取缓存文件夹的绝对路径，存储在每个视频当前的路劲下
// csf-cache/当前的视频文件名(不带后缀)
func (f *FFMPEGInfo) GetCacheFolderFPath() string {
	noExtVideoName := strings.ReplaceAll(filepath.Base(f.VideoFullPath), filepath.Ext(f.VideoFullPath), "")
	return filepath.Join(filepath.Dir(f.VideoFullPath), cacheFolder, noExtVideoName)
}

// IsExported 是否已经导出过，如果没有导出或者导出不完整为 false
func (f *FFMPEGInfo) IsExported() bool {

	nowCacheFolder := f.GetCacheFolderFPath()
	// 首先存储的缓存目录要存在
	if pkg.IsDir(nowCacheFolder) == false {
		return false
	}
	// 字幕都要导出了
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
	// 音频是否导出了
	for index, audioInfo := range f.AudioInfoList {
		audioFPath := filepath.Join(nowCacheFolder, audioInfo.GetName()+extPCM)
		if pkg.IsFile(audioFPath) == false {
			return false
		} else {
			f.AudioInfoList[index].FullPath = audioFPath
		}
	}

	return true
}

const cacheFolder = "csf-cache"
