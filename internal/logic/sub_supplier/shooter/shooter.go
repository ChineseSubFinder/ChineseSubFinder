package shooter

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/task_queue"
	pkgcommon "github.com/allanpk716/ChineseSubFinder/internal/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Supplier struct {
	settings *settings.Settings
	log      *logrus.Logger
	topic    int
	isAlive  bool
}

func NewSupplier(_settings *settings.Settings, _logger *logrus.Logger) *Supplier {

	sup := Supplier{}
	sup.log = _logger
	sup.topic = common2.DownloadSubsPerSite
	sup.isAlive = true // 默认是可以使用的，如果 check 后，再调整状态

	sup.settings = _settings
	if sup.settings.AdvancedSettings.Topic > 0 && sup.settings.AdvancedSettings.Topic != sup.topic {
		sup.topic = sup.settings.AdvancedSettings.Topic
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
	// 对于这个接口暂时没有限制
	return false
}

func (s *Supplier) GetSupplierName() string {
	return common2.SubSiteShooter
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

	if my_util.IsFile(filePath) == false {
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
		return nil, common2.ShooterFileHashIsEmpty
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

			data, _, err := my_util.DownFile(s.log, file.Link)
			if err != nil {
				s.log.Error(err)
				continue
			}
			// 下载成功需要统计到今天的次数中
			_, err = task_queue.AddDailyDownloadCount(s.GetSupplierName(),
				my_util.GetPublicIP(s.settings.AdvancedSettings.TaskQueue, s.settings.AdvancedSettings.ProxySettings))
			if err != nil {
				s.log.Warningln(s.GetSupplierName(), "getSubListFromFile.AddDailyDownloadCount", err)
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

func (s *Supplier) getSubInfos(fileHash, fileName, qLan string) ([]SublistShooter, error) {

	var jsonList []SublistShooter

	httpClient := my_util.NewHttpClient(s.settings.AdvancedSettings.ProxySettings)
	resp, err := httpClient.R().
		SetFormData(map[string]string{
			"filehash": fileHash,
			"pathinfo": fileName,
			"format":   "json",
			"lang":     qLan,
		}).
		SetResult(&jsonList).
		Post(s.settings.AdvancedSettings.SuppliersSettings.Shooter.RootUrl)
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
		return "", common2.VideoFileIsTooSmall
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
