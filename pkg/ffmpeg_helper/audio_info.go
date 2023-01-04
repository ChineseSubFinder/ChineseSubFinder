package ffmpeg_helper

import (
	"fmt"

	language2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
)

type AudioInfo struct {
	Index     int
	CodecName string
	CodecType string
	timeBase  string
	startTime string
	language  string
	FullPath  string
	Duration  float64
}

func NewAudioInfo(index int, codecName, codecType, timeBase, startTime, language string) *AudioInfo {
	return &AudioInfo{
		Index:     index,
		CodecName: codecName,
		CodecType: codecType,
		timeBase:  timeBase,
		startTime: startTime,
		language:  language,
		Duration:  0,
	}
}

// GetLanguage 获取音频的语言类型
func (a AudioInfo) GetLanguage() language2.MyLanguage {
	return language.ISOString2SupportLang(a.language)
}

// GetName 获取音频名称，这里以音频的名称（中文）+ 索引的位置类描述
func (a AudioInfo) GetName() string {
	return fmt.Sprintf("%s_%d", language.Lang2ChineseString(a.GetLanguage()), a.Index)
}
