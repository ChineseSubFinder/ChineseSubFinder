package video_list_helper

import "github.com/allanpk716/ChineseSubFinder/pkg/sub_helper"

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
