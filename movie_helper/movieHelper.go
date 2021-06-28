package movie_helper

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	_interface "github.com/allanpk716/ChineseSubFinder/interface"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/jinzhu/now"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

// OneMovieDlSubInAllSite 一部电影在所有的网站下载相应的字幕
func OneMovieDlSubInAllSite(Suppliers []_interface.ISupplier, oneVideoFullPath string, i int) []common.SupplierSubInfo {
	var outSUbInfos = make([]common.SupplierSubInfo, 0)
	// 同时进行查询
	subInfosChannel := make(chan []common.SupplierSubInfo)
	model.GetLogger().Infoln("DlSub Start", oneVideoFullPath)
	for _, supplier := range Suppliers {
		supplier := supplier
		go func() {
			subInfos, err := OneMovieDlSubInOneSite(oneVideoFullPath, i, supplier)
			if err != nil {
				model.GetLogger().Errorln(supplier.GetSupplierName(), "oneMovieDlSubInOneSite", err)
			}
			subInfosChannel <- subInfos
		}()
	}
	for i := 0; i < len(Suppliers); i++ {
		v, ok := <-subInfosChannel
		if ok == true {
			outSUbInfos = append(outSUbInfos, v...)
		}
	}
	model.GetLogger().Infoln("DlSub End", oneVideoFullPath)
	return outSUbInfos
}

// OneMovieDlSubInOneSite 一部电影在一个站点下载字幕
func OneMovieDlSubInOneSite(oneVideoFullPath string, i int, supplier _interface.ISupplier) ([]common.SupplierSubInfo, error) {
	defer func() {
		model.GetLogger().Infoln(i, supplier.GetSupplierName(), "End...")
	}()
	model.GetLogger().Infoln(i, supplier.GetSupplierName(), "Start...")
	subInfos, err := supplier.GetSubListFromFile4Movie(oneVideoFullPath)
	if err != nil {
		return nil, err
	}
	// 把后缀名给改好
	model.ChangeVideoExt2SubExt(subInfos)

	return subInfos, nil
}

// MovieHasSub 这个视频文件的目录下面有字幕文件了没有
func MovieHasSub(videoFilePath string) (bool, error) {
	dir := filepath.Dir(videoFilePath)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, curFile := range files {
		if curFile.IsDir() {
			continue
		} else {
			// 文件
			if model.IsSubExtWanted(curFile.Name()) == true {
				return true, nil
			}
		}
	}

	return false, nil
}

func SkipChineseMovie(videoFullPath string, _reqParam ...common.ReqParam) (bool, error) {
	var reqParam common.ReqParam
	if len(_reqParam) > 0 {
		reqParam = _reqParam[0]
	}
	imdbInfo, err := model.GetImdbInfo4Movie(videoFullPath)
	if err != nil {
		return false, err
	}
	t, err := model.GetVideoInfoFromIMDB(imdbInfo.ImdbId, reqParam)
	if err != nil {
		return false, err
	}
	if len(t.Languages) > 0 && strings.ToLower(t.Languages[0]) == "chinese" {
		model.GetLogger().Infoln("Skip", videoFullPath, "Sub Download, because movie is Chinese")
		return true, nil
	}
	return false, nil
}

func MovieNeedDlSub(videoFullPath string) (bool, error) {
	// 视频下面有不有字幕
	found, err := MovieHasSub(videoFullPath)
	if err != nil {
		return false, err
	}
	// 资源下载的时间后的多少天内都进行字幕的自动下载，替换原有的字幕
	currentTime := time.Now()
	dayRange, _ := time.ParseDuration(common.DownloadSubDuring3Months)
	mInfo, modifyTime, err := model.GetVideoInfoFromFileFullPath(videoFullPath)
	if err != nil {
		return false, err
	}
	// 如果这个视频发布的时间早于现在有两个年的间隔
	if mInfo.Year > 0 &&  currentTime.Year() - 2 > mInfo.Year {
		if found == false {
			// 需要下载的
			return true, nil
		} else {
			// 有字幕了，没必要每次都刷新，跳过
			model.GetLogger().Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because movie has sub and published more than 2 years")
			return false, nil
		}
	} else {
		// 读取不到 IMDB 信息也能接受
		videoIMDBInfo, err := model.GetImdbInfo4Movie(videoFullPath)
		if err != nil {
			model.GetLogger().Errorln("MovieNeedDlSub.GetImdbInfo4Movie", err)
		}
		// 如果播出时间能够读取到，那么就以这个完后推算 3个月
		// 如果读取不到 Aired Time 那么，下载后的 ModifyTime 3个月天内，都进行字幕的下载
		var baseTime time.Time
		if videoIMDBInfo.ReleaseDate != "" {
			baseTime, err = now.Parse(videoIMDBInfo.ReleaseDate)
			if err != nil {
				model.GetLogger().Errorln("Movie parse AiredTime", err)
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
				model.GetLogger().Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because movie has sub and downloaded or aired more than 3 months")
				return false, nil
			}
			if found == true {
				model.GetLogger().Infoln("Skip", filepath.Base(videoFullPath), "Sub Download, because sub file found")
				return false, nil
			}

			return false, nil
		}
	}
}
