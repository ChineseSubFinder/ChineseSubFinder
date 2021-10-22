package ffmpeg_helper

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
)

type SubtitleInfo struct {
	Index      int
	CodecName  string
	CodecType  string
	timeBase   string
	startTime  string
	durationTS int
	duration   string
	language   string
	content    string
}

func NewSubtitileInfo(index int, codecName, codecType, timeBase, startTime string, durationTS int, duration, language string) *SubtitleInfo {
	return &SubtitleInfo{
		Index:      index,
		CodecName:  codecName,
		CodecType:  codecType,
		timeBase:   timeBase,
		startTime:  startTime,
		durationTS: durationTS,
		duration:   duration,
		language:   language,
	}
}

// SetContent 设置字幕的内容，同时进行字幕语言的判断
func (s *SubtitleInfo) SetContent(content string) error {
	s.content = content
	return nil
}

// GetLanguage 获取字幕语言的类型
func (s SubtitleInfo) GetLanguage() types.Language {
	return language.ISOString2SupportLang(s.language)
}

// GetName 获取字幕名称，这里以语言的名称（中文）+ 索引的位置类描述
func (s SubtitleInfo) GetName() string {
	return fmt.Sprintf("%s_%d", language.Lang2ChineseString(s.GetLanguage()), s.Index)
}
