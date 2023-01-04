package video_list_helper

import (
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
)

// GetMoviePoster 获取电影的海报，如果为空就是没有找到
func (v *VideoListHelper) GetMoviePoster(movieFPath string) string {
	/*
		ext 只考虑 jpg, png, bmp 三种格式
		参考 TMM 的设置
		1. poster.ext
		2. movie.ext
		3. folder.ext
		4. <movie filename>-poster.ext
		5. <movie filename>.ext
		6. cover.ext
	*/
	for _, ext := range extList {

		movieRootDir := filepath.Dir(movieFPath)
		movieName := filepath.Base(movieFPath)
		movieNameWithoutExt := movieName[:len(movieName)-len(filepath.Ext(movieName))]
		// 1. poster.ext
		posterFPath := filepath.Join(movieRootDir, "poster"+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
		// 2. movie.ext
		posterFPath = filepath.Join(movieRootDir, "movie"+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
		// 3. folder.ext
		posterFPath = filepath.Join(movieRootDir, "folder"+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
		// 4. <movie filename>-poster.ext
		posterFPath = filepath.Join(movieRootDir, movieNameWithoutExt+"-poster"+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
		// 5. <movie filename>.ext
		posterFPath = filepath.Join(movieRootDir, movieNameWithoutExt+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
		// 6. cover.ext
		posterFPath = filepath.Join(movieRootDir, "cover"+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
	}
	return ""
}

// GetSeriesPoster 获取电视剧的海报，如果为空就是没有找到
func (v *VideoListHelper) GetSeriesPoster(seriesDir string) string {
	/*
		参考 TMM 的设置
		连续剧的
		1. poster.ext
		2. folder.ext
		Emby 的
		3. fanart.ext
		Season的
		1. seasonXX-poster.ext
		2. <season folder>/seasonXX.ext
		3. <season folder>/folder.ext
	*/
	// 获取主封面
	for _, ext := range extList {
		// 1. poster.ext
		posterFPath := filepath.Join(seriesDir, "poster"+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
		// 2. folder.ext
		posterFPath = filepath.Join(seriesDir, "folder"+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
		// 3. fanart.ext
		posterFPath = filepath.Join(seriesDir, "fanart"+ext)
		if pkg.IsFile(posterFPath) {
			return posterFPath
		}
	}

	return ""
}

var (
	extList = []string{".jpg", ".png", ".bmp"}
)

type SeriesPosterInfo struct {
	SeriesPoster    string
	SeasonPosterMap map[int]string
}
