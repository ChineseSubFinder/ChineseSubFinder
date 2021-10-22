package ffmpeg_helper

import (
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

const cacheFolder = "csf-cache"
