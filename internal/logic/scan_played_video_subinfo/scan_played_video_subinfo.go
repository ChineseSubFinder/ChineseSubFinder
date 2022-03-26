package scan_played_video_subinfo

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/imdb_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/task_control"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"sync"
)

type ScanPlayedVideoSubInfo struct {
	settings settings.Settings
	log      *logrus.Logger

	embyHelper *embyHelper.EmbyHelper

	taskControl  *task_control.TaskControl
	canceled     bool
	canceledLock sync.Mutex

	subParserHub *sub_parser_hub.SubParserHub

	movieSubMap  map[string]string
	seriesSubMap map[string]string
}

func NewScanPlayedVideoSubInfo(_settings settings.Settings) (*ScanPlayedVideoSubInfo, error) {
	var err error
	var scanPlayedVideoSubInfo ScanPlayedVideoSubInfo
	scanPlayedVideoSubInfo.log = log_helper.GetLogger()
	// 参入设置信息
	scanPlayedVideoSubInfo.settings = _settings
	// 检测是否某些参数超出范围
	scanPlayedVideoSubInfo.settings.Check()
	// 初始化 Emby API 接口
	if scanPlayedVideoSubInfo.settings.EmbySettings.Enable == true && scanPlayedVideoSubInfo.settings.EmbySettings.AddressUrl != "" && scanPlayedVideoSubInfo.settings.EmbySettings.APIKey != "" {
		scanPlayedVideoSubInfo.embyHelper = embyHelper.NewEmbyHelper(*scanPlayedVideoSubInfo.settings.EmbySettings)
	}

	// 初始化任务控制
	scanPlayedVideoSubInfo.taskControl, err = task_control.NewTaskControl(scanPlayedVideoSubInfo.settings.CommonSettings.Threads, log_helper.GetLogger())
	if err != nil {
		return nil, err
	}
	// 字幕解析器
	scanPlayedVideoSubInfo.subParserHub = sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

	return &scanPlayedVideoSubInfo, nil
}

func (s *ScanPlayedVideoSubInfo) Cancel() {
	s.canceledLock.Lock()
	s.canceled = true
	s.canceledLock.Unlock()

	s.taskControl.Release()
}

func (s *ScanPlayedVideoSubInfo) GetPlayedItemsSubtitle() (bool, error) {

	var err error
	// 是否是通过 emby_helper api 获取的列表
	if s.embyHelper == nil {
		// 没有填写 emby_helper api 的信息，那么就跳过
		s.log.Infoln("Skip ScanPlayedVideoSubInfo, Emby Settings is null")
		return false, nil
	}

	s.movieSubMap, s.seriesSubMap, err = s.embyHelper.GetPlayedItemsSubtitle()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *ScanPlayedVideoSubInfo) ScanMovie() error {

	defer func() {
		s.log.Infoln("ScanPlayedVideoSubInfo Movie Sub End")
	}()

	s.log.Infoln("-----------------------------------------------")
	s.log.Infoln("ScanPlayedVideoSubInfo Movie Sub Start...")

	imdbInfoCache := make(map[string]*models.IMDBInfo)
	for movieFPath, subFPath := range s.movieSubMap {

		// 通过视频的绝对路径，从本地的视频文件对应的 nfo 获取到这个视频的 IMDB ID,
		imdbInfo4Movie, err := decode.GetImdbInfo4Movie(movieFPath)
		if err != nil {
			// 如果找不到当前电影的 IMDB Info 本地文件，那么就跳过
			s.log.Warningln("ScanPlayedVideoSubInfo.ScanMovie.GetImdbInfo4Movie", movieFPath, err)
			continue
		}
		// 使用 shooter 的技术 hash 的算法，得到视频的唯一 ID
		fileHash, err := shooter.ComputeFileHash(movieFPath)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.ScanMovie.ComputeFileHash", movieFPath, err)
			continue
		}

		var imdbInfo *models.IMDBInfo
		var ok bool
		if imdbInfo, ok = imdbInfoCache[imdbInfo4Movie.ImdbId]; ok == false {
			// 不存在，那么就去查询和新建缓存
			imdbInfo, err = imdb_helper.GetVideoIMDBInfoFromLocal(imdbInfo4Movie.ImdbId, *s.settings.AdvancedSettings.ProxySettings)
			if err != nil {
				return err
			}
			imdbInfoCache[imdbInfo4Movie.ImdbId] = imdbInfo
		}

		// 查找关联的 VideoSubInfo
		var videoSubInfos []models.VideoSubInfo
		err = dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Find(&videoSubInfos)
		if err != nil {
			return err
		}
		// 判断找到的关联字幕信息是否已经存在了，不存在则新增关联
		var exist bool
		for _, info := range videoSubInfos {

			// 首先，这里会进行已有缓存字幕是否存在的判断，把不存在的字幕给删除了
			if my_util.IsFile(info.StoreFPath) == false {
				// 关联删除了，但是不会删除这些对象，所以后续还需要再次删除
				err := dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Delete(&info)
				if err != nil {
					s.log.Warningln("ScanPlayedVideoSubInfo.ScanMovie.Delete Association", info.Feature, err)
					continue
				}
				// 继续删除这个对象
				dao.GetDb().Delete(&info)
				s.log.Infoln("Delete Not Exist Sub Association", info.SubName, err)
				continue
			}

			if info.Feature == fileHash {
				exist = true
				break
			}
		}
		if exist == true {
			// 存在
			continue
		}
		// 不存在，插入，建立关系
		bok, fileInfo, err := s.subParserHub.DetermineFileTypeFromFile(subFPath)
		if err != nil {
			return err
		}
		if bok == false {
			s.log.Warningln("ScanPlayedVideoSubInfo.ScanMovie.DetermineFileTypeFromFile", imdbInfo4Movie.Title, err)
			continue
		}

		oneVideoSubInfo := models.NewVideoSubInfo(
			fileHash,
			filepath.Base(subFPath),
			"language_iso",
			language.IsBilingualSubtitle(fileInfo.Lang),
			"chinese_iso",
			fileInfo.Lang.String(),
			subFPath,
		)

		err = dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Append(oneVideoSubInfo)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.ScanMovie.Append Association", oneVideoSubInfo.SubName, err)
			continue
		}

	}

	return nil
}

func (s *ScanPlayedVideoSubInfo) ScanSeries() error {

	defer func() {
		s.log.Infoln("ScanPlayedVideoSubInfo Series Sub End")
	}()

	s.log.Infoln("ScanPlayedVideoSubInfo Series Sub Start...")

	for episodeFPath, subFPath := range s.seriesSubMap {

		// 通过视频的绝对路径，，从本地的视频文件对应的 nfo 获取到这个视频的 IMDB ID
		imdbInfo4Series, err := decode.GetSeriesImdbInfoFromEpisode(episodeFPath)
		if err != nil {
			// 如果找不到当前电影的 IMDB Info 本地文件，那么就跳过
			s.log.Warningln("ScanPlayedVideoSubInfo.ScanMovie.GetSeriesImdbInfoFromEpisode", episodeFPath, err)
			continue
		}
		// 使用 shooter 的技术 hash 的算法，得到视频的唯一 ID
		fileHash, err := shooter.ComputeFileHash(episodeFPath)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.ScanMovie.ComputeFileHash", episodeFPath, err)
			continue
		}

		println(subFPath)
		println(imdbInfo4Series)
		println(fileHash)
	}

	return nil
}
