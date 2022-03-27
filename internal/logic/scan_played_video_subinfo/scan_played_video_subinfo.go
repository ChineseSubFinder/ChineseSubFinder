package scan_played_video_subinfo

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
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
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_share_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/task_control"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
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

	subFormatter ifaces.ISubFormatter
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
	// 字幕命名格式解析器
	scanPlayedVideoSubInfo.subFormatter = emby.NewFormatter()

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

func (s *ScanPlayedVideoSubInfo) Scan() error {

	err := s.scan(s.movieSubMap, true)
	if err != nil {
		return err
	}

	err = s.scan(s.seriesSubMap, false)
	if err != nil {
		return err
	}

	return nil
}

func (s *ScanPlayedVideoSubInfo) scan(videos map[string]string, isMovie bool) error {

	shareRootDir, err := my_util.GetShareSubRootFolder()
	if err != nil {
		return err
	}

	videoTypes := ""
	if isMovie == true {
		videoTypes = "Movie"
	} else {
		videoTypes = "Series"
	}

	defer func() {
		s.log.Infoln("ScanPlayedVideoSubInfo", videoTypes, "Sub End")
		s.log.Infoln("-----------------------------------------------")
	}()

	s.log.Infoln("-----------------------------------------------")
	s.log.Infoln("ScanPlayedVideoSubInfo", videoTypes, "Sub Start...")

	imdbInfoCache := make(map[string]*models.IMDBInfo)
	for movieFPath, orgSubFPath := range videos {

		if my_util.IsFile(orgSubFPath) == false {

			log_helper.GetLogger().Errorln("Skip", orgSubFPath, "not exist")
			continue
		}

		// 通过视频的绝对路径，从本地的视频文件对应的 nfo 获取到这个视频的 IMDB ID,
		var err error
		var imdbInfo4Video types.VideoIMDBInfo

		if isMovie == true {
			imdbInfo4Video, err = decode.GetImdbInfo4Movie(movieFPath)
		} else {
			imdbInfo4Video, err = decode.GetSeriesImdbInfoFromEpisode(movieFPath)
		}
		if err != nil {
			// 如果找不到当前电影的 IMDB Info 本地文件，那么就跳过
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".GetImdbInfo4Movie", movieFPath, err)
			continue
		}
		// 使用 shooter 的技术 hash 的算法，得到视频的唯一 ID
		fileHash, err := shooter.ComputeFileHash(movieFPath)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".ComputeFileHash", movieFPath, err)
			continue
		}

		var imdbInfo *models.IMDBInfo
		var ok bool
		if imdbInfo, ok = imdbInfoCache[imdbInfo4Video.ImdbId]; ok == false {
			// 不存在，那么就去查询和新建缓存
			imdbInfo, err = imdb_helper.GetVideoIMDBInfoFromLocal(imdbInfo4Video.ImdbId, *s.settings.AdvancedSettings.ProxySettings)
			if err != nil {
				return err
			}
			imdbInfoCache[imdbInfo4Video.ImdbId] = imdbInfo
		}
		// 判断找到的关联字幕信息是否已经存在了，不存在则新增关联
		var exist bool
		for _, info := range imdbInfo.VideoSubInfos {

			// 转绝对路径存储
			// 首先，这里会进行已有缓存字幕是否存在的判断，把不存在的字幕给删除了
			if my_util.IsFile(filepath.Join(shareRootDir, info.StoreRPath)) == false {
				// 关联删除了，但是不会删除这些对象，所以后续还需要再次删除
				err := dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Delete(&info)
				if err != nil {
					s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".Delete Association", info.SubName, err)
					continue
				}
				// 继续删除这个对象
				dao.GetDb().Delete(&info)
				s.log.Infoln("Delete Not Exist Sub Association", info.SubName, err)
				continue
			}
			// 文件对应的视频唯一 ID 一致
			if info.Feature == fileHash {
				exist = true
				break
			}
		}
		if exist == true {
			// 存在
			continue
		}

		// 把现有的字幕 copy 到缓存目录中
		bok, subCacheFPath := sub_share_center.CopySub2Cache(orgSubFPath, imdbInfo.Year)
		if bok == false {
			continue
		}

		// 不存在，插入，建立关系
		bok, fileInfo, err := s.subParserHub.DetermineFileTypeFromFile(subCacheFPath)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".DetermineFileTypeFromFile", imdbInfo4Video.ImdbId, err)
			continue
		}
		if bok == false {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".DetermineFileTypeFromFile == false", imdbInfo4Video.ImdbId)
			continue
		}

		// 特指 emby 字幕的情况
		bok, _, _, _, extraSubPreName := s.subFormatter.IsMatchThisFormat(filepath.Base(subCacheFPath))
		if bok == false {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".IsMatchThisFormat == false", imdbInfo4Video.ImdbId)
			continue
		}
		// 转相对路径存储
		subRelPath, err := filepath.Rel(shareRootDir, subCacheFPath)
		if err != nil {
			return err
		}

		oneVideoSubInfo := models.NewVideoSubInfo(
			fileHash,
			filepath.Base(subCacheFPath),
			language.MyLang2ISO_639_1_String(fileInfo.Lang),
			language.IsBilingualSubtitle(fileInfo.Lang),
			language.MyLang2ChineseISO(fileInfo.Lang),
			fileInfo.Lang.String(),
			subRelPath,
			extraSubPreName,
		)

		if isMovie == false {
			// 连续剧的时候，如果可能应该获取是 第几季  第几集
			torrentInfo, _, err := decode.GetVideoInfoFromFileFullPath(subCacheFPath)
			if err != nil {
				s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".GetVideoInfoFromFileFullPath", imdbInfo4Video.Title, err)
				continue
			}
			oneVideoSubInfo.Season = torrentInfo.Season
			oneVideoSubInfo.Episode = torrentInfo.Episode
		}

		err = dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Append(oneVideoSubInfo)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".Append Association", oneVideoSubInfo.SubName, err)
			continue
		}

	}

	return nil
}
