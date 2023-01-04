package video_scan_and_refresh_helper

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/scan_logic"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/search"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/path_helper"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ifaces"
	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/emby"
	task_queue2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"

	embyHelper "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/emby_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/forced_scan_and_down_sub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/movie_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/restore_fix_timeline_bk"
	seriesHelper "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/series_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	subSupplier "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/xunlei"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/imdb_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/mix_media_info"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sort_things"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_file_hash"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_share_center"
	subTimelineFixerPKG "github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_timeline_fixer"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/task_control"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/task_queue"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/jinzhu/now"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type VideoScanAndRefreshHelper struct {
	log                      *logrus.Logger                  // 日志实例
	fileDownloader           *file_downloader.FileDownloader // 文件下载器
	NeedForcedScanAndDownSub bool                            // 将会强制扫描所有的视频，下载字幕，替换已经存在的字幕，不进行时间段和已存在则跳过的判断。且不会进过 Emby API 的逻辑，智能进行强制去以本程序的方式去扫描。
	NeedRestoreFixTimeLineBK bool                            // 从 csf-bk 文件还原时间轴修复前的字幕文件
	embyHelper               *embyHelper.EmbyHelper          // Emby 的实例
	downloadQueue            *task_queue.TaskQueue           // 需要下载的视频的队列
	subSupplierHub           *subSupplier.SubSupplierHub     // 字幕提供源的集合，仅仅是 check 是否需要下载字幕是足够的，如果要下载则需要额外的初始化和检查
	taskControl              *task_control.TaskControl       // 任务控制器
	running                  bool                            // 是否正在运行
	locker                   sync.Mutex                      // 互斥锁
	SubParserHub             *sub_parser_hub.SubParserHub    // 字幕解析器
	subFormatter             ifaces.ISubFormatter            // 字幕格式化器

	processLocker sync.Mutex
}

func NewVideoScanAndRefreshHelper(inSubFormatter ifaces.ISubFormatter, fileDownloader *file_downloader.FileDownloader, downloadQueue *task_queue.TaskQueue) *VideoScanAndRefreshHelper {
	v := VideoScanAndRefreshHelper{log: fileDownloader.Log, downloadQueue: downloadQueue,
		subSupplierHub: subSupplier.NewSubSupplierHub(
			xunlei.NewSupplier(fileDownloader),
		),
		fileDownloader: fileDownloader,
		// 字幕解析器
		SubParserHub: sub_parser_hub.NewSubParserHub(fileDownloader.Log, ass.NewParser(fileDownloader.Log), srt.NewParser(fileDownloader.Log)),
		subFormatter: inSubFormatter,
	}

	var err error
	v.taskControl, err = task_control.NewTaskControl(settings.Get().CommonSettings.Threads, v.log)
	if err != nil {
		fileDownloader.Log.Panicln(err)
	}
	return &v
}

func (v *VideoScanAndRefreshHelper) Start(scanLogic *scan_logic.ScanLogic) error {

	v.locker.Lock()
	if v.running == true {
		v.locker.Unlock()
		v.log.Infoln("VideoScanAndRefreshHelper is already running")
		return nil
	}
	v.running = true
	v.locker.Unlock()

	startT := time.Now()
	defer func() {

		v.locker.Lock()
		v.running = false
		v.locker.Unlock()

		v.log.Infoln("VideoScanAndRefreshHelper finished, cost:", time.Since(startT).Minutes(), "min")
		v.log.Infoln("Video Scan End")
		v.log.Infoln("------------------------------------")
	}()

	v.log.Infoln("------------------------------------")
	v.log.Infoln("Video Scan Started...")
	// 先进行扫描
	scanResult, err := v.ScanNormalMovieAndSeries()
	if err != nil {
		v.log.Errorln("ScanNormalMovieAndSeries", err)
		return err
	}
	err = v.ScanEmbyMovieAndSeries(scanResult)
	if err != nil {
		v.log.Errorln("ScanEmbyMovieAndSeries", err)
		return err
	}
	// 过滤出需要下载的视频有那些，并放入队列中
	err = v.FilterMovieAndSeriesNeedDownload(scanResult, scanLogic)
	if err != nil {
		v.log.Errorln("FilterMovieAndSeriesNeedDownload", err)
		return err
	}
	if settings.Get().ExperimentalFunction.ShareSubSettings.ShareSubEnabled == true {
		v.log.Infoln("ShareSubEnabled is true, will scan share sub")
		// 根据上面得到的 scanResult 的 Normal 部分进行字幕的扫描，也存入到 VideoSubInfo 中，但是需要标记这个是低可信度的
		//v.scanLowVideoSubInfo(scanResult)
	}

	return nil
}

func (v *VideoScanAndRefreshHelper) Cancel() {

	v.locker.Lock()
	if v.running == false {
		v.locker.Unlock()
		v.log.Infoln("VideoScanAndRefreshHelper is not running")
		return
	}
	v.locker.Unlock()

	defer func() {
		v.log.Infoln("VideoScanAndRefreshHelper.Cancel()")
	}()

	v.taskControl.Release()
}

// ReadSpeFile 优先级最高。读取特殊文件，启用一些特殊的功能，比如 forced_scan_and_down_sub
func (v *VideoScanAndRefreshHelper) ReadSpeFile() error {
	// 理论上是一次性的，用了这个文件就应该没了
	// 强制的字幕扫描
	needProcessForcedScanAndDownSub, err := forced_scan_and_down_sub.CheckSpeFile()
	if err != nil {
		return err
	}
	v.NeedForcedScanAndDownSub = needProcessForcedScanAndDownSub
	// 从 csf-bk 文件还原时间轴修复前的字幕文件
	needProcessRestoreFixTimelineBK, err := restore_fix_timeline_bk.CheckSpeFile()
	if err != nil {
		return err
	}
	v.NeedRestoreFixTimeLineBK = needProcessRestoreFixTimelineBK

	v.log.Infoln("NeedRestoreFixTimeLineBK ==", needProcessRestoreFixTimelineBK)

	return nil
}

// ScanNormalMovieAndSeries 没有媒体服务器，扫描出有那些电影、连续剧需要进行字幕下载的
func (v *VideoScanAndRefreshHelper) ScanNormalMovieAndSeries() (*ScanVideoResult, error) {

	defer func() {
		v.log.Infoln("ScanNormalMovieAndSeries End")
	}()
	v.log.Infoln("ScanNormalMovieAndSeries Start...")

	var err error
	outScanVideoResult := ScanVideoResult{}
	// ------------------------------------------------------------------------------
	// 由于需要进行视频信息的缓存，用于后续的逻辑，那么本地视频的扫描默认都会进行
	normalScanResult := NormalScanVideoResult{}
	// 直接由本程序自己去扫描视频视频有哪些
	// 全扫描
	if v.NeedForcedScanAndDownSub == true {
		v.log.Infoln("Forced Scan And DownSub")
	}
	wg := sync.WaitGroup{}
	var errMovie, errSeries error
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		// --------------------------------------------------
		// 电影
		// 没有填写 emby_helper api 的信息，那么就走常规的全文件扫描流程
		normalScanResult.MoviesDirMap, errMovie = search.MatchedVideoFileFromDirs(v.log, settings.Get().CommonSettings.MoviePaths)
	}()
	wg.Add(1)
	go func() {

		defer func() {
			wg.Done()
		}()
		// --------------------------------------------------
		// 连续剧
		// 遍历连续剧总目录下的第一层目录
		normalScanResult.SeriesDirMap, errSeries = seriesHelper.GetSeriesListFromDirs(v.log, settings.Get().CommonSettings.SeriesPaths)
		// ------------------------------------------------------------------------------
		// 输出调试信息，有那些连续剧文件夹名称
		if normalScanResult.SeriesDirMap == nil {
			return
		}
		normalScanResult.SeriesDirMap.Each(func(key interface{}, value interface{}) {
			for i, s := range value.([]string) {
				v.log.Debugln("embyHelper == nil GetSeriesList", i, s)
			}
		})
	}()
	wg.Wait()
	if errMovie != nil {
		return nil, errMovie
	}
	if errSeries != nil {
		return nil, errSeries
	}
	// ------------------------------------------------------------------------------
	outScanVideoResult.Normal = &normalScanResult
	// ------------------------------------------------------------------------------
	// 将扫描到的信息缓存到本地中，用于后续的 Video 展示界面 和 Emby IMDB ID 匹配进行路径的转换
	err = v.updateLocalVideoCacheInfo(&outScanVideoResult)
	if err != nil {
		return nil, err
	}

	return &outScanVideoResult, nil
}

// ScanEmbyMovieAndSeries Emby媒体服务器，扫描出有那些电影、连续剧需要进行字幕下载的
func (v *VideoScanAndRefreshHelper) ScanEmbyMovieAndSeries(scanVideoResult *ScanVideoResult) error {

	defer func() {
		v.log.Infoln("ScanEmbyMovieAndSeries End")
	}()
	v.log.Infoln("ScanEmbyMovieAndSeries Start...")

	if settings.Get().EmbySettings.Enable == false {
		v.embyHelper = nil
		v.log.Infoln("EmbyHelper == nil")
	} else {

		if v.NeedForcedScanAndDownSub == true {

			v.log.Infoln("Forced Scan And DownSub, tmpSetting.EmbySettings.MaxRequestVideoNumber = 1000000")
			// 如果是强制，那么就临时修改 Setting 的 Emby MaxRequestVideoNumber 参数为 1000000
			v.embyHelper = embyHelper.NewEmbyHelper(v.fileDownloader.MediaInfoDealers)
		} else {
			v.log.Infoln("Not Forced Scan And DownSub")
			v.embyHelper = embyHelper.NewEmbyHelper(v.fileDownloader.MediaInfoDealers)
		}
	}
	var err error

	// ------------------------------------------------------------------------------
	// 从 Emby 获取视频
	if v.embyHelper != nil {
		// TODO 如果后续支持了 Jellyfin、Plex 那么这里需要额外正在对应的扫描逻辑
		// 进过 emby_helper api 的信息读取
		embyScanResult := EmbyScanVideoResult{}
		v.log.Infoln("Movie Sub Dl From Emby API...")
		// Emby 情况，从 Emby 获取视频信息
		err = v.refreshEmbySubList(settings.Get().EmbySettings, false, common2.EmbyApiGetItemsLimitMax)
		if err != nil {
			v.log.Errorln("refreshEmbySubList", err)
			return err
		}
		// ------------------------------------------------------------------------------
		// 有哪些更新的视频列表，包含电影、连续剧
		embyScanResult.MovieSubNeedDlEmbyMixInfoList, embyScanResult.SeriesSubNeedDlEmbyMixInfoMap, err = v.getUpdateVideoListFromEmby()
		if err != nil {
			v.log.Errorln("getUpdateVideoListFromEmby", err)
			return err
		}
		// ------------------------------------------------------------------------------
		scanVideoResult.Emby = &embyScanResult
	}

	return nil
}

// FilterMovieAndSeriesNeedDownload 过滤出需要下载字幕的视频，比如是否跳过中文的剧集，是否超过3个月的下载时间，丢入队列中
func (v *VideoScanAndRefreshHelper) FilterMovieAndSeriesNeedDownload(scanVideoResult *ScanVideoResult, scanLogic *scan_logic.ScanLogic) error {

	if scanVideoResult.Normal != nil && settings.Get().EmbySettings.Enable == false {
		err := v.filterMovieAndSeriesNeedDownloadNormal(scanVideoResult.Normal, scanLogic)
		if err != nil {
			return err
		}
	}

	if scanVideoResult.Emby != nil && settings.Get().EmbySettings.Enable == true {

		// 先获取缓存的 Emby 视频信息，有那些已经在这次扫描的时候播放过了

		// 然后才是过滤有哪些需要下载的
		err := v.filterMovieAndSeriesNeedDownloadEmby(scanVideoResult.Emby, scanLogic)
		if err != nil {
			return err
		}
	}

	return nil
}

// RefreshMediaServerSubList 刷新媒体服务器的字幕列表
func (v *VideoScanAndRefreshHelper) RefreshMediaServerSubList() error {

	if settings.Get().EmbySettings.Enable == false {
		return nil
	}
	v.log.Infoln("Refresh Media Server Sub List...")
	defer func() {
		v.log.Infoln("Refresh Media Server Sub List End")
	}()

	v.log.Infoln("tmpSetting.EmbySettings.MaxRequestVideoNumber = 1000000")
	// 如果是强制，那么就临时修改 Setting 的 Emby MaxRequestVideoNumber 参数为 1000000
	v.embyHelper = embyHelper.NewEmbyHelper(v.fileDownloader.MediaInfoDealers)
	var err error
	if v.embyHelper != nil {
		// TODO 如果后续支持了 Jellyfin、Plex 那么这里需要额外正在对应的扫描逻辑
		// Emby 情况，从 Emby 获取视频信息
		err = v.refreshEmbySubList(settings.Get().EmbySettings, false, common2.EmbyApiGetItemsLimitMax)
		if err != nil {
			v.log.Errorln("refreshEmbySubList", err)
			return err
		}
	}

	return nil
}

// scanLowVideoSubInfo 扫描低可信度的字幕信息
func (v *VideoScanAndRefreshHelper) scanLowVideoSubInfo(scanVideoResult *ScanVideoResult) {

	// 需要根据搜索到的字幕或者视频信息与 VideoSubInfo 的信息进行交叉
	if scanVideoResult.Normal == nil {
		return
	}

	shareRootDir, err := pkg.GetShareSubRootFolder()
	if err != nil {
		v.log.Errorln("scanLowVideoSubInfo.GetShareSubRootFolder", err)
		return
	}

	// 先处理电影
	scanVideoResult.Normal.MoviesDirMap.Any(func(movieDirRootPath interface{}, movieFPath interface{}) bool {

		videoFPathList := movieFPath.([]string)
		for videoIndex, videoFPath := range videoFPathList {

			v.log.Infoln("--------------------------------------------------------------------------------")
			v.log.Infoln("scanLowVideoSubInfo.MovieHasChineseSub", videoIndex, videoFPath)
			mixMediaInfo, err := mix_media_info.GetMixMediaInfo(v.fileDownloader.MediaInfoDealers, videoFPath, true)
			if err != nil {
				v.log.Warningln("scanLowVideoSubInfo.GetMixMediaInfo", videoFPath, err)
				continue
			}
			// 这个视频有对应的文中字幕
			bFoundChineseSub, _, chineseSubFitVideoNameFullPathList, err := movie_helper.MovieHasChineseSub(v.log, videoFPath)
			if err != nil {
				v.log.Warningln("scanLowVideoSubInfo.MovieHasChineseSub", videoFPath, err)
				continue
			}
			if bFoundChineseSub == false {
				// 没有找到中文字幕，那么就不需要下载了
				v.log.Infoln("scanLowVideoSubInfo.MovieHasChineseSub", videoFPath, "not found chinese sub")
				continue
			}

			// 使用本程序的 hash 的算法，得到视频的唯一 ID
			fileHash, err := sub_file_hash.Calculate(videoFPath)
			if err != nil {
				v.log.Warningln("scanLowVideoSubInfo.ComputeFileHash", videoFPath, err)
				continue
			}
			// 得到的这些字幕需要进行一次 sha256 的排除，因为是多个站点下载的，可能是重复的
			subSha256Map := make(map[string]string)
			for _, orgSubFPath := range chineseSubFitVideoNameFullPathList {
				// 计算需要插入字幕的 sha256
				saveSHA256String, err := pkg.GetFileSHA256String(orgSubFPath)
				if err != nil {
					v.log.Warningln("scanLowVideoSubInfo.GetFileSHA256String", orgSubFPath, err)
					continue
				}
				subSha256Map[saveSHA256String] = orgSubFPath
			}
			// 排除重复 sha256 后的字幕
			for _, orgSubFPath := range subSha256Map {
				v.log.Infoln(orgSubFPath)
				// 需要得到这个视频对应的字幕的绝对地址
				v.addLowVideoSubInfo(true, 0, 0, orgSubFPath, mixMediaInfo, shareRootDir, fileHash)
			}
		}

		return false
	})
	// 处理连续剧  media root dir -- series dir
	scanVideoResult.Normal.SeriesDirMap.Any(func(seriesDirRootPath interface{}, seriesFPath interface{}) bool {

		seriesDirRootFPathLisst := seriesFPath.([]string)
		for seriesDirIndex, seriesDirRootFPath := range seriesDirRootFPathLisst {

			seriesInfo, err := seriesHelper.ReadSeriesInfoFromDir(v.fileDownloader.MediaInfoDealers, seriesDirRootFPath, 90, true, true)
			if err != nil {
				v.log.Warningln("scanLowVideoSubInfo.ReadSeriesInfoFromDir", seriesDirRootFPath, err)
				return false
			}

			if len(seriesInfo.EpList) < 1 {
				continue
			}

			mixMediaInfo, err := mix_media_info.GetMixMediaInfo(v.fileDownloader.MediaInfoDealers, seriesInfo.EpList[0].FileFullPath, false)
			if err != nil {
				v.log.Warningln("scanLowVideoSubInfo.GetMixMediaInfo", seriesInfo.EpList[0].FileFullPath, err)
				continue
			}

			for i, episodeInfo := range seriesInfo.EpList {

				videoFPath := episodeInfo.FileFullPath
				v.log.Infoln("--------------------------------------------------------------------------------")
				v.log.Infoln("scanLowVideoSubInfo.ReadSeriesInfoFromDir", seriesDirIndex, i, videoFPath)
				// 使用本程序的 hash 的算法，得到视频的唯一 ID
				fileHash, err := sub_file_hash.Calculate(videoFPath)
				if err != nil {
					v.log.Warningln("scanLowVideoSubInfo.ComputeFileHash", videoFPath, err)
					continue
				}
				// 得到的这些字幕需要进行一次 sha256 的排除，因为是多个站点下载的，可能是重复的
				subSha256Map := make(map[string]string)
				// 这个视频有对应的文中字幕
				for _, subInfo := range episodeInfo.SubAlreadyDownloadedList {

					orgSubFPath := subInfo.FileFullPath
					if language.HasChineseLang(subInfo.Language) == false {
						v.log.Warningln("scanLowVideoSubInfo.HasChineseLang Skip", videoFPath, subInfo.Language)
						continue
					}
					// 计算需要插入字幕的 sha256
					saveSHA256String, err := pkg.GetFileSHA256String(orgSubFPath)
					if err != nil {
						v.log.Warningln("scanLowVideoSubInfo.GetFileSHA256String", orgSubFPath, err)
						continue
					}
					subSha256Map[saveSHA256String] = orgSubFPath
				}
				// 排除重复 sha256 后的字幕
				for _, orgSubFPath := range subSha256Map {
					v.log.Infoln(orgSubFPath)
					// 需要得到这个视频对应的字幕的绝对地址
					v.addLowVideoSubInfo(false, episodeInfo.Season, episodeInfo.Episode, orgSubFPath, mixMediaInfo, shareRootDir, fileHash)
				}
			}
		}

		return false
	})
}

// 从绝对字幕路径和 mixMediaInfo 信息判断是否需要存储这个低可信度的字幕
func (v *VideoScanAndRefreshHelper) addLowVideoSubInfo(isMovie bool, Season, Eps int, orgSubFPath string, mixMediaInfo *models.MediaInfo, shareRootDir string, fileHash string) {

	// 计算需要插入字幕的 sha256
	saveSHA256String, err := pkg.GetFileSHA256String(orgSubFPath)
	if err != nil {
		v.log.Warningln("scanLowVideoSubInfo.GetFileSHA256String", orgSubFPath, err)
		return
	}
	// 这个字幕文件是否已经存在了 LowVideoSubInfo
	var lowVideoSubInfos []models.LowVideoSubInfo
	dao.GetDb().Where("sha256 = ?", saveSHA256String).Find(&lowVideoSubInfos)
	if len(lowVideoSubInfos) > 0 {
		// 存在，跳过
		v.log.Infoln("scanLowVideoSubInfo.SHA256 LowVideoSubInfo Exist == true, Skip", orgSubFPath)
		return
	}
	// 这个字幕文件是否已经存在了 LowVideoSubInfo
	var videoSubInfos []models.VideoSubInfo
	dao.GetDb().Where("sha256 = ?", saveSHA256String).Find(&videoSubInfos)
	if len(videoSubInfos) > 0 {
		// 存在，跳过
		v.log.Infoln("scanLowVideoSubInfo.SHA256 VideoSubInfo Exist == true, Skip", orgSubFPath)
		return
	}

	parseTime, err := now.Parse(mixMediaInfo.Year)
	if err != nil {
		v.log.Warningln("ParseTime", mixMediaInfo.Year, err)
		return
	}
	// 把现有的字幕 copy 到缓存目录中
	bok, subCacheFPath := sub_share_center.CopySub2Cache(v.log, orgSubFPath, mixMediaInfo.ImdbId, parseTime.Year(), true)
	if bok == false {
		v.log.Warningln("scanLowVideoSubInfo.CopySub2Cache", orgSubFPath, err)
		return
	}
	// 不存在，插入，建立关系
	bok, fileInfo, err := v.SubParserHub.DetermineFileTypeFromFile(subCacheFPath)
	if err != nil {
		v.log.Warningln("scanLowVideoSubInfo.DetermineFileTypeFromFile", mixMediaInfo.ImdbId, err)
		return
	}
	if bok == false {
		v.log.Warningln("scanLowVideoSubInfo.DetermineFileTypeFromFile == false", mixMediaInfo.ImdbId)
		return
	}
	// 转相对路径存储
	subRelPath, err := filepath.Rel(shareRootDir, subCacheFPath)
	if err != nil {
		v.log.Warningln("scanLowVideoSubInfo.Rel", mixMediaInfo.ImdbId, err)
		return
	}
	// 字幕的情况
	_, _, _, _, extraSubPreName := v.subFormatter.IsMatchThisFormat(filepath.Base(subCacheFPath))

	oneLowVideoSubInfo := models.NewLowVideoSubInfo(
		mixMediaInfo.ImdbId,
		mixMediaInfo.TmdbId,
		fileHash,
		filepath.Base(subCacheFPath),
		language.MyLang2ISO_639_1_String(fileInfo.Lang),
		language.IsBilingualSubtitle(fileInfo.Lang),
		language.MyLang2ChineseISO(fileInfo.Lang),
		fileInfo.Lang.String(),
		subRelPath,
		extraSubPreName,
		saveSHA256String,
		isMovie,
	)

	oneLowVideoSubInfo.Season = Season
	oneLowVideoSubInfo.Episode = Eps

	dao.GetDb().Save(oneLowVideoSubInfo)
	return
}

func (v *VideoScanAndRefreshHelper) ScrabbleUpVideoList(scanVideoResult *ScanVideoResult, pathUrlMap map[string]string) ([]backend2.MovieInfo, []backend2.SeasonInfo) {

	defer func() {
		scanVideoResult = nil
	}()

	if scanVideoResult.Normal != nil && settings.Get().EmbySettings.Enable == false {
		return v.scrabbleUpVideoListNormal(scanVideoResult.Normal, pathUrlMap)
	}

	if scanVideoResult.Emby != nil && settings.Get().EmbySettings.Enable == true {
		return v.scrabbleUpVideoListEmby(scanVideoResult.Emby, pathUrlMap)
	}

	return nil, nil
}

func (v *VideoScanAndRefreshHelper) scrabbleUpVideoListNormal(normal *NormalScanVideoResult, pathUrlMap map[string]string) ([]backend2.MovieInfo, []backend2.SeasonInfo) {

	movieInfos := make([]backend2.MovieInfo, 0)
	seasonInfos := make([]backend2.SeasonInfo, 0)

	if normal == nil {
		return movieInfos, seasonInfos
	}
	// 电影
	movieProcess := func(ctx context.Context, inData interface{}) error {

		taskData := inData.(*task_control.TaskData)
		scrabbleUpVideoMovieNormalInput := taskData.DataEx.(ScrabbleUpVideoMovieNormalInput)
		oneMovieDirRootPath := scrabbleUpVideoMovieNormalInput.OneMovieDirRootPath
		oneMovieFPath := scrabbleUpVideoMovieNormalInput.OneMovieFPath

		v.processLocker.Lock()
		desUrl, found := pathUrlMap[oneMovieDirRootPath]
		if found == false {
			v.processLocker.Unlock()
			// 没有找到对应的 URL
			v.log.Warningln("scrabbleUpVideoListNormal.movieProcess.processLocker.Unlock", oneMovieDirRootPath)
			return nil
		}
		v.processLocker.Unlock()

		// 匹配上了前缀就替换这个，并记录
		movieFUrl := path_helper.ChangePhysicalPathToSharePath(oneMovieFPath, oneMovieDirRootPath, desUrl)
		//movieFUrl := strings.ReplaceAll(oneMovieFPath, oneMovieDirRootPath, desUrl)
		oneMovieInfo := backend2.MovieInfo{
			Name:         filepath.Base(movieFUrl),
			DirRootUrl:   filepath.Dir(movieFUrl),
			VideoFPath:   oneMovieFPath,
			VideoUrl:     movieFUrl,
			SubFPathList: make([]string, 0),
		}
		// 搜索字幕
		matchedSubFileByOneVideo, err := sub_helper.SearchMatchedSubFileByOneVideo(v.log, oneMovieFPath)
		if err != nil {
			v.log.Errorln("SearchMatchedSubFileByOneVideo", err)
		}
		matchedSubFileByOneVideoUrl := make([]string, 0)
		for _, oneSubFPath := range matchedSubFileByOneVideo {
			oneSubFUrl := path_helper.ChangePhysicalPathToSharePath(oneSubFPath, oneMovieDirRootPath, desUrl)
			//oneSubFUrl := strings.ReplaceAll(oneSubFPath, oneMovieDirRootPath, desUrl)
			matchedSubFileByOneVideoUrl = append(matchedSubFileByOneVideoUrl, oneSubFUrl)
		}
		oneMovieInfo.SubFPathList = append(oneMovieInfo.SubFPathList, matchedSubFileByOneVideoUrl...)

		v.processLocker.Lock()
		movieInfos = append(movieInfos, oneMovieInfo)
		v.processLocker.Unlock()

		return nil
	}
	// ----------------------------------------
	v.taskControl.SetCtxProcessFunc("updateLocalVideoCacheInfo", movieProcess, common2.ScanPlayedSubTimeOut)
	// ----------------------------------------
	normal.MoviesDirMap.Any(func(movieDirRootPath interface{}, moviesFPath interface{}) bool {

		oneMovieDirRootPath := movieDirRootPath.(string)
		for i, oneMovieFPath := range moviesFPath.([]string) {

			v.log.Debugln("scrabbleUpVideoListNormal.movieDirMap.Any", oneMovieDirRootPath, oneMovieFPath)
			// 放入队列
			err := v.taskControl.Invoke(&task_control.TaskData{
				Index: i,
				Count: len(moviesFPath.([]string)),
				DataEx: ScrabbleUpVideoMovieNormalInput{
					OneMovieDirRootPath: oneMovieDirRootPath,
					OneMovieFPath:       oneMovieFPath,
				},
			})
			if err != nil {
				v.log.Errorln("scrabbleUpVideoListNormal.normal.MoviesDirMap.taskControl.Invoke", err)
				return true
			}
		}

		return false
	})
	v.taskControl.Hold()
	// ----------------------------------------
	// 连续剧
	// seriesDirMap: dir <--> seriesList
	seriesProcess := func(ctx context.Context, inData interface{}) error {

		taskData := inData.(*task_control.TaskData)
		scrabbleUpVideoSeriesNormalInput := taskData.DataEx.(ScrabbleUpVideoSeriesNormalInput)
		oneSeriesRootPathName := scrabbleUpVideoSeriesNormalInput.OneSeriesRootPathName
		oneSeriesRootDir := scrabbleUpVideoSeriesNormalInput.OneSeriesRootDir

		v.processLocker.Lock()
		desUrl, found := pathUrlMap[oneSeriesRootPathName]
		if found == false {
			v.processLocker.Unlock()
			// 没有找到对应的 URL
			v.log.Warningln("scrabbleUpVideoListNormal.seriesProcess.processLocker.Unlock", oneSeriesRootPathName)
			return nil
		}
		v.processLocker.Unlock()

		bNeedDlSub, seriesInfo, err := v.subSupplierHub.SeriesNeedDlSub(
			v.fileDownloader.MediaInfoDealers,
			oneSeriesRootDir,
			v.NeedForcedScanAndDownSub, false)
		if err != nil {
			v.log.Errorln("filterMovieAndSeriesNeedDownloadNormal.SeriesNeedDlSub", err)
			return err
		}
		if bNeedDlSub == false {
			v.log.Debugln("filterMovieAndSeriesNeedDownloadNormal.SeriesNeedDlSub bNeedDlSub == false", oneSeriesRootPathName)
			return nil
		}
		seriesDirRootFUrl := path_helper.ChangePhysicalPathToSharePath(oneSeriesRootDir, oneSeriesRootPathName, desUrl)
		//seriesDirRootFUrl := strings.ReplaceAll(oneSeriesRootDir, oneSeriesRootPathName, desUrl)
		oneSeasonInfo := backend2.SeasonInfo{
			Name:          filepath.Base(oneSeriesRootDir),
			RootDirPath:   oneSeriesRootDir,
			DirRootUrl:    seriesDirRootFUrl,
			OneVideoInfos: make([]backend2.OneVideoInfo, 0),
		}
		for _, epsInfo := range seriesInfo.EpList {

			videoFUrl := path_helper.ChangePhysicalPathToSharePath(epsInfo.FileFullPath, oneSeriesRootPathName, desUrl)
			//videoFUrl := strings.ReplaceAll(epsInfo.FileFullPath, oneSeriesRootPathName, desUrl)
			oneVideoInfo := backend2.OneVideoInfo{
				Name:         epsInfo.Title,
				VideoFPath:   epsInfo.FileFullPath,
				VideoUrl:     videoFUrl,
				Season:       epsInfo.Season,
				Episode:      epsInfo.Episode,
				SubFPathList: make([]string, 0),
			}

			// 搜索字幕
			matchedSubFileByOneVideo, err := sub_helper.SearchMatchedSubFileByOneVideo(v.log, epsInfo.FileFullPath)
			if err != nil {
				v.log.Errorln("SearchMatchedSubFileByOneVideo", err)
			}
			matchedSubFileByOneVideoUrl := make([]string, 0)
			for _, oneSubFPath := range matchedSubFileByOneVideo {
				oneSubFUrl := path_helper.ChangePhysicalPathToSharePath(oneSubFPath, oneSeriesRootPathName, desUrl)
				//oneSubFUrl := strings.ReplaceAll(oneSubFPath, oneSeriesRootPathName, desUrl)
				matchedSubFileByOneVideoUrl = append(matchedSubFileByOneVideoUrl, oneSubFUrl)
			}
			oneVideoInfo.SubFPathList = append(oneVideoInfo.SubFPathList, matchedSubFileByOneVideoUrl...)
			oneSeasonInfo.OneVideoInfos = append(oneSeasonInfo.OneVideoInfos, oneVideoInfo)
		}

		v.processLocker.Lock()
		seasonInfos = append(seasonInfos, oneSeasonInfo)
		v.processLocker.Unlock()

		return nil
	}
	// ----------------------------------------
	v.taskControl.SetCtxProcessFunc("updateLocalVideoCacheInfo", seriesProcess, common2.ScanPlayedSubTimeOut)
	// ----------------------------------------
	normal.SeriesDirMap.Any(func(seriesRootPathName interface{}, seriesNames interface{}) bool {

		oneSeriesRootPathName := seriesRootPathName.(string)
		for i, oneSeriesRootDir := range seriesNames.([]string) {

			v.log.Debugln("scrabbleUpVideoListNormal.seriesDirMap.Any", oneSeriesRootPathName, oneSeriesRootDir)
			// 放入队列
			err := v.taskControl.Invoke(&task_control.TaskData{
				Index: i,
				Count: len(seriesNames.([]string)),
				DataEx: ScrabbleUpVideoSeriesNormalInput{
					OneSeriesRootDir:      oneSeriesRootDir,
					OneSeriesRootPathName: oneSeriesRootPathName,
				},
			})
			if err != nil {
				v.log.Errorln("scrabbleUpVideoListNormal.normal.SeriesDirMap.taskControl.Invoke", err)
				return true
			}
		}

		return false
	})
	v.taskControl.Hold()
	// ----------------------------------------

	return movieInfos, seasonInfos
}

func (v *VideoScanAndRefreshHelper) scrabbleUpVideoListEmby(emby *EmbyScanVideoResult, pathUrlMap map[string]string) ([]backend2.MovieInfo, []backend2.SeasonInfo) {

	movieInfos := make([]backend2.MovieInfo, 0)
	seasonInfos := make([]backend2.SeasonInfo, 0)

	if emby == nil {
		return movieInfos, seasonInfos
	}
	// 排序得到匹配上的路径，最长的那个
	sortMoviePaths := sort_things.SortStringSliceByLength(settings.Get().CommonSettings.MoviePaths)
	sortSeriesPaths := sort_things.SortStringSliceByLength(settings.Get().CommonSettings.SeriesPaths)
	// ----------------------------------------
	// Emby 过滤，电影

	movieProcess := func(ctx context.Context, inData interface{}) error {

		taskData := inData.(*task_control.TaskData)
		scrabbleUpVideoMovieEmbyInput := taskData.DataEx.(ScrabbleUpVideoMovieEmbyInput)
		oneMovieMixInfo := scrabbleUpVideoMovieEmbyInput.OneMovieMixInfo
		// 首先需要找到对应的最长的视频媒体库路径，x://ABC  x://ABC/DEF
		for _, oneMovieDirPath := range sortMoviePaths {

			if strings.HasPrefix(oneMovieMixInfo.PhysicalVideoFileFullPath, oneMovieDirPath.Path) {
				// 匹配上了
				v.processLocker.Lock()
				desUrl, found := pathUrlMap[oneMovieDirPath.Path]
				if found == false {
					v.processLocker.Unlock()
					// 没有找到对应的 URL
					v.log.Warningln("scrabbleUpVideoListEmby.movieProcess.pathUrlMap", oneMovieDirPath.Path)
					return nil
				}
				v.processLocker.Unlock()
				// 匹配上了前缀就替换这个，并记录
				movieFUrl := path_helper.ChangePhysicalPathToSharePath(oneMovieMixInfo.PhysicalVideoFileFullPath, oneMovieDirPath.Path, desUrl)
				//movieFUrl := strings.ReplaceAll(oneMovieMixInfo.PhysicalVideoFileFullPath, oneMovieDirPath.Path, desUrl)
				oneMovieInfo := backend2.MovieInfo{
					Name:                     filepath.Base(movieFUrl),
					DirRootUrl:               filepath.Dir(movieFUrl),
					VideoFPath:               oneMovieMixInfo.PhysicalVideoFileFullPath,
					VideoUrl:                 movieFUrl,
					MediaServerInsideVideoID: oneMovieMixInfo.VideoInfo.Id,
					SubFPathList:             make([]string, 0),
				}

				// 搜索字幕
				matchedSubFileByOneVideo, err := sub_helper.SearchMatchedSubFileByOneVideo(v.log, oneMovieMixInfo.PhysicalVideoFileFullPath)
				if err != nil {
					v.log.Errorln("SearchMatchedSubFileByOneVideo", err)
				}
				matchedSubFileByOneVideoUrl := make([]string, 0)
				for _, oneSubFPath := range matchedSubFileByOneVideo {
					oneSubFUrl := path_helper.ChangePhysicalPathToSharePath(oneSubFPath, oneMovieDirPath.Path, desUrl)
					//oneSubFUrl := strings.ReplaceAll(oneSubFPath, oneMovieDirPath.Path, desUrl)
					matchedSubFileByOneVideoUrl = append(matchedSubFileByOneVideoUrl, oneSubFUrl)
				}
				oneMovieInfo.SubFPathList = append(oneMovieInfo.SubFPathList, matchedSubFileByOneVideoUrl...)

				v.processLocker.Lock()
				movieInfos = append(movieInfos, oneMovieInfo)
				v.processLocker.Unlock()

				break
			}
		}

		return nil
	}
	// ----------------------------------------
	v.taskControl.SetCtxProcessFunc("updateLocalVideoCacheInfo", movieProcess, common2.ScanPlayedSubTimeOut)
	// ----------------------------------------
	for i, oneMovieMixInfo := range emby.MovieSubNeedDlEmbyMixInfoList {

		if oneMovieMixInfo.PhysicalVideoFileFullPath == "" {
			v.log.Warningln("oneMovieMixInfo.PhysicalVideoFileFullPath is empty")
			continue
		}

		// 放入队列
		err := v.taskControl.Invoke(&task_control.TaskData{
			Index: i,
			Count: len(emby.MovieSubNeedDlEmbyMixInfoList),
			DataEx: ScrabbleUpVideoMovieEmbyInput{
				OneMovieMixInfo: oneMovieMixInfo,
			},
		})
		if err != nil {
			v.log.Errorln("scrabbleUpVideoListEmby.movieProcess.taskControl.Invoke", err)
			break
		}
	}
	v.taskControl.Hold()
	// ----------------------------------------
	// Emby 过滤，连续剧
	seriesProcess := func(ctx context.Context, inData interface{}) error {

		taskData := inData.(*task_control.TaskData)
		scrabbleUpVideoSeriesEmbyInput := taskData.DataEx.(ScrabbleUpVideoSeriesEmbyInput)

		oneSeasonInfo := scrabbleUpVideoSeriesEmbyInput.OneSeasonInfo
		oneEpsMixInfo := scrabbleUpVideoSeriesEmbyInput.OneEpsMixInfo
		// 首先需要找到对应的最长的视频媒体库路径，x://ABC  x://ABC/DEF
		for _, oneSeriesDirPath := range sortSeriesPaths {

			if strings.HasPrefix(oneEpsMixInfo.PhysicalVideoFileFullPath, oneSeriesDirPath.Path) {
				// 匹配上了
				v.processLocker.Lock()
				desUrl, found := pathUrlMap[oneSeriesDirPath.Path]
				if found == false {
					v.processLocker.Unlock()
					// 没有找到对应的 URL
					v.log.Warningln("scrabbleUpVideoListEmby.seriesProcess.pathUrlMap", oneSeriesDirPath.Path)
					continue
				}
				v.processLocker.Unlock()

				videoFileName := filepath.Base(oneEpsMixInfo.PhysicalVideoFileFullPath)
				epsVideoNfoInfo, err := decode.GetVideoNfoInfo4OneSeriesEpisode(oneEpsMixInfo.PhysicalVideoFileFullPath)
				if err != nil {
					v.log.Warningln("GetVideoInfoFromFileName", err)
					continue
				}
				// 匹配上了前缀就替换这个，并记录
				epsFUrl := path_helper.ChangePhysicalPathToSharePath(oneEpsMixInfo.PhysicalVideoFileFullPath, oneSeriesDirPath.Path, desUrl)
				//epsFUrl := strings.ReplaceAll(oneEpsMixInfo.PhysicalVideoFileFullPath, oneSeriesDirPath.Path, desUrl)
				oneVideoInfo := backend2.OneVideoInfo{
					Name:                     videoFileName,
					VideoFPath:               oneEpsMixInfo.PhysicalVideoFileFullPath,
					VideoUrl:                 epsFUrl,
					Season:                   epsVideoNfoInfo.Season,
					Episode:                  epsVideoNfoInfo.Episode,
					MediaServerInsideVideoID: oneEpsMixInfo.VideoInfo.Id,
					SubFPathList:             make([]string, 0),
				}

				// 搜索字幕
				matchedSubFileByOneVideo, err := sub_helper.SearchMatchedSubFileByOneVideo(v.log, oneEpsMixInfo.PhysicalVideoFileFullPath)
				if err != nil {
					v.log.Errorln("SearchMatchedSubFileByOneVideo", err)
				}
				matchedSubFileByOneVideoUrl := make([]string, 0)
				for _, oneSubFPath := range matchedSubFileByOneVideo {
					oneSubFUrl := path_helper.ChangePhysicalPathToSharePath(oneSubFPath, oneSeriesDirPath.Path, desUrl)
					//oneSubFUrl := strings.ReplaceAll(oneSubFPath, oneSeriesDirPath.Path, desUrl)
					matchedSubFileByOneVideoUrl = append(matchedSubFileByOneVideoUrl, oneSubFUrl)
				}
				oneVideoInfo.SubFPathList = append(oneVideoInfo.SubFPathList, matchedSubFileByOneVideoUrl...)

				v.processLocker.Lock()
				oneSeasonInfo.OneVideoInfos = append(oneSeasonInfo.OneVideoInfos, oneVideoInfo)
				v.processLocker.Unlock()

				break
			}
		}
		return nil
	}
	// ----------------------------------------
	v.taskControl.SetCtxProcessFunc("updateLocalVideoCacheInfo", seriesProcess, common2.ScanPlayedSubTimeOut)
	// ----------------------------------------
	for seriesName, oneSeriesMixInfo := range emby.SeriesSubNeedDlEmbyMixInfoMap {

		var oneSeasonInfo backend2.SeasonInfo
		// 需要先得到 oneSeasonInfo 的信息
		for _, oneEpsMixInfo := range oneSeriesMixInfo {

			if oneEpsMixInfo.PhysicalVideoFileFullPath == "" {
				v.log.Warningln("oneEpsMixInfo.PhysicalVideoFileFullPath is empty")
				continue
			}
			// 首先需要找到对应的最长的视频媒体库路径，x://ABC  x://ABC/DEF
			for _, oneSeriesDirPath := range sortSeriesPaths {

				if strings.HasPrefix(oneEpsMixInfo.PhysicalVideoFileFullPath, oneSeriesDirPath.Path) == false {
					continue
				}
				// 匹配上了
				desUrl, found := pathUrlMap[oneSeriesDirPath.Path]
				if found == false {
					// 没有找到对应的 URL
					v.log.Warningln("scrabbleUpVideoListEmby.seriesProcess.pathUrlMap", oneSeriesDirPath.Path)
					continue
				}

				//v.log.Infoln("scrabbleUpVideoListEmby", oneEpsMixInfo.PhysicalVideoFileFullPath, oneSeriesDirPath.Path, desUrl)
				dirRootUrl := path_helper.ChangePhysicalPathToSharePath(oneEpsMixInfo.PhysicalSeriesRootDir, oneSeriesDirPath.Path, desUrl)
				//dirRootUrl := strings.ReplaceAll(oneEpsMixInfo.PhysicalSeriesRootDir, oneSeriesDirPath.Path, desUrl)

				oneSeasonInfo = backend2.SeasonInfo{
					Name:          seriesName,
					RootDirPath:   oneEpsMixInfo.PhysicalSeriesRootDir,
					DirRootUrl:    dirRootUrl,
					OneVideoInfos: make([]backend2.OneVideoInfo, 0),
				}
				break
			}
			if oneSeasonInfo.Name != "" {
				// 这个结构初始化过了
				break
			}
		}

		if oneSeasonInfo.Name == "" {
			// 说明找了一圈没有找到匹配的，那么后续的也没必要继续
			v.log.Warningln("oneSeasonInfo.Name is empty")
			continue
		}

		// 然后再开始处理每一集的信息
		for i, oneEpsMixInfo := range oneSeriesMixInfo {

			if oneEpsMixInfo.PhysicalVideoFileFullPath == "" {
				v.log.Warningln("oneEpsMixInfo.PhysicalVideoFileFullPath is empty")
				continue
			}

			// 放入队列
			err := v.taskControl.Invoke(&task_control.TaskData{
				Index: i,
				Count: len(oneSeriesMixInfo),
				DataEx: ScrabbleUpVideoSeriesEmbyInput{
					OneSeasonInfo: &oneSeasonInfo,
					OneEpsMixInfo: oneEpsMixInfo,
				},
			})
			if err != nil {
				v.log.Errorln("scrabbleUpVideoListEmby.seriesProcess.taskControl.Invoke", err)
				break
			}
		}
		v.taskControl.Hold()

		seasonInfos = append(seasonInfos, oneSeasonInfo)
	}

	return movieInfos, seasonInfos
}

func (v *VideoScanAndRefreshHelper) refreshEmbySubList(embySettings *settings.EmbySettings, SkipWatched bool, maxRequestVideoNumber int) error {

	if v.embyHelper == nil {
		return nil
	}

	bRefresh := false
	defer func() {
		if bRefresh == true {
			v.log.Infoln("Refresh Emby Sub List Success")
		} else {
			v.log.Errorln("Refresh Emby Sub List Error")
		}
	}()
	v.log.Infoln("Refresh Emby Sub List Start...")
	//------------------------------------------------------
	bRefresh, err := v.embyHelper.RefreshEmbySubList(embySettings, SkipWatched, maxRequestVideoNumber)
	if err != nil {
		return err
	}

	return nil
}

// updateLocalVideoCacheInfo 将扫描到的信息缓存到本地中，用于后续的 Video 展示界面 和 Emby IMDB ID 匹配进行路径的转换
func (v *VideoScanAndRefreshHelper) updateLocalVideoCacheInfo(scanVideoResult *ScanVideoResult) error {
	// 这里只使用 Normal 情况下获取到的信息
	if scanVideoResult.Normal == nil {
		return nil
	}
	// ------------------------------------------------------------------------------
	// 电影
	movieProcess := func(ctx context.Context, inData interface{}) error {

		taskData := inData.(*task_control.TaskData)
		movieInputData := taskData.DataEx.(TaskInputData)
		v.log.Infoln("updateLocalVideoCacheInfo", movieInputData.Index, movieInputData.InputPath)
		videoImdbInfo, err := decode.GetVideoNfoInfo4Movie(movieInputData.InputPath)
		if err != nil {
			// 允许的错误，跳过，继续进行文件名的搜索
			v.log.Warningln("GetVideoNfoInfo4Movie", movieInputData.Index, err)
			return err
		}
		// 获取 IMDB 信息
		localIMDBInfo, err := imdb_helper.GetIMDBInfoFromVideoNfoInfo(v.fileDownloader.MediaInfoDealers, videoImdbInfo)
		if err != nil {
			v.log.Warningln("GetIMDBInfoFromVideoNfoInfo,IMDB:", videoImdbInfo.ImdbId, movieInputData.InputPath, err)
			return err
		}

		movieDirPath := filepath.Dir(movieInputData.InputPath)
		if (movieDirPath != "" && localIMDBInfo.RootDirPath != movieDirPath) || localIMDBInfo.IsMovie != true {
			// 更新数据
			localIMDBInfo.RootDirPath = movieDirPath
			localIMDBInfo.IsMovie = true
			dao.GetDb().Save(localIMDBInfo)
		}

		return nil
	}
	// ------------------------------------------------------------------------------
	v.taskControl.SetCtxProcessFunc("updateLocalVideoCacheInfo", movieProcess, common2.ScanPlayedSubTimeOut)
	// ------------------------------------------------------------------------------
	scanVideoResult.Normal.MoviesDirMap.Any(func(movieDirRootPath interface{}, movieFPath interface{}) bool {

		//oneMovieDirRootPath := movieDirRootPath.(string)
		for i, oneMovieFPath := range movieFPath.([]string) {
			err := v.taskControl.Invoke(&task_control.TaskData{
				Index: i,
				Count: len(movieFPath.([]string)),
				DataEx: TaskInputData{
					Index:     i,
					InputPath: oneMovieFPath,
				},
			})
			if err != nil {
				v.log.Errorln("updateLocalVideoCacheInfo.MoviesDirMap.Invoke", err)
				return true
			}
		}

		return false
	})
	v.taskControl.Hold()
	// ------------------------------------------------------------------------------
	seriesProcess := func(ctx context.Context, inData interface{}) error {

		taskData := inData.(*task_control.TaskData)
		seriesInputData := taskData.DataEx.(TaskInputData)
		v.log.Infoln("updateLocalVideoCacheInfo", seriesInputData.Index, seriesInputData.InputPath)

		videoNfoInfo, err := decode.GetVideoNfoInfo4SeriesDir(seriesInputData.InputPath)
		if err != nil {
			v.log.Warningln("GetVideoNfoInfo4SeriesDir", seriesInputData.InputPath, err)
			return err
		}

		// 获取 IMDB 信息
		localIMDBInfo, err := imdb_helper.GetIMDBInfoFromVideoNfoInfo(v.fileDownloader.MediaInfoDealers, videoNfoInfo)
		if err != nil {
			v.log.Warningln("GetIMDBInfoFromVideoNfoInfo,IMDB:", videoNfoInfo.ImdbId, seriesInputData.InputPath, err)
			return err
		}
		if (seriesInputData.InputPath != "" && localIMDBInfo.RootDirPath != seriesInputData.InputPath) || localIMDBInfo.IsMovie != false {
			// 更新数据
			localIMDBInfo.RootDirPath = seriesInputData.InputPath
			localIMDBInfo.IsMovie = false
			dao.GetDb().Save(localIMDBInfo)
		}

		return nil
	}
	// ------------------------------------------------------------------------------
	v.taskControl.SetCtxProcessFunc("updateLocalVideoCacheInfo", seriesProcess, common2.ScanPlayedSubTimeOut)
	// ------------------------------------------------------------------------------
	// 连续剧
	scanVideoResult.Normal.SeriesDirMap.Each(func(seriesRootPathName interface{}, seriesNames interface{}) {

		for i, oneSeriesRootDir := range seriesNames.([]string) {
			err := v.taskControl.Invoke(&task_control.TaskData{
				Index: i,
				Count: len(seriesNames.([]string)),
				DataEx: TaskInputData{
					Index:     i,
					InputPath: oneSeriesRootDir,
				},
			})
			if err != nil {
				v.log.Errorln(err)
				return
			}
		}
	})
	v.taskControl.Hold()

	return nil
}

func (v *VideoScanAndRefreshHelper) filterMovieAndSeriesNeedDownloadNormal(normal *NormalScanVideoResult, scanLogic *scan_logic.ScanLogic) error {
	// ----------------------------------------
	// Normal 过滤，电影
	movieProcess := func(ctx context.Context, inData interface{}) error {

		taskData := inData.(*task_control.TaskData)
		movieInputData := taskData.DataEx.(TaskInputData)
		if v.subSupplierHub.MovieNeedDlSub(v.fileDownloader.MediaInfoDealers, movieInputData.InputPath, v.NeedForcedScanAndDownSub) == false {
			return nil
		}
		bok, err := v.downloadQueue.Add(*task_queue2.NewOneJob(
			common2.Movie, movieInputData.InputPath, task_queue.DefaultTaskPriorityLevel,
		))
		if err != nil {
			v.log.Errorln("filterMovieAndSeriesNeedDownloadNormal.Movie.NewOneJob", err)
			return err
		}
		if bok == false {
			v.log.Warningln(common2.Movie.String(), movieInputData.InputPath, "downloadQueue isExisted")
		}

		return nil
	}
	// ----------------------------------------
	v.taskControl.SetCtxProcessFunc("updateLocalVideoCacheInfo", movieProcess, common2.ScanPlayedSubTimeOut)
	// ----------------------------------------
	normal.MoviesDirMap.Any(func(movieDirRootPath interface{}, movieFPath interface{}) bool {

		//oneMovieDirRootPath := movieDirRootPath.(string)
		for i, oneMovieFPath := range movieFPath.([]string) {

			// 判断是否需要跳过
			if scanLogic.Get(0, oneMovieFPath) == true {
				v.log.Debugln("filterMovieAndSeriesNeedDownloadNormal.Movie", oneMovieFPath, "skip")
				continue
			}
			// 放入队列
			err := v.taskControl.Invoke(&task_control.TaskData{
				Index: i,
				Count: len(movieFPath.([]string)),
				DataEx: TaskInputData{
					Index:     i,
					InputPath: oneMovieFPath,
				},
			})
			if err != nil {
				v.log.Errorln(err)
				return true
			}
		}

		return false
	})
	v.taskControl.Hold()
	// ----------------------------------------
	// Normal 过滤，连续剧
	seriesProcess := func(ctx context.Context, inData interface{}) error {

		taskData := inData.(*task_control.TaskData)
		seriesInputData := taskData.DataEx.(TaskInputData)
		// 因为可能回去 Web 获取 IMDB 信息，所以这里的错误不返回
		bNeedDlSub, seriesInfo, err := v.subSupplierHub.SeriesNeedDlSub(
			v.fileDownloader.MediaInfoDealers,
			seriesInputData.InputPath,
			v.NeedForcedScanAndDownSub, false)
		if err != nil {
			v.log.Errorln("filterMovieAndSeriesNeedDownloadNormal.SeriesNeedDlSub", err)
			return err
		}
		if bNeedDlSub == false {
			return nil
		}

		for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

			// 判断是否需要跳过
			if scanLogic.Get(1, episodeInfo.FileFullPath) == true {
				v.log.Debugln("filterMovieAndSeriesNeedDownloadNormal.Series", episodeInfo.FileFullPath, "skip")
				continue
			}

			// 放入队列
			oneJob := task_queue2.NewOneJob(
				common2.Series, episodeInfo.FileFullPath, task_queue.DefaultTaskPriorityLevel,
			)
			oneJob.Season = episodeInfo.Season
			oneJob.Episode = episodeInfo.Episode
			oneJob.SeriesRootDirPath = seriesInfo.DirPath

			bok, err := v.downloadQueue.Add(*oneJob)
			if err != nil {
				v.log.Errorln("filterMovieAndSeriesNeedDownloadNormal.Series.NewOneJob", err)
				continue
			}
			if bok == false {
				v.log.Warningln(common2.Series.String(), episodeInfo.FileFullPath, "downloadQueue isExisted")
			}
		}

		return nil
	}
	// ----------------------------------------
	v.taskControl.SetCtxProcessFunc("updateLocalVideoCacheInfo", seriesProcess, common2.ScanPlayedSubTimeOut)
	// ----------------------------------------
	// seriesDirMap: dir <--> seriesList
	normal.SeriesDirMap.Each(func(seriesRootPathName interface{}, seriesNames interface{}) {

		for i, oneSeriesRootDir := range seriesNames.([]string) {

			// 放入队列
			err := v.taskControl.Invoke(&task_control.TaskData{
				Index: i,
				Count: len(seriesNames.([]string)),
				DataEx: TaskInputData{
					Index:     i,
					InputPath: oneSeriesRootDir,
				},
			})
			if err != nil {
				v.log.Errorln(err)
				return
			}
		}
	})
	v.taskControl.Hold()
	// ----------------------------------------
	return nil
}

func (v *VideoScanAndRefreshHelper) filterMovieAndSeriesNeedDownloadEmby(emby *EmbyScanVideoResult, scanLogic *scan_logic.ScanLogic) error {

	playedVideoIdMap := make(map[string]bool)
	if settings.Get().EmbySettings.SkipWatched == true {
		playedVideoIdMap = v.embyHelper.GetVideoIDPlayedMap(settings.Get().EmbySettings, settings.Get().EmbySettings.MaxRequestVideoNumber)
	}
	// ----------------------------------------
	// Emby 过滤，电影
	for _, oneMovieMixInfo := range emby.MovieSubNeedDlEmbyMixInfoList {

		// 判断是否需要跳过
		if scanLogic.Get(0, oneMovieMixInfo.PhysicalVideoFileFullPath) == true {
			v.log.Debugln("filterMovieAndSeriesNeedDownloadEmby.Movie", oneMovieMixInfo.PhysicalVideoFileFullPath, "skip")
			continue
		}

		// 放入队列
		if v.subSupplierHub.MovieNeedDlSub(v.fileDownloader.MediaInfoDealers, oneMovieMixInfo.PhysicalVideoFileFullPath, v.NeedForcedScanAndDownSub) == false {
			continue
		}
		nowOneJob := task_queue2.NewOneJob(
			common2.Movie, oneMovieMixInfo.PhysicalVideoFileFullPath, task_queue.DefaultTaskPriorityLevel,
			oneMovieMixInfo.VideoInfo.Id,
		)
		bok, err := v.downloadQueue.Add(*nowOneJob)
		if err != nil {
			v.log.Errorln("filterMovieAndSeriesNeedDownloadEmby.Movie.NewOneJob", err)
			continue
		}
		if bok == false {

			v.log.Warningln(common2.Movie.String(), oneMovieMixInfo.PhysicalVideoFileFullPath, "downloadQueue isExisted")
			// 如果任务存在了，需要判断这个任务的视频已经被看过了，如果是，那么就需要标记 Skip
			_, bok = playedVideoIdMap[oneMovieMixInfo.VideoInfo.Id]
			if bok == true {
				// 找到了,那么就是看过了
				nowOneJob.JobStatus = task_queue2.Ignore
				bok, err = v.downloadQueue.Update(*nowOneJob)
				if err != nil {
					v.log.Errorln("filterMovieAndSeriesNeedDownloadEmby.Movie.Update", err)
					continue
				}
				if bok == false {
					v.log.Warningln(common2.Movie.String(), oneMovieMixInfo.PhysicalVideoFileFullPath, "downloadQueue isExisted")
					continue
				}
			}
		}
	}
	// Emby 过滤，连续剧
	for _, embyMixInfos := range emby.SeriesSubNeedDlEmbyMixInfoMap {

		if len(embyMixInfos) < 1 {
			continue
		}

		// 只需要从一集取信息即可
		for _, mixInfo := range embyMixInfos {

			// 判断是否需要跳过
			if scanLogic.Get(0, mixInfo.PhysicalVideoFileFullPath) == true {
				v.log.Debugln("filterMovieAndSeriesNeedDownloadEmby.Series", mixInfo.PhysicalVideoFileFullPath, "skip")
				continue
			}

			// 在 GetRecentlyAddVideoListWithNoChineseSubtitle 的时候就进行了筛选，所以这里就直接加入队列了
			// 放入队列
			oneJob := task_queue2.NewOneJob(
				common2.Series, mixInfo.PhysicalVideoFileFullPath, task_queue.DefaultTaskPriorityLevel,
				mixInfo.VideoInfo.Id,
			)

			info, err := decode.GetVideoNfoInfoFromEpisode(mixInfo.PhysicalVideoFileFullPath)
			if err != nil {
				v.log.Warningln("filterMovieAndSeriesNeedDownloadEmby.Series.GetVideoInfoFromFileFullPath", err)
				continue
			}
			oneJob.Season = info.Season
			oneJob.Episode = info.Episode
			oneJob.SeriesRootDirPath = mixInfo.PhysicalSeriesRootDir

			bok, err := v.downloadQueue.Add(*oneJob)
			if err != nil {
				v.log.Errorln("filterMovieAndSeriesNeedDownloadEmby.Series.NewOneJob", err)
				continue
			}
			if bok == false {

				v.log.Warningln(common2.Series.String(), mixInfo.PhysicalVideoFileFullPath, "downloadQueue isExisted")
				// 如果任务存在了，需要判断这个任务的视频已经被看过了，如果是，那么就需要标记 Skip
				_, bok = playedVideoIdMap[mixInfo.VideoInfo.Id]
				if bok == true {
					// 找到了,那么就是看过了
					oneJob.JobStatus = task_queue2.Ignore
					bok, err = v.downloadQueue.Update(*oneJob)
					if err != nil {
						v.log.Errorln("filterMovieAndSeriesNeedDownloadEmby.Series.Update", err)
						continue
					}
					if bok == false {
						v.log.Warningln(common2.Series.String(), mixInfo.PhysicalVideoFileFullPath, "downloadQueue isExisted")
						continue
					}
				}
			}
		}
	}

	return nil
}

// getUpdateVideoListFromEmby 这里首先会进行近期影片的获取，然后对这些影片进行刷新，然后在获取字幕列表，最终得到需要字幕获取的 video 列表
func (v *VideoScanAndRefreshHelper) getUpdateVideoListFromEmby() ([]emby.EmbyMixInfo, map[string][]emby.EmbyMixInfo, error) {
	if v.embyHelper == nil {
		return nil, nil, nil
	}
	defer func() {
		v.log.Infoln("getUpdateVideoListFromEmby End")
	}()
	v.log.Infoln("getUpdateVideoListFromEmby Start...")
	//------------------------------------------------------
	var err error
	var movieList []emby.EmbyMixInfo
	var seriesSubNeedDlMap map[string][]emby.EmbyMixInfo //  多个需要搜索字幕的连续剧目录，连续剧文件夹名称 -- 每一集的 EmbyMixInfo List
	movieList, seriesSubNeedDlMap, err = v.embyHelper.GetRecentlyAddVideoListWithNoChineseSubtitle(settings.Get().EmbySettings, v.NeedForcedScanAndDownSub)
	if err != nil {
		return nil, nil, err
	}
	// 输出调试信息
	v.log.Debugln("getUpdateVideoListFromEmby - DebugInfo - movieFileFullPathList Start")
	for _, info := range movieList {
		v.log.Debugln(info.PhysicalVideoFileFullPath)
	}
	v.log.Debugln("getUpdateVideoListFromEmby - DebugInfo - movieFileFullPathList End")

	v.log.Debugln("getUpdateVideoListFromEmby - DebugInfo - seriesSubNeedDlMap Start")
	for s := range seriesSubNeedDlMap {
		v.log.Debugln(s)
	}
	v.log.Debugln("getUpdateVideoListFromEmby - DebugInfo - seriesSubNeedDlMap End")

	return movieList, seriesSubNeedDlMap, nil
}

func (v *VideoScanAndRefreshHelper) RestoreFixTimelineBK() error {

	defer v.log.Infoln("End Restore Fix Timeline BK")
	v.log.Infoln("Start Restore Fix Timeline BK...")
	//------------------------------------------------------
	_, err := subTimelineFixerPKG.Restore(v.log, settings.Get().CommonSettings.MoviePaths, settings.Get().CommonSettings.SeriesPaths)
	if err != nil {
		return err
	}
	return nil
}

type ScanVideoResult struct {
	Normal *NormalScanVideoResult
	Emby   *EmbyScanVideoResult
}

type NormalScanVideoResult struct {
	MoviesDirMap *treemap.Map
	SeriesDirMap *treemap.Map
}

type EmbyScanVideoResult struct {
	MovieSubNeedDlEmbyMixInfoList []emby.EmbyMixInfo
	SeriesSubNeedDlEmbyMixInfoMap map[string][]emby.EmbyMixInfo
}

type TaskInputData struct {
	Index     int
	InputPath string
}

type ScrabbleUpVideoMovieNormalInput struct {
	OneMovieDirRootPath string
	OneMovieFPath       string
}

type ScrabbleUpVideoSeriesNormalInput struct {
	OneSeriesRootDir      string
	OneSeriesRootPathName string
}

type ScrabbleUpVideoMovieEmbyInput struct {
	OneMovieMixInfo emby.EmbyMixInfo
}

type ScrabbleUpVideoSeriesEmbyInput struct {
	OneSeasonInfo *backend2.SeasonInfo
	OneEpsMixInfo emby.EmbyMixInfo
}
