package ffmpeg_helper

import (
	"fmt"
	language2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
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
	FullPath   string
}

func NewSubtitleInfo(index int, codecName, codecType, timeBase, startTime string, durationTS int, duration, language string) *SubtitleInfo {
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
func (s SubtitleInfo) GetLanguage() language2.MyLanguage {
	return language.ISOString2SupportLang(s.language)
}

// GetName 获取字幕名称，这里以语言的名称（中文）+ 索引的位置类描述
func (s SubtitleInfo) GetName() string {
	return fmt.Sprintf("%s_%d", language.Lang2ChineseString(s.GetLanguage()), s.Index)
}
