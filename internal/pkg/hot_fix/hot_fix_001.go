package hot_fix

import (
	"errors"
	movieHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/movie_helper"
	seriesHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
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
	movieRootDir  string
	seriesRootDir string
}

func NewHotFix001(movieRootDir string, seriesRootDir string) *HotFix001 {
	return &HotFix001{movieRootDir: movieRootDir, seriesRootDir: seriesRootDir}
}

func (h HotFix001) GetKey() string {
	return "001"
}

func (h HotFix001) Process() (interface{}, error) {

	var err error
	outStruct := OutStruct001{}
	outStruct.RenamedFiles = make([]string, 0)
	outStruct.ErrFiles = make([]string, 0)
	if pkg.IsDir(h.movieRootDir) == false {
		return outStruct, errors.New("movieRootDir path not exist: " + h.movieRootDir)
	}
	if pkg.IsDir(h.seriesRootDir) == false {
		return outStruct, errors.New("seriesRootDir path not exist: " + h.seriesRootDir)
	}
	// 先找出有那些电影文件夹和连续剧文件夹
	var movieFullPathList = make([]string, 0)
	movieFullPathList, err = pkg.SearchMatchedVideoFile(h.movieRootDir)
	if err != nil {
		return outStruct, err
	}
	seriesDirList, err := seriesHelper.GetSeriesList(h.seriesRootDir)
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
