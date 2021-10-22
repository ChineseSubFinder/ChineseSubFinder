package ffmpeg_helper

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
)

type AudioInfo struct {
	Index     int
	CodecName string
	CodecType string
	timeBase  string
	startTime string
	language  string
}

func NewAudioInfo(index int, codecName, codecType, timeBase, startTime, language string) *AudioInfo {
	return &AudioInfo{
		Index:     index,
		CodecName: codecName,
		CodecType: codecType,
		timeBase:  timeBase,
		startTime: startTime,
		language:  language,
	}
}

// GetLanguage 获取音频的语言类型
func (a AudioInfo) GetLanguage() types.Language {
	return language.ISOString2SupportLang(a.language)
}

// GetName 获取音频名称，这里以音频的名称（中文）+ 索引的位置类描述
func (a AudioInfo) GetName() string {
	return fmt.Sprintf("%s_%d", language.Lang2ChineseString(a.GetLanguage()), a.Index)
}
