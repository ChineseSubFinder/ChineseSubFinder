package hot_fix

import (
	"errors"
	movieHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/movie_helper"
	seriesHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/old"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"os"
)

/*
	本模块的目标是解决开发过程中遗留的功能缺陷需要升级的问题
	之前字幕的命名不规范，现在需要进行一次批量的替换
	chs_en[shooter] -> Chinese(中英,shooter)
*/
type HotFix001 struct {
	movieRootDirs  []string
	seriesRootDirs []string
}

func NewHotFix001(movieRootDirs []string, seriesRootDirs []string) *HotFix001 {
	return &HotFix001{movieRootDirs: movieRootDirs, seriesRootDirs: seriesRootDirs}
}

func (h HotFix001) GetKey() string {
	return "001"
}

func (h HotFix001) Process() (interface{}, error) {
	return h.process()
}

func (h HotFix001) process() (interface{}, error) {

	outStruct := OutStruct001{}
	outStruct.RenamedFiles = make([]string, 0)
	outStruct.ErrFiles = make([]string, 0)

	for i, dir := range h.movieRootDirs {
		log_helper.GetLogger().Infoln("Fix Movie Dir Index", i, dir, "Start...")
		fixMovie, err := h.fixMovie(dir)
		if err != nil {
			log_helper.GetLogger().Infoln("Fix Movie Dir Index", i, dir, "End...")
			return nil, err
		}

		outStruct.RenamedFiles = append(outStruct.RenamedFiles, fixMovie.(OutStruct001).RenamedFiles...)
		outStruct.ErrFiles = append(outStruct.ErrFiles, fixMovie.(OutStruct001).ErrFiles...)

		log_helper.GetLogger().Infoln("Fix Movie Dir Index", i, dir, "End...")
	}

	for i, dir := range h.seriesRootDirs {
		log_helper.GetLogger().Infoln("Fix Series Dir Index", i, dir, "Start...")
		fixSeries, err := h.fixMovie(dir)
		if err != nil {
			log_helper.GetLogger().Infoln("Fix Series Dir Index", i, dir, "End...")
			return nil, err
		}

		outStruct.RenamedFiles = append(outStruct.RenamedFiles, fixSeries.(OutStruct001).RenamedFiles...)
		outStruct.ErrFiles = append(outStruct.ErrFiles, fixSeries.(OutStruct001).ErrFiles...)

		log_helper.GetLogger().Infoln("Fix Series Dir Index", i, dir, "End...")
	}

	return outStruct, nil
}

func (h HotFix001) fixMovie(movieRootDir string) (interface{}, error) {

	var err error
	outStruct := OutStruct001{}
	outStruct.RenamedFiles = make([]string, 0)
	outStruct.ErrFiles = make([]string, 0)
	if my_util.IsDir(movieRootDir) == false {
		return outStruct, errors.New("movieRootDir path not exist: " + movieRootDir)
	}
	// 先找出有那些电影文件夹和连续剧文件夹
	var movieFullPathList = make([]string, 0)
	movieFullPathList, err = my_util.SearchMatchedVideoFile(movieRootDir)
	if err != nil {
		return outStruct, err
	}
	// 搜索所有的字幕，找到相关的字幕进行修改
	for _, one := range movieFullPathList {
		found := false
		var fitMovieNameSubList = make([]string, 0)
		found, _, fitMovieNameSubList, err = movieHelper.MovieHasChineseSub(one)
		if err != nil || found == false {
			continue
		}
		// 判断是否是符合要求
		for _, fitSubName := range fitMovieNameSubList {
			bFix, _, newSubFileName := old.IsOldVersionSubPrefixName(fitSubName)
			if bFix == false {
				continue
			}
			err = os.Rename(fitSubName, newSubFileName)
			if err != nil {
				outStruct.ErrFiles = append(outStruct.ErrFiles, fitSubName)
				continue
			}
			outStruct.RenamedFiles = append(outStruct.RenamedFiles, newSubFileName)
		}
	}
	return outStruct, nil
}

func (h HotFix001) fixSeries(seriesRootDir string) (interface{}, error) {
	var err error
	outStruct := OutStruct001{}
	outStruct.RenamedFiles = make([]string, 0)
	outStruct.ErrFiles = make([]string, 0)
	if my_util.IsDir(seriesRootDir) == false {
		return outStruct, errors.New("seriesRootDir path not exist: " + seriesRootDir)
	}
	// 先找出有那些电影文件夹和连续剧文件夹
	seriesDirList, err := seriesHelper.GetSeriesList(seriesRootDir)
	if err != nil {
		return outStruct, err
	}
	// 连续剧
	var seriesSubFiles = make([]string, 0)
	for _, oneSeriesDir := range seriesDirList {
		seriesSubFiles, err = sub_helper.SearchMatchedSubFileByDir(oneSeriesDir)
		if err != nil {
			return outStruct, err
		}
		// 判断是否是符合要求
		for _, fitSubName := range seriesSubFiles {
			bFix, _, newSubFileName := old.IsOldVersionSubPrefixName(fitSubName)
			if bFix == false {
				continue
			}
			err = os.Rename(fitSubName, newSubFileName)
			if err != nil {
				outStruct.ErrFiles = append(outStruct.ErrFiles, fitSubName)
				continue
			}
			outStruct.RenamedFiles = append(outStruct.RenamedFiles, newSubFileName)
		}
	}

	return outStruct, nil
}

type OutStruct001 struct {
	RenamedFiles []string
	ErrFiles     []string
}
