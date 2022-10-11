package my_util

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/allanpk716/ChineseSubFinder/pkg/sub_parser_hub"

	"github.com/allanpk716/ChineseSubFinder/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/pkg/filter"
	"github.com/allanpk716/ChineseSubFinder/pkg/sort_things"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/sirupsen/logrus"
)

// VideoNameSearchKeywordMaker 拼接视频搜索的 title 和 年份
func VideoNameSearchKeywordMaker(l *logrus.Logger, title string, year string) string {
	iYear, err := strconv.Atoi(year)
	if err != nil {
		// 允许的错误
		l.Errorln("VideoNameSearchKeywordMaker", "year to int", err)
		iYear = 0
	}
	searchKeyword := title
	if iYear >= 2020 {
		searchKeyword = searchKeyword + " " + year
	}

	return searchKeyword
}

// SearchMatchedVideoFileFromDirs 搜索符合后缀名的视频文件
func SearchMatchedVideoFileFromDirs(l *logrus.Logger, dirs []string) (*treemap.Map, error) {

	defer func() {
		l.Infoln("SearchMatchedVideoFileFromDirs End")
		l.Infoln(" --------------------------------------------------")
	}()
	l.Infoln(" --------------------------------------------------")
	l.Infoln("SearchMatchedVideoFileFromDirs Start...")

	var fileFullPathMap = treemap.NewWithStringComparator()
	for _, dir := range dirs {

		matchedVideoFile, err := SearchMatchedVideoFile(l, dir)
		if err != nil {
			return nil, err
		}
		value, found := fileFullPathMap.Get(dir)
		if found == false {
			fileFullPathMap.Put(dir, matchedVideoFile)
		} else {
			value = append(value.([]string), matchedVideoFile...)
			fileFullPathMap.Put(dir, value)
		}
	}

	fileFullPathMap.Each(func(seriesRootPathName interface{}, seriesNames interface{}) {

		oneSeriesRootPathName := seriesRootPathName.(string)
		fileFullPathList := seriesNames.([]string)
		// 排序，从最新的到最早的
		fileFullPathList = sort_things.SortByModTime(fileFullPathList)
		for _, s := range fileFullPathList {
			l.Debugln(s)
		}
		fileFullPathMap.Put(oneSeriesRootPathName, fileFullPathList)
	})

	return fileFullPathMap, nil
}

// SearchMatchedVideoFile 搜索符合后缀名的视频文件，现在也会把 BDMV 的文件搜索出来，但是这个并不是一个视频文件，需要在后续特殊处理
func SearchMatchedVideoFile(l *logrus.Logger, dir string) ([]string, error) {

	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 内层的错误就无视了
			oneList, _ := SearchMatchedVideoFile(l, fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			bok, fakeBDMVVideoFile := FileNameIsBDMV(fullPath)
			if bok == true {
				// 这类文件后续的扫描字幕操作需要额外的处理
				fileFullPathList = append(fileFullPathList, fakeBDMVVideoFile)
				continue
			}
			if IsWantedVideoExtDef(curFile.Name()) == false {
				// 不是期望的视频后缀名则跳过
				continue
			} else {
				// 这里还有一种情况，就是蓝光， BDMV 下面会有一个 STREAM 文件夹，里面很多 m2ts 的视频组成
				if filepath.Base(filepath.Dir(fullPath)) == "STREAM" {
					l.Debugln("SearchMatchedVideoFile, Skip BDMV.STREAM:", fullPath)
					continue
				}

				if filter.SkipFileInfo(l, curFile) == true {
					continue
				}

				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

func SearchTVNfo(l *logrus.Logger, dir string) ([]string, error) {

	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 内层的错误就无视了
			oneList, _ := SearchTVNfo(l, fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if strings.ToLower(curFile.Name()) != decode.MetadateTVNfo {
				continue
			} else {

				if filter.SkipFileInfo(l, curFile) == true {
					continue
				}
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// SearchSeriesAllEpsAndSubtitles 遍历这个连续剧目录下的所有视频文件，以及这个视频文件对应的字幕文件
func SearchSeriesAllEpsAndSubtitles(l *logrus.Logger, dir string) {

	pathVideoMap := make(map[string][]string, 0)
	pathSubsMap := make(map[string][]string, 0)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() == true {
			return nil
		}
		if IsWantedVideoExtDef(filepath.Ext(d.Name())) == true {
			// 如果是符合视频的后缀名，那么就缓存起来
			_, found := pathVideoMap[path]
			if found == false {
				pathVideoMap[path] = make([]string, 0)
			}
			pathVideoMap[path] = append(pathVideoMap[path], path)
			return nil
		}

		if sub_parser_hub.IsSubExtWanted(filepath.Ext(d.Name())) == true {
			// 如果是符合字幕的后缀名，那么就缓存起来
			_, found := pathSubsMap[path]
			if found == false {
				pathSubsMap[path] = make([]string, 0)
			}
			pathSubsMap[path] = append(pathSubsMap[path], path)
			return nil
		}

		return nil
	})
	if err != nil {
		return
	}
}
