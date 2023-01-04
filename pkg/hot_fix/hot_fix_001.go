package hot_fix

import (
	"errors"
	"os"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/search"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	movieHelper "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/movie_helper"
	seriesHelper "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/series_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/old"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/sirupsen/logrus"
)

/*
	本模块的目标是解决开发过程中遗留的功能缺陷需要升级的问题
	之前字幕的命名不规范，现在需要进行一次批量的替换
	chs_en[shooter] -> Chinese(中英,shooter)
*/
type HotFix001 struct {
	log            *logrus.Logger
	movieRootDirs  []string
	seriesRootDirs []string
}

func NewHotFix001(log *logrus.Logger, movieRootDirs []string, seriesRootDirs []string) *HotFix001 {
	return &HotFix001{log: log, movieRootDirs: movieRootDirs, seriesRootDirs: seriesRootDirs}
}

func (h HotFix001) GetKey() string {
	return "001"
}

func (h HotFix001) Process() (interface{}, error) {

	defer func() {
		h.log.Infoln("Hotfix", h.GetKey(), "End")
	}()

	h.log.Infoln("Hotfix", h.GetKey(), "Start...")

	return h.process()
}

func (h HotFix001) process() (OutStruct001, error) {

	outStruct := OutStruct001{}
	outStruct.RenamedFiles = make([]string, 0)
	outStruct.ErrFiles = make([]string, 0)

	for i, dir := range h.movieRootDirs {
		h.log.Infoln("Fix Movie Dir Index", i, dir, "Start...")
		fixMovie, err := h.fixMovie(dir)
		if err != nil {
			h.log.Errorln("Fix Movie Dir Index", i, dir, "End With Error", err)
			return outStruct, err
		}

		outStruct.RenamedFiles = append(outStruct.RenamedFiles, fixMovie.RenamedFiles...)
		outStruct.ErrFiles = append(outStruct.ErrFiles, fixMovie.ErrFiles...)

		h.log.Infoln("Fix Movie Dir Index", i, dir, "End...")
	}

	for i, dir := range h.seriesRootDirs {
		h.log.Infoln("Fix Series Dir Index", i, dir, "Start...")
		fixSeries, err := h.fixSeries(dir)
		if err != nil {
			h.log.Errorln("Fix Series Dir Index", i, dir, "End With Error", err)
			return outStruct, err
		}

		outStruct.RenamedFiles = append(outStruct.RenamedFiles, fixSeries.RenamedFiles...)
		outStruct.ErrFiles = append(outStruct.ErrFiles, fixSeries.ErrFiles...)

		h.log.Infoln("Fix Series Dir Index", i, dir, "End...")
	}

	return outStruct, nil
}

func (h HotFix001) fixMovie(movieRootDir string) (OutStruct001, error) {

	var err error
	outStruct := OutStruct001{}
	outStruct.RenamedFiles = make([]string, 0)
	outStruct.ErrFiles = make([]string, 0)
	if pkg.IsDir(movieRootDir) == false {
		return outStruct, errors.New("movieRootDir path not exist: " + movieRootDir)
	}
	// 先找出有那些电影文件夹和连续剧文件夹
	var movieFullPathList = make([]string, 0)
	movieFullPathList, err = search.MatchedVideoFile(h.log, movieRootDir)
	if err != nil {
		return outStruct, err
	}
	// 搜索所有的字幕，找到相关的字幕进行修改
	for _, one := range movieFullPathList {
		found := false
		var fitMovieNameSubList = make([]string, 0)
		found, _, fitMovieNameSubList, err = movieHelper.MovieHasChineseSub(h.log, one)
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

func (h HotFix001) fixSeries(seriesRootDir string) (OutStruct001, error) {
	var err error
	outStruct := OutStruct001{}
	outStruct.RenamedFiles = make([]string, 0)
	outStruct.ErrFiles = make([]string, 0)
	if pkg.IsDir(seriesRootDir) == false {
		return outStruct, errors.New("seriesRootDir path not exist: " + seriesRootDir)
	}
	// 先找出有那些电影文件夹和连续剧文件夹
	seriesDirList, err := seriesHelper.GetSeriesList(h.log, seriesRootDir)
	if err != nil {
		return outStruct, err
	}
	// 连续剧
	var seriesSubFiles = make([]string, 0)
	for _, oneSeriesDir := range seriesDirList {
		seriesSubFiles, err = sub_helper.SearchMatchedSubFileByDir(h.log, oneSeriesDir)
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
