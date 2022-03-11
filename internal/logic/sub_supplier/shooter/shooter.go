package shooter

import (
	"crypto/md5"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	pkgcommon "github.com/allanpk716/ChineseSubFinder/internal/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/huandu/go-clone"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Supplier struct {
	settings settings.Settings
	log      *logrus.Logger
	topic    int
}

func NewSupplier(_settings settings.Settings) *Supplier {

	sup := Supplier{}
	sup.log = log_helper.GetLogger()
	sup.topic = common.DownloadSubsPerSite

	sup.settings = clone.Clone(_settings).(settings.Settings)
	if sup.settings.AdvancedSettings.Topic > 0 && sup.settings.AdvancedSettings.Topic != sup.topic {
		sup.topic = sup.settings.AdvancedSettings.Topic
	}

	return &sup
}

func (s Supplier) CheckAlive() (bool, int64) {
	// 计算当前时间
	startT := time.Now()
	_, err := s.getSubInfos(checkFileHash, checkFileName, qLan)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Error", err)
		return false, 0
	}
	// 计算耗时
	return true, time.Since(startT).Milliseconds()
}

func (s Supplier) GetSupplierName() string {
	return common.SubSiteShooter
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
	jsonList, err = s.getSubInfos(hash, fileName, qLan)
	if err != nil {
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

func (s Supplier) getSubInfos(fileHash, fileName, qLan string) ([]SublistShooter, error) {

	var jsonList []SublistShooter

	httpClient := my_util.NewHttpClient(*s.settings.AdvancedSettings.ProxySettings)
	resp, err := httpClient.R().
		SetFormData(map[string]string{
			"filehash": fileHash,
			"pathinfo": fileName,
			"format":   "json",
			"lang":     qLan,
		}).
		SetResult(&jsonList).
		Post(common.SubShooterRootUrl)
	if err != nil {
		if resp != nil {
			s.log.Errorln(s.GetSupplierName(), "NewHttpClient:", fileName, err.Error())
			notify_center.Notify.Add(s.GetSupplierName()+" NewHttpClient", fmt.Sprintf("filePath: %s, resp: %s, error: %s", fileName, resp.String(), err.Error()))
		}
		return nil, err
	}

	return jsonList, nil
}

func (s Supplier) computeFileHash(filePath string) (string, error) {
	hash := ""
	fp, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = fp.Close()
	}()
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

	index := 0
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		index++
		pkgcommon.SetSubScanJobStatusScanSeriesSub(index, len(seriesInfo.NeedDlEpsKeyList),
			fmt.Sprintf("%v - S%v-E%v", episodeInfo.Title, episodeInfo.Season, episodeInfo.Episode))

		one, err := s.getSubListFromFile(episodeInfo.FileFullPath)
		if err != nil {
			s.log.Errorln(s.GetSupplierName(), "getSubListFromFile", episodeInfo.FileFullPath)
			continue
		}
		if one == nil {
			// 没有搜索到字幕
			s.log.Infoln(s.GetSupplierName(), "Not Find Sub can be download",
				episodeInfo.Title, episodeInfo.Season, episodeInfo.Episode)
			continue
		}
		// 需要赋值给字幕结构
		for i := range one {
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

const (
	qLan          = "Chn"
	checkFileHash = "234b0ff3685d6c46164b6b48cd39d69f;8be57624909f9d365dc81df43399d496;436de72e3c36a05a07875cc3249ae31a;237f498cfee89c67a22564e61047b053"
	checkFileName = "S05E09.mkv"
)
