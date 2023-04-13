package search

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/filter"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sort_things"
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

// MatchedVideoFileFromDirs 搜索符合后缀名的视频文件
func MatchedVideoFileFromDirs(l *logrus.Logger, dirs []string) (*treemap.Map, error) {

	defer func() {
		l.Infoln("MatchedVideoFileFromDirs End")
		l.Infoln(" --------------------------------------------------")
	}()
	l.Infoln(" --------------------------------------------------")
	l.Infoln("MatchedVideoFileFromDirs Start...")

	var fileFullPathMap = treemap.NewWithStringComparator()
	for _, dir := range dirs {

		matchedVideoFile, err := MatchedVideoFile(l, dir)
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

// MatchedVideoFile 搜索符合后缀名的视频文件，现在也会把 BDMV 的文件搜索出来，但是这个并不是一个视频文件，需要在后续特殊处理
func MatchedVideoFile(l *logrus.Logger, dir string) ([]string, error) {

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
			oneList, _ := MatchedVideoFile(l, fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			bok, fakeBDMVVideoFile := pkg.FileNameIsBDMV(fullPath)
			if bok == true {
				// 这类文件后续的扫描字幕操作需要额外的处理
				fileFullPathList = append(fileFullPathList, fakeBDMVVideoFile)
				continue
			}
			if pkg.IsWantedVideoExtDef(curFile.Name()) == false {
				// 不是期望的视频后缀名则跳过
				continue
			} else {
				// 这里还有一种情况，就是蓝光， BDMV 下面会有一个 STREAM 文件夹，里面很多 m2ts 的视频组成
				if filepath.Base(filepath.Dir(fullPath)) == "STREAM" {
					l.Debugln("MatchedVideoFile, Skip BDMV.STREAM:", fullPath)
					continue
				}

				if filter.SkipFileInfo(l, curFile, fullPath) == true {
					continue
				}

				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

func TVNfo(l *logrus.Logger, dir string) ([]string, error) {

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
			oneList, _ := TVNfo(l, fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if strings.ToLower(curFile.Name()) != decode.MetadateTVNfo {
				continue
			} else {

				//if filter.SkipFileInfo(l, curFile, fullPath) == true {
				//	continue
				//}
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// SeriesAllEpsAndSubtitles 遍历这个连续剧目录下的所有视频文件，以及这个视频文件对应的字幕文件，这里无法转换给出静态文件服务器的路径，需要额外再对回去到的信息进行处理
func SeriesAllEpsAndSubtitles(l *logrus.Logger, dir string) (*backend.SeasonInfo, error) {

	seasonInfo := backend.SeasonInfo{
		Name:          filepath.Base(dir),
		RootDirPath:   dir,
		OneVideoInfos: make([]backend.OneVideoInfo, 0),
	}
	pathVideoMap := make(map[string][]string, 0)
	pathSubsMap := make(map[string][]string, 0)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}
		if d.IsDir() == true {
			// 跳过文件夹
			return nil
		}

		if filter.SkipFileInfo(l, d, path) == true {
			return nil
		}

		if pkg.IsWantedVideoExtDef(d.Name()) == true {
			// 如果是符合视频的后缀名，那么就缓存起来
			tmpDir := filepath.Dir(path)
			_, found := pathVideoMap[tmpDir]
			if found == false {
				pathVideoMap[tmpDir] = make([]string, 0)
			}
			pathVideoMap[tmpDir] = append(pathVideoMap[tmpDir], path)
			return nil
		}

		if sub_parser_hub.IsSubExtWanted(d.Name()) == true {
			// 如果是符合字幕的后缀名，那么就缓存起来
			tmpDir := filepath.Dir(path)
			_, found := pathSubsMap[tmpDir]
			if found == false {
				pathSubsMap[tmpDir] = make([]string, 0)
			}
			pathSubsMap[tmpDir] = append(pathSubsMap[tmpDir], path)
			return nil
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	// 交叉比对，找到对应的字幕
	for pathKey, videos := range pathVideoMap {

		nowPathSubs, _ := pathSubsMap[pathKey]
		//if found == false {
		//	// 没有找到对应的字幕
		//	continue
		//}
		for _, oneVideo := range videos {

			videoName := strings.ReplaceAll(filepath.Base(oneVideo), filepath.Ext(oneVideo), "")

			skipInfo := models.NewSkipScanInfoBySeriesEx(oneVideo, true)

			if skipInfo.Season() == -1 || skipInfo.Eps() == -1 {
				// 无法解析的视频，跳过
				l.Errorln("SeriesAllEpsAndSubtitles, Skip UnParse Video:", oneVideo)
				continue
			}

			nowOneVideoInfo := backend.OneVideoInfo{
				Name:         filepath.Base(oneVideo),
				VideoFPath:   oneVideo,
				Season:       skipInfo.Season(),
				Episode:      skipInfo.Eps(),
				SubFPathList: make([]string, 0),
				SubUrlList:   make([]string, 0),
			}
			// 解析这个视频的 SxxExx 信息
			for _, oneSub := range nowPathSubs {

				if strings.HasPrefix(filepath.Base(oneSub), videoName) == false {
					continue
				}
				// 找到了对应的字幕
				nowOneVideoInfo.SubFPathList = append(nowOneVideoInfo.SubFPathList, oneSub)
			}

			seasonInfo.OneVideoInfos = append(seasonInfo.OneVideoInfos, nowOneVideoInfo)
		}
	}

	return &seasonInfo, nil
}
