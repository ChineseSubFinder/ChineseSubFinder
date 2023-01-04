package movie_helper

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/media_info_dealers"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ifaces"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/imdb_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/jinzhu/now"
	"github.com/sirupsen/logrus"
)

// OneMovieDlSubInAllSite 一部电影在所有的网站下载相应的字幕
func OneMovieDlSubInAllSite(logger *logrus.Logger, Suppliers []ifaces.ISupplier, oneVideoFullPath string, i int64) []supplier.SubInfo {

	defer func() {
		logger.Infoln(common.QueueName, i, "DlSub End", oneVideoFullPath)
	}()

	var outSUbInfos = make([]supplier.SubInfo, 0)
	logger.Infoln(common.QueueName, i, "DlSub Start", oneVideoFullPath)
	for _, oneSupplier := range Suppliers {

		logger.Infoln(common.QueueName, i, oneSupplier.GetSupplierName(), oneVideoFullPath)

		if oneSupplier.OverDailyDownloadLimit() == true {
			logger.Infoln(common.QueueName, i, oneSupplier.GetSupplierName(), "Over Daily Download Limit")
			continue
		}

		subInfos, err := OneMovieDlSubInOneSite(logger, oneVideoFullPath, i, oneSupplier)
		if err != nil {
			logger.Errorln(common.QueueName, i, oneSupplier.GetSupplierName(), "oneMovieDlSubInOneSite", err)
			continue
		}
		outSUbInfos = append(outSUbInfos, subInfos...)
	}

	for index, info := range outSUbInfos {
		logger.Debugln(common.QueueName, i, "OneMovieDlSubInAllSite get sub", index, "Name:", info.Name, "FileUrl:", info.FileUrl)
	}

	return outSUbInfos
}

// OneMovieDlSubInOneSite 一部电影在一个站点下载字幕
func OneMovieDlSubInOneSite(logger *logrus.Logger, oneVideoFullPath string, i int64, supplier ifaces.ISupplier) ([]supplier.SubInfo, error) {
	defer func() {
		logger.Infoln(common.QueueName, i, supplier.GetSupplierName(), "End...")
	}()
	logger.Infoln(common.QueueName, i, supplier.GetSupplierName(), "Start...")
	subInfos, err := supplier.GetSubListFromFile4Movie(oneVideoFullPath)
	if err != nil {
		return nil, err
	}
	// 把后缀名给改好
	sub_helper.ChangeVideoExt2SubExt(subInfos)

	return subInfos, nil
}

// MovieHasChineseSub 这个视频文件的目录下面有字幕文件了没有
func MovieHasChineseSub(logger *logrus.Logger, videoFilePath string) (bool, []string, []string, error) {
	dir := filepath.Dir(videoFilePath)
	videoFileName := filepath.Base(videoFilePath)
	videoFileName = strings.ReplaceAll(videoFileName, filepath.Ext(videoFileName), "")
	files, err := os.ReadDir(dir)
	if err != nil {
		return false, nil, nil, err
	}
	// 所有的中文字幕列表
	var chineseSubFullPathList = make([]string, 0)
	// 所有的中文字幕列表，需要文件名与视频名称一样，也就是 Sub 文件半酣 Video name 即可
	var chineseSubFitVideoNameFullPathList = make([]string, 0)
	bFoundChineseSub := false
	for _, curFile := range files {
		if curFile.IsDir() {
			continue
		} else {
			// 文件
			if sub_parser_hub.IsSubExtWanted(curFile.Name()) == false {
				continue
			}
			// 字幕文件是否包含中文
			subFileFullPath := filepath.Join(dir, curFile.Name())
			subParserHub := sub_parser_hub.NewSubParserHub(logger, ass.NewParser(logger), srt.NewParser(logger))
			bFind, subParserFileInfo, err := subParserHub.DetermineFileTypeFromFile(subFileFullPath)
			if err != nil {
				logger.Errorln("DetermineFileTypeFromFile", subFileFullPath, err)
				continue
			}
			if bFind == false {
				logger.Warnln("DetermineFileTypeFromFile", subFileFullPath, "not support SubType")
				continue
			}
			if subParserHub.IsSubHasChinese(subParserFileInfo) == true {
				if bFoundChineseSub == false {
					bFoundChineseSub = true
				}
				chineseSubFullPathList = append(chineseSubFullPathList, subFileFullPath)

				if strings.Contains(curFile.Name(), videoFileName) == true {
					chineseSubFitVideoNameFullPathList = append(chineseSubFitVideoNameFullPathList, subFileFullPath)
				}
			}
		}
	}

	return bFoundChineseSub, chineseSubFullPathList, chineseSubFitVideoNameFullPathList, nil
}

// SkipChineseMovie 跳过中文的电影
func SkipChineseMovie(dealers *media_info_dealers.Dealers, videoFullPath string) (bool, error) {

	imdbInfo, err := decode.GetVideoNfoInfo4Movie(videoFullPath)
	if err != nil {
		return false, err
	}
	isChineseVideo, _, err := imdb_helper.IsChineseVideo(dealers, imdbInfo)
	if err != nil {
		return false, err
	}
	if isChineseVideo == true {
		dealers.Logger.Infoln("Skip", videoFullPath, "Sub Download, because movie is Chinese")
		return true, nil
	} else {
		return false, nil
	}
}

func MovieNeedDlSub(logger *logrus.Logger, videoFullPath string, ExpirationTime int) (bool, error) {
	// 视频下面有不有字幕
	found, _, _, err := MovieHasChineseSub(logger, videoFullPath)
	if err != nil {
		return false, err
	}
	// 资源下载的时间后的多少天内都进行字幕的自动下载，替换原有的字幕
	currentTime := time.Now()
	videoNfoInfo4Movie, modifyTime, err := decode.GetVideoInfoFromFileFullPath(videoFullPath, true)
	if err != nil {
		return false, err
	}
	// 如果这个视频发布的时间早于现在有两个年的间隔
	if videoNfoInfo4Movie.GetYear() > 0 && currentTime.Year()-2 > videoNfoInfo4Movie.GetYear() {
		if found == false {
			// 需要下载的
			return true, nil
		} else {
			// 有字幕了，没必要每次都刷新，跳过
			logger.Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because movie has sub and published more than 2 years")
			return false, nil
		}
	} else {
		// 如果播出时间能够读取到，那么就以这个完后推算 3个月
		// 如果读取不到 Aired Time 那么，下载后的 ModifyTime 3个月天内，都进行字幕的下载
		var baseTime time.Time
		if videoNfoInfo4Movie.ReleaseDate != "" {
			baseTime, err = now.Parse(videoNfoInfo4Movie.ReleaseDate)
			if err != nil {
				logger.Errorln("Movie parse AiredTime", err)
				baseTime = modifyTime
			}
		} else {
			baseTime = modifyTime
		}

		// 3个月内，或者没有字幕都要进行下载
		if baseTime.AddDate(0, 0, ExpirationTime).After(currentTime) == true || found == false {
			// 需要下载的
			return true, nil
		} else {
			if baseTime.AddDate(0, 0, ExpirationTime).After(currentTime) == false {
				logger.Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because movie has sub and downloaded or aired more than 3 months")
				return false, nil
			}
			if found == true {
				logger.Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because sub file found")
				return false, nil
			}

			return false, nil
		}
	}
}
