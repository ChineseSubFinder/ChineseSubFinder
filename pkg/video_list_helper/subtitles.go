package video_list_helper

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
)

// GetMovieSubtitles 获取一部电影的字幕文件列表
func (v *VideoListHelper) GetMovieSubtitles(movieFPath string) []string {

	outSubtitles := make([]string, 0)
	subtitles, err := sub_helper.SearchMatchedSubFileByOneVideo(v.log, movieFPath)
	if err != nil {
		v.log.Error("GetMovieSubtitles.SearchMatchedSubFileByOneVideo", err)
		return outSubtitles
	}
	outSubtitles = append(outSubtitles, subtitles...)
	return outSubtitles
}

func (v *VideoListHelper) GetSeriesSubtitles(seriesDir string) []string {

	// 一次性搜索这个连续剧所有的视频，然后针对每一集再搜索对应的字幕出来
	// 1. 获取这个连续剧的所有视频
	return nil
}
