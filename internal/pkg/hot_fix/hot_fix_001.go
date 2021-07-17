package hot_fix

import (
	movieHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/movie_helper"
	seriesHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
)

/*
	本模块的目标是解决开发过程中遗留的功能缺陷需要升级的问题
	之前字幕的命名不规范，现在需要进行一次批量的替换
	chs_en[shooter] -> Chinese(中英,shooter)
*/
type HotFix001 struct {
	movieRootDir string
	seriesRootDir string
}

func NewHotFix001(movieRootDir string, seriesRootDir string) *HotFix001 {
	return &HotFix001{movieRootDir: movieRootDir, seriesRootDir: seriesRootDir}
}

func (h HotFix001) GetKey() string {
	return "001"
}

func (h HotFix001) Process() error {

	var err error
	// 先找出有那些电影文件夹和连续剧文件夹
	movieFullPathList, err := pkg.SearchMatchedVideoFile(h.movieRootDir)
	if err != nil {
		return err
	}
	seriesDirList, err := seriesHelper.GetSeriesList(h.seriesRootDir)
	if err != nil {
		return err
	}
	// 搜索所有的字幕，找到相关的字幕进行修改
	for _, one := range movieFullPathList {
		found, _, fitMovieNameSubList, err := movieHelper.MovieHasChineseSub(one)
		if err != nil || found == false {
			continue
		}


	}

	seriesSubFiles, err := sub_helper.SearchMatchedSubFile(h.seriesRootDir)
	if err != nil {
		return err
	}
	// 从找到的这些字幕中，找出，结尾是

	println(len(movieFullPathList))
	println(len(seriesDirList))
	println(len(seriesSubFiles))

	return nil
}
