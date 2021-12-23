package shooter

import (
	"crypto/md5"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Supplier struct {
	reqParam types.ReqParam
	log      *logrus.Logger
	topic    int
}

func NewSupplier(_reqParam ...types.ReqParam) *Supplier {

	sup := Supplier{}
	sup.log = log_helper.GetLogger()
	sup.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		sup.reqParam = _reqParam[0]
		if sup.reqParam.Topic > 0 && sup.reqParam.Topic != sup.topic {
			sup.topic = sup.reqParam.Topic
		}
	}
	return &sup
}

func (s Supplier) GetSupplierName() string {
	return common.SubSiteShooter
}

func (s Supplier) GetReqParam() types.ReqParam {
	return s.reqParam
}

func (s Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {
	return s.getSubListFromFile(filePath)
}

func (s Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s Supplier) getSubListFromFile(filePath string) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), filePath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), filePath, "Start...")

	// 可以提供的字幕查询 eng或者chn
	const qLan = "Chn"
	var outSubInfoList []supplier.SubInfo
	var jsonList []SublistShooter

	hash, err := s.computeFileHash(filePath)
	if err != nil {
		return nil, err
	}
	if hash == "" {
		return nil, common.ShooterFileHashIsEmpty
	}

	fileName := filepath.Base(filePath)

	httpClient := my_util.NewHttpClient(s.reqParam)
	resp, err := httpClient.R().
		SetFormData(map[string]string{
			"filehash": hash,
			"pathinfo": fileName,
			"format":   "json",
			"lang":     qLan,
		}).
		SetResult(&jsonList).
		Post(common.SubShooterRootUrl)
	if err != nil {
		if resp != nil {
			s.log.Errorln(s.GetSupplierName(), "NewHttpClient:", filePath, err.Error())
			notify_center.Notify.Add(s.GetSupplierName()+" NewHttpClient", fmt.Sprintf("filePath: %s, resp: %s, error: %s", filePath, resp.String(), err.Error()))
		}
		return nil, err
	}
	for i, shooter := range jsonList {
		for _, file := range shooter.Files {
			subExt := file.Ext
			if strings.Contains(file.Ext, ".") == false {
				subExt = "." + subExt
			}

			data, _, err := my_util.DownFile(file.Link)
			if err != nil {
				s.log.Error(err)
				continue
			}

			onSub := supplier.NewSubInfo(s.GetSupplierName(), int64(i), fileName, language.ChineseSimple, file.Link, 0, shooter.Delay, subExt, data)
			outSubInfoList = append(outSubInfoList, *onSub)
			// 如果够了那么多个字幕就返回
			if len(outSubInfoList) >= s.topic {
				return outSubInfoList, nil
			}
			// 一层里面，下载一个文件就行了
			break
		}
	}
	return outSubInfoList, nil
}

func (s Supplier) getSubListFromKeyword(keyword string) ([]supplier.SubInfo, error) {
	panic("not implemented")
}

func (s Supplier) computeFileHash(filePath string) (string, error) {
	hash := ""
	fp, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fp.Close()
	stat, err := fp.Stat()
	if err != nil {
		return "", err
	}
	size := float64(stat.Size())
	if size < 0xF000 {
		return "", common.VideoFileIsTooSmall
	}
	samplePositions := [4]int64{
		4 * 1024,
		int64(math.Floor(size / 3 * 2)),
		int64(math.Floor(size / 3)),
		int64(size - 8*1024)}
	var samples [4][]byte
	for i, position := range samplePositions {
		samples[i] = make([]byte, 4*1024)
		_, err = fp.ReadAt(samples[i], position)
		if err != nil {
			return "", err
		}
	}
	for _, sample := range samples {
		if len(hash) > 0 {
			hash += ";"
		}
		hash += fmt.Sprintf("%x", md5.Sum(sample))
	}

	return hash, nil
}

func (s Supplier) downloadSub4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	var allSupplierSubInfo = make([]supplier.SubInfo, 0)
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {
		one, err := s.getSubListFromFile(episodeInfo.FileFullPath)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "getSubListFromFile", episodeInfo.FileFullPath)
			continue
		}
		// 需要赋值给字幕结构
		for i, _ := range one {
			one[i].Season = episodeInfo.Season
			one[i].Episode = episodeInfo.Episode
		}
		allSupplierSubInfo = append(allSupplierSubInfo, one...)
	}
	// 返回前，需要把每一个 Eps 的 Season Episode 信息填充到每个 SubInfo 中
	return allSupplierSubInfo, nil
}

type FilesShooter struct {
	Ext  string `json:"ext"`
	Link string `json:"link"`
}
type SublistShooter struct {
	Desc  string         `json:"desc"`
	Delay int64          `json:"delay"`
	Files []FilesShooter `json:"files"`
}
