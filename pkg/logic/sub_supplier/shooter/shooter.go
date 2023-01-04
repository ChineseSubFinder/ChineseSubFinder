package shooter

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/series"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/notify_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/sirupsen/logrus"
)

type Supplier struct {
	log            *logrus.Logger
	fileDownloader *file_downloader.FileDownloader
	topic          int
	isAlive        bool
}

func NewSupplier(fileDownloader *file_downloader.FileDownloader) *Supplier {

	sup := Supplier{}
	sup.log = fileDownloader.Log
	sup.fileDownloader = fileDownloader
	sup.topic = common.DownloadSubsPerSite
	sup.isAlive = true // 默认是可以使用的，如果 check 后，再调整状态

	if settings.Get().AdvancedSettings.Topic > 0 && settings.Get().AdvancedSettings.Topic != sup.topic {
		sup.topic = settings.Get().AdvancedSettings.Topic
	}

	return &sup
}

func (s *Supplier) CheckAlive() (bool, int64) {
	// 计算当前时间
	startT := time.Now()
	_, err := s.getSubInfos(checkFileHash, checkFileName, qLan)
	if err != nil {
		s.log.Errorln(s.GetSupplierName(), "CheckAlive", "Error", err)
		s.isAlive = false
		return false, 0
	}
	// 计算耗时
	s.isAlive = true
	return true, time.Since(startT).Milliseconds()
}

func (s *Supplier) IsAlive() bool {
	return s.isAlive
}

func (s *Supplier) OverDailyDownloadLimit() bool {

	if settings.Get().AdvancedSettings.SuppliersSettings.Shooter.DailyDownloadLimit == 0 {
		s.log.Warningln(s.GetSupplierName(), "DailyDownloadLimit is 0, will Skip Download")
		return true
	}
	// 对于这个接口暂时没有限制
	return false
}

func (s *Supplier) GetLogger() *logrus.Logger {
	return s.log
}

func (s *Supplier) GetSupplierName() string {
	return common.SubSiteShooter
}

func (s *Supplier) GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error) {
	return s.getSubListFromFile(filePath)
}

func (s *Supplier) GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	return s.downloadSub4Series(seriesInfo)
}

func (s *Supplier) getSubListFromFile(filePath string) ([]supplier.SubInfo, error) {

	defer func() {
		s.log.Debugln(s.GetSupplierName(), filePath, "End...")
	}()

	s.log.Debugln(s.GetSupplierName(), filePath, "Start...")

	// 可以提供的字幕查询 eng或者chn
	var outSubInfoList []supplier.SubInfo
	var jsonList []SublistShooter

	if pkg.IsFile(filePath) == false {
		// 这里传入的可能是蓝光结构的伪造存在的视频文件，需要检查一次这个文件是否存在
		bok, _, _ := decode.IsFakeBDMVWorked(filePath)
		if bok == false {

			nowError := errors.New(fmt.Sprintf("%s %s %s",
				s.GetSupplierName(),
				filePath,
				"not exist, and it`s not a Blue ray Video FakeFileName"))

			s.log.Errorln(nowError)

			return nil, nowError
		}
	}

	hash, err := ComputeFileHash(filePath)
	if err != nil {
		return nil, err
	}
	if hash == "" {
		return nil, common.ShooterFileHashIsEmpty
	}

	videoFileName := filepath.Base(filePath)
	jsonList, err = s.getSubInfos(hash, videoFileName, qLan)
	if err != nil {
		return nil, err
	}

	for i, shooter := range jsonList {
		for _, file := range shooter.Files {

			subInfo, err := s.fileDownloader.Get(s.GetSupplierName(), int64(i), videoFileName, file.Link, 0, shooter.Delay)
			if err != nil {
				s.log.Error("FileDownloader.Get", err)
				continue
			}

			outSubInfoList = append(outSubInfoList, *subInfo)
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

func (s *Supplier) getSubInfos(fileHash, fileName, qLan string) ([]SublistShooter, error) {

	var jsonList []SublistShooter

	httpClient, err := pkg.NewHttpClient()
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.R().
		SetFormData(map[string]string{
			"filehash": fileHash,
			"pathinfo": fileName,
			"format":   "json",
			"lang":     qLan,
		}).
		SetResult(&jsonList).
		Post(settings.Get().AdvancedSettings.SuppliersSettings.Shooter.RootUrl)
	if err != nil {
		if resp != nil {
			s.log.Errorln(s.GetSupplierName(), "NewHttpClient:", fileName, err.Error())
			notify_center.Notify.Add(s.GetSupplierName()+" NewHttpClient", fmt.Sprintf("filePath: %s, resp: %s, error: %s", fileName, resp.String(), err.Error()))
		}
		return nil, err
	}

	return jsonList, nil
}

func ComputeFileHash(filePath string) (string, error) {
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

func (s *Supplier) downloadSub4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error) {
	var allSupplierSubInfo = make([]supplier.SubInfo, 0)

	index := 0
	// 这里拿到的 seriesInfo ，里面包含了，需要下载字幕的 Eps 信息
	for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		index++
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
