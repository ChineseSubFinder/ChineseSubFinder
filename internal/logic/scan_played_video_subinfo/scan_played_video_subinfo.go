package scan_played_video_subinfo

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/imdb_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_file_hash"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_share_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/task_control"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
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
	// 最大获取的视频数目设置到 100W
	_settings.EmbySettings.MaxRequestVideoNumber = 1000000
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

	// -----------------------------------------------------
	// 并发控制
	s.taskControl.SetCtxProcessFunc("ScanSubPlayedPool", s.scan, common.ScanPlayedSubTimeOut)
	// -----------------------------------------------------

	err := s.taskControl.Invoke(&task_control.TaskData{
		Index: 0,
		Count: len(s.movieSubMap),
		DataEx: ScanInputData{
			Videos:  s.movieSubMap,
			IsMovie: true,
		},
	})
	if err != nil {
		s.log.Errorln("ScanPlayedVideoSubInfo.Movie Sub Error", err)
	}

	err = s.taskControl.Invoke(&task_control.TaskData{
		Index: 0,
		Count: len(s.seriesSubMap),
		DataEx: ScanInputData{
			Videos:  s.seriesSubMap,
			IsMovie: false,
		},
	})
	if err != nil {
		s.log.Errorln("ScanPlayedVideoSubInfo.Series Sub Error", err)
	}

	s.taskControl.Hold()

	return nil
}

func (s *ScanPlayedVideoSubInfo) scan(ctx context.Context, inData interface{}) error {

	taskData := inData.(*task_control.TaskData)
	scanInputData := taskData.DataEx.(ScanInputData)

	videoTypes := ""
	if scanInputData.IsMovie == true {
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

	shareRootDir, err := my_util.GetShareSubRootFolder()
	if err != nil {
		return err
	}

	imdbInfoCache := make(map[string]*models.IMDBInfo)
	index := 0
	for videoFPath, orgSubFPath := range scanInputData.Videos {

		index++
		stage := make(chan interface{}, 1)
		go func() {
			s.dealOneVideo(index, videoFPath, orgSubFPath, videoTypes, shareRootDir, scanInputData.IsMovie, imdbInfoCache)
			stage <- 1
		}()

		select {
		case <-ctx.Done():
			{
				return errors.New(fmt.Sprintf("cancel at scan: %s", videoFPath))
			}
		case <-stage:
			break
		}
	}

	return nil
}

func (s *ScanPlayedVideoSubInfo) dealOneVideo(index int, videoFPath, orgSubFPath, videoTypes, shareRootDir string,
	isMovie bool,
	imdbInfoCache map[string]*models.IMDBInfo) {

	s.log.Infoln(index, orgSubFPath)

	if my_util.IsFile(orgSubFPath) == false {

		log_helper.GetLogger().Errorln("Skip", orgSubFPath, "not exist")
		return
	}

	s.log.Debugln(0)

	// 通过视频的绝对路径，从本地的视频文件对应的 nfo 获取到这个视频的 IMDB ID,
	var err error
	var imdbInfo4Video types.VideoIMDBInfo

	if isMovie == true {
		imdbInfo4Video, err = decode.GetImdbInfo4Movie(videoFPath)
	} else {
		imdbInfo4Video, err = decode.GetSeriesImdbInfoFromEpisode(videoFPath)
	}
	if err != nil {
		// 如果找不到当前电影的 IMDB Info 本地文件，那么就跳过
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".GetImdbInfo", videoFPath, err)
		return
	}

	s.log.Debugln(1)

	// 使用 shooter 的技术 hash 的算法，得到视频的唯一 ID
	fileHash, err := sub_file_hash.Calculate(videoFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".ComputeFileHash", videoFPath, err)
		return
	}

	s.log.Debugln(2)

	var imdbInfo *models.IMDBInfo
	var ok bool
	if imdbInfo, ok = imdbInfoCache[imdbInfo4Video.ImdbId]; ok == false {
		// 不存在，那么就去查询和新建缓存
		imdbInfo, err = imdb_helper.GetVideoIMDBInfoFromLocal(imdbInfo4Video.ImdbId, *s.settings.AdvancedSettings.ProxySettings)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".GetVideoIMDBInfoFromLocal", videoFPath, err)
			return
		}
		imdbInfoCache[imdbInfo4Video.ImdbId] = imdbInfo
	}

	s.log.Debugln(3)

	// 判断找到的关联字幕信息是否已经存在了，不存在则新增关联
	var skip bool
	for tt, cacheInfo := range imdbInfo.VideoSubInfos {

		s.log.Debugln(4, tt)

		skip = false
		// 转绝对路径存储
		// 首先，这里会进行已有缓存字幕是否存在的判断，把不存在的字幕给删除了
		cacheSubFPath := filepath.Join(shareRootDir, cacheInfo.StoreRPath)

		tmpSHA1String, err := my_util.GetFileSHA1String(cacheSubFPath)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, "VideoSubInfos.GetFileSHA1String", videoFPath, err)
			return
		}

		if my_util.IsFile(cacheSubFPath) == false {
			// 如果文件不存在，那么就删除之前的关联
			// 关联删除了，但是不会删除这些对象，所以后续还需要再次删除
			err := dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Delete(&cacheInfo)
			if err != nil {
				s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".Delete Association", cacheInfo.SubName, err)
				continue
			}
			// 继续删除这个对象
			dao.GetDb().Delete(&cacheInfo)
			s.log.Infoln("Delete Not Exist or SHA1 not the same， Sub Association", cacheInfo.SubName)
			continue
		}
		if cacheInfo.SHA1 == tmpSHA1String {
			// 如果 SHA1 一样，那么就说明存在了，这里其实并不是很关心文件名是否是一样的
			skip = true
			break
		}
	}
	if skip == true {
		// 存在，跳过
		return
	}

	s.log.Debugln(5)

	// 新增插入
	// 把现有的字幕 copy 到缓存目录中
	bok, subCacheFPath := sub_share_center.CopySub2Cache(orgSubFPath, imdbInfo.IMDBID, imdbInfo.Year)
	if bok == false {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".CopySub2Cache", orgSubFPath, err)
		return
	}

	s.log.Debugln(6)

	// 不存在，插入，建立关系
	bok, fileInfo, err := s.subParserHub.DetermineFileTypeFromFile(subCacheFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".DetermineFileTypeFromFile", imdbInfo4Video.ImdbId, err)
		return
	}
	if bok == false {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".DetermineFileTypeFromFile == false", imdbInfo4Video.ImdbId)
		return
	}

	s.log.Debugln(7)

	// 特指 emby 字幕的情况
	_, _, _, _, extraSubPreName := s.subFormatter.IsMatchThisFormat(filepath.Base(subCacheFPath))
	// 转相对路径存储
	subRelPath, err := filepath.Rel(shareRootDir, subCacheFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".Rel", imdbInfo4Video.ImdbId, err)
		return
	}

	s.log.Debugln(8)

	// 计算需要插入字幕的 sha1
	saveSHA1String, err := my_util.GetFileSHA1String(subCacheFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, "GetFileSHA1String", videoFPath, err)
		return
	}

	s.log.Debugln(9)

	oneVideoSubInfo := models.NewVideoSubInfo(
		fileHash,
		filepath.Base(subCacheFPath),
		language.MyLang2ISO_639_1_String(fileInfo.Lang),
		language.IsBilingualSubtitle(fileInfo.Lang),
		language.MyLang2ChineseISO(fileInfo.Lang),
		fileInfo.Lang.String(),
		subRelPath,
		extraSubPreName,
		saveSHA1String,
	)

	if isMovie == false {
		// 连续剧的时候，如果可能应该获取是 第几季  第几集
		torrentInfo, _, err := decode.GetVideoInfoFromFileFullPath(subCacheFPath)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".GetVideoInfoFromFileFullPath", imdbInfo4Video.Title, err)
			return
		}
		oneVideoSubInfo.Season = torrentInfo.Season
		oneVideoSubInfo.Episode = torrentInfo.Episode
	}

	s.log.Debugln(10)

	err = dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Append(oneVideoSubInfo)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".Append Association", oneVideoSubInfo.SubName, err)
		return
	}
}

type ScanInputData struct {
	Videos  map[string]string
	IsMovie bool
}
