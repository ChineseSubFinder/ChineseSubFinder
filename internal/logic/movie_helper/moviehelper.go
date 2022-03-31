package movie_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/imdb_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/jinzhu/now"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OneMovieDlSubInAllSite 一部电影在所有的网站下载相应的字幕
func OneMovieDlSubInAllSite(Suppliers []ifaces.ISupplier, oneVideoFullPath string, i int) []supplier.SubInfo {

	defer func() {
		log_helper.GetLogger().Infoln(common.QueueName, i, "DlSub End", oneVideoFullPath)
	}()

	var outSUbInfos = make([]supplier.SubInfo, 0)
	// TODO 资源占用较高，把这里的并发给取消
	//// 同时进行查询
	//subInfosChannel := make(chan []supplier.SubInfo)
	//log_helper.GetLogger().Infoln(common.QueueName, i, "DlSub Start", oneVideoFullPath)
	//for _, oneSupplier := range Suppliers {
	//	nowSupplier := oneSupplier
	//	go func() {
	//		subInfos, err := OneMovieDlSubInOneSite(oneVideoFullPath, i, nowSupplier)
	//		if err != nil {
	//			log_helper.GetLogger().Errorln(common.QueueName, i, nowSupplier.GetSupplierName(), "oneMovieDlSubInOneSite", err)
	//		}
	//		subInfosChannel <- subInfos
	//	}()
	//}
	//for index := 0; index < len(Suppliers); index++ {
	//	v, ok := <-subInfosChannel
	//	if ok == true && v != nil {
	//		outSUbInfos = append(outSUbInfos, v...)
	//	}
	//}

	log_helper.GetLogger().Infoln(common.QueueName, i, "DlSub Start", oneVideoFullPath)
	for _, oneSupplier := range Suppliers {

		log_helper.GetLogger().Infoln(common.QueueName, i, oneSupplier.GetSupplierName(), oneVideoFullPath)

		subInfos, err := OneMovieDlSubInOneSite(oneVideoFullPath, i, oneSupplier)
		if err != nil {
			log_helper.GetLogger().Errorln(common.QueueName, i, oneSupplier.GetSupplierName(), "oneMovieDlSubInOneSite", err)
			continue
		}
		outSUbInfos = append(outSUbInfos, subInfos...)
	}

	for index, info := range outSUbInfos {
		log_helper.GetLogger().Debugln(common.QueueName, i, "OneMovieDlSubInAllSite get sub", index, "Name:", info.Name, "FileUrl:", info.FileUrl)
	}

	return outSUbInfos
}

// OneMovieDlSubInOneSite 一部电影在一个站点下载字幕
func OneMovieDlSubInOneSite(oneVideoFullPath string, i int, supplier ifaces.ISupplier) ([]supplier.SubInfo, error) {
	defer func() {
		log_helper.GetLogger().Infoln(common.QueueName, i, supplier.GetSupplierName(), "End...")
	}()
	log_helper.GetLogger().Infoln(common.QueueName, i, supplier.GetSupplierName(), "Start...")
	subInfos, err := supplier.GetSubListFromFile4Movie(oneVideoFullPath)
	if err != nil {
		return nil, err
	}
	// 把后缀名给改好
	sub_helper.ChangeVideoExt2SubExt(subInfos)

	return subInfos, nil
}

// MovieHasChineseSub 这个视频文件的目录下面有字幕文件了没有
func MovieHasChineseSub(videoFilePath string) (bool, []string, []string, error) {
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
			if sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser()).IsSubHasChinese(subFileFullPath) == true {
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
func SkipChineseMovie(videoFullPath string, _proxySettings ...settings.ProxySettings) (bool, error) {

	var proxySettings settings.ProxySettings
	if len(_proxySettings) > 0 {
		proxySettings = _proxySettings[0]
	}
	imdbInfo, err := decode.GetImdbInfo4Movie(videoFullPath)
	if err != nil {
		return false, err
	}
	isChineseVideo, _, err := imdb_helper.IsChineseVideo(imdbInfo.ImdbId, proxySettings)
	if err != nil {
		return false, err
	}
	if isChineseVideo == true {
		log_helper.GetLogger().Infoln("Skip", videoFullPath, "Sub Download, because movie is Chinese")
		return true, nil
	} else {
		return false, nil
	}
}

func MovieNeedDlSub(videoFullPath string) (bool, error) {
	// 视频下面有不有字幕
	found, _, _, err := MovieHasChineseSub(videoFullPath)
	if err != nil {
		return false, err
	}
	// 资源下载的时间后的多少天内都进行字幕的自动下载，替换原有的字幕
	currentTime := time.Now()
	dayRange, _ := time.ParseDuration(common.DownloadSubDuring3Months)
	mInfo, modifyTime, err := decode.GetVideoInfoFromFileFullPath(videoFullPath)
	if err != nil {
		return false, err
	}
	// 如果这个视频发布的时间早于现在有两个年的间隔
	if mInfo.Year > 0 && currentTime.Year()-2 > mInfo.Year {
		if found == false {
			// 需要下载的
			return true, nil
		} else {
			// 有字幕了，没必要每次都刷新，跳过
			log_helper.GetLogger().Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because movie has sub and published more than 2 years")
			return false, nil
		}
	} else {
		// 读取不到 IMDB 信息也能接受
		videoIMDBInfo, err := decode.GetImdbInfo4Movie(videoFullPath)
		if err != nil {
			log_helper.GetLogger().Errorln("MovieNeedDlSub.GetImdbInfo4Movie", err)
		}
		// 如果播出时间能够读取到，那么就以这个完后推算 3个月
		// 如果读取不到 Aired Time 那么，下载后的 ModifyTime 3个月天内，都进行字幕的下载
		var baseTime time.Time
		if videoIMDBInfo.ReleaseDate != "" {
			baseTime, err = now.Parse(videoIMDBInfo.ReleaseDate)
			if err != nil {
				log_helper.GetLogger().Errorln("Movie parse AiredTime", err)
				baseTime = modifyTime
			}
		} else {
			baseTime = modifyTime
		}

		// 3个月内，或者没有字幕都要进行下载
		if baseTime.Add(dayRange).After(currentTime) == true || found == false {
			// 需要下载的
			return true, nil
		} else {
			if baseTime.Add(dayRange).After(currentTime) == false {
				log_helper.GetLogger().Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because movie has sub and downloaded or aired more than 3 months")
				return false, nil
			}
			if found == true {
				log_helper.GetLogger().Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because sub file found")
				return false, nil
			}

			return false, nil
		}
	}
}
