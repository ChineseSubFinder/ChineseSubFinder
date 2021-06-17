package sub_supplier

import (
	"github.com/allanpk716/ChineseSubFinder/interface"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/allanpk716/ChineseSubFinder/movie_helper"
	"github.com/allanpk716/ChineseSubFinder/series_helper"
	"github.com/sirupsen/logrus"
)

type SubSupplierHub struct {
	Suppliers []_interface.ISupplier
	log *logrus.Logger
}

func NewSubSupplierHub(one _interface.ISupplier,_inSupplier ..._interface.ISupplier) *SubSupplierHub {
	s := SubSupplierHub{}
	s.log = model.GetLogger()
	s.Suppliers = make([]_interface.ISupplier, 0)
	s.Suppliers = append(s.Suppliers, one)
	if len(_inSupplier) > 0 {
		for _, supplier := range _inSupplier {
			s.Suppliers = append(s.Suppliers, supplier)
		}
	}
	return &s
}

// DownloadSub4Movie 某一个电影字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub4Movie(videoFullPath string, index int) ([]string, error) {
	// 先清理缓存文件夹
	err := model.ClearTmpFolder()
	if err != nil {
		d.log.Error(err)
	}
	// 跳过中文的电影，不是一定要跳过的
	skip, err := movie_helper.SkipChineseMovie(videoFullPath, d.Suppliers[0].GetReqParam())
	if err != nil {
		d.log.Error("SkipChineseMovie", err)
	}
	if skip == true {
		return nil, nil
	}

	var organizeSubFiles []string
	needDlSub, err := movie_helper.MovieNeedDlSub(videoFullPath)
	if err != nil {
		return nil, err
	}
	if needDlSub == true {
		// 需要下载字幕
		// 下载所有字幕
		subInfos := movie_helper.OneMovieDlSubInAllSite(d.Suppliers, videoFullPath, index)
		// 整理字幕，比如解压什么的
		organizeSubFiles, err = model.OrganizeDlSubFiles(subInfos)
		if err != nil {
			return nil, err
		}
		return organizeSubFiles, nil
	} else {
		// 无需下载字幕
		return nil, nil
	}
}

// DownloadSub4Series 某一个视频的字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub4Series(seriesDirPath string, index int) ([]string, error) {
	// 先清理缓存文件夹
	err := model.ClearTmpFolder()
	if err != nil {
		d.log.Error(err)
	}
	// 跳过中文的连续剧，不是一定要跳过的
	skip, err := series_helper.SkipChineseSeries(seriesDirPath, d.Suppliers[0].GetReqParam())
	if err != nil {
		d.log.Error("SkipChineseSeries", err)
	}
	if skip == true {
		return nil, nil
	}
	// 读取本地的视频和字幕信息
	seriesInfo, err := series_helper.ReadSeriesInfoFromDir(seriesDirPath)
	if err != nil {
		return nil, err
	}
	// 下载好的字幕
	subInfos := series_helper.OneSeriesDlSubInAllSite(d.Suppliers, seriesInfo)
	// 整理字幕，比如解压什么的
	organizeSubFiles, err := model.OrganizeDlSubFiles(subInfos)

	return organizeSubFiles, nil
}