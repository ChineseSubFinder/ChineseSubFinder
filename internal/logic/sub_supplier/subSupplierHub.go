package sub_supplier

import (
	_interface2 "github.com/allanpk716/ChineseSubFinder/internal/interface"
	movie_helper2 "github.com/allanpk716/ChineseSubFinder/internal/logic/movie_helper"
	series_helper2 "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/sirupsen/logrus"
	"gopkg.in/errgo.v2/fmt/errors"
	"path/filepath"
)

type SubSupplierHub struct {
	Suppliers []_interface2.ISupplier

	log *logrus.Logger
}

func NewSubSupplierHub(one _interface2.ISupplier,_inSupplier ..._interface2.ISupplier) *SubSupplierHub {
	s := SubSupplierHub{}
	s.log = pkg.GetLogger()
	s.Suppliers = make([]_interface2.ISupplier, 0)
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

	// 跳过中文的电影，不是一定要跳过的
	skip, err := movie_helper2.SkipChineseMovie(videoFullPath, d.Suppliers[0].GetReqParam())
	if err != nil {
		d.log.Warnln("SkipChineseMovie", videoFullPath, err)
	}
	if skip == true {
		return nil, nil
	}

	needDlSub, err := movie_helper2.MovieNeedDlSub(videoFullPath)
	if err != nil {
		return nil, errors.Newf("MovieNeedDlSub %v %v", videoFullPath , err)
	}
	if needDlSub == true {
		// 需要下载字幕
		// 下载所有字幕
		subInfos := movie_helper2.OneMovieDlSubInAllSite(d.Suppliers, videoFullPath, index)
		// 整理字幕，比如解压什么的
		organizeSubFiles, err := pkg.OrganizeDlSubFiles(filepath.Base(videoFullPath), subInfos)
		if err != nil {
			return nil, errors.Newf("OrganizeDlSubFiles %v %v", videoFullPath , err)
		}
		// 因为是下载电影，需要合并返回
		var outSubFileFullPathList = make([]string, 0)
		for s, _ := range organizeSubFiles {
			outSubFileFullPathList = append(outSubFileFullPathList, organizeSubFiles[s]...)
		}
		return outSubFileFullPathList, nil
	} else {
		// 无需下载字幕
		return nil, nil
	}
}

// DownloadSub4Series 某一部连续剧的字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub4Series(seriesDirPath string, index int) (*series.SeriesInfo, map[string][]string, error) {

	// 跳过中文的连续剧，不是一定要跳过的
	skip, imdbInfo, err := series_helper2.SkipChineseSeries(seriesDirPath, d.Suppliers[0].GetReqParam())
	if err != nil {
		d.log.Warnln("SkipChineseSeries", seriesDirPath, err)
	}
	if skip == true {
		return nil, nil, nil
	}
	// 读取本地的视频和字幕信息
	seriesInfo, err := series_helper2.ReadSeriesInfoFromDir(seriesDirPath, imdbInfo)
	if err != nil {
		return nil, nil, errors.Newf("ReadSeriesInfoFromDir %v %v", seriesDirPath , err)
	}
	organizeSubFiles, err := d.dlSubFromSeriesInfo(seriesDirPath, index, seriesInfo, err)
	if err != nil {
		return nil, nil, err
	}
	return seriesInfo, organizeSubFiles, nil
}

// DownloadSub4SeriesFromEmby 通过 Emby 查询到的信息进行字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub4SeriesFromEmby(seriesDirPath string, seriesList []emby.EmbyMixInfo, index int) (*series.SeriesInfo, map[string][]string, error) {

	// 跳过中文的连续剧，不是一定要跳过的
	skip, imdbInfo, err := series_helper2.SkipChineseSeries(seriesDirPath, d.Suppliers[0].GetReqParam())
	if err != nil {
		d.log.Warnln("SkipChineseSeries", seriesDirPath, err)
	}
	if skip == true {
		return nil, nil, nil
	}
	// 读取本地的视频和字幕信息
	seriesInfo, err := series_helper2.ReadSeriesInfoFromEmby(seriesDirPath, imdbInfo, seriesList)
	if err != nil {
		return nil, nil, errors.Newf("ReadSeriesInfoFromDir %v %v", seriesDirPath , err)
	}
	organizeSubFiles, err := d.dlSubFromSeriesInfo(seriesDirPath, index, seriesInfo, err)
	if err != nil {
		return nil, nil, err
	}
	return seriesInfo, organizeSubFiles, nil
}

func (d SubSupplierHub) dlSubFromSeriesInfo(seriesDirPath string, index int, seriesInfo *series.SeriesInfo, err error) (map[string][]string, error) {
	// 下载好的字幕
	subInfos := series_helper2.OneSeriesDlSubInAllSite(d.Suppliers, seriesInfo, index)
	// 整理字幕，比如解压什么的
	// 每一集 SxEx - 对应解压整理后的字幕列表
	organizeSubFiles, err := pkg.OrganizeDlSubFiles(filepath.Base(seriesDirPath), subInfos)
	if err != nil {
		return nil, errors.Newf("OrganizeDlSubFiles %v %v", seriesDirPath, err)
	}
	return organizeSubFiles, nil
}