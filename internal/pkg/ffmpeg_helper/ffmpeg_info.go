package ffmpeg_helper

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
