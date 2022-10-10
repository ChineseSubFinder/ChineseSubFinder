package video_list_helper

import (
	"testing"
)

func TestVideoListHelper_GetMoviePoster(t *testing.T) {

	movieFPath := "X:\\movie\\The King's Man (2021)\\The King's Man (2021) WEBDL-1080p.mkv"
	v := VideoListHelper{}
	println("Poster:", v.GetMoviePoster(movieFPath))
}

func TestVideoListHelper_GetSeriesPoster(t *testing.T) {

	seriesDir := "X:\\连续剧\\良医 (2017)"
	v := VideoListHelper{}
	println("Poster:", v.GetSeriesPoster(seriesDir))
}
