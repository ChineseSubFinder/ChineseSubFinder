package scan_played_video_subinfo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ifaces"
	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	embyHelper "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/emby_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/mix_media_info"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/imdb_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_file_hash"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/emby"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_share_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/task_control"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type ScanPlayedVideoSubInfo struct {
	log            *logrus.Logger
	fileDownloader *file_downloader.FileDownloader
	embyHelper     *embyHelper.EmbyHelper
	taskControl    *task_control.TaskControl
	canceled       bool
	canceledLock   sync.Mutex

	movieSubMap  map[string]string
	seriesSubMap map[string]string

	subFormatter ifaces.ISubFormatter

	shareRootDir string

	imdbInfoCache            map[string]*models.IMDBInfo
	cacheImdbInfoCacheLocker sync.Mutex
}

func NewScanPlayedVideoSubInfo(log *logrus.Logger, fileDownloader *file_downloader.FileDownloader) (*ScanPlayedVideoSubInfo, error) {
	var err error
	var scanPlayedVideoSubInfo ScanPlayedVideoSubInfo
	scanPlayedVideoSubInfo.log = log
	// 下载实例
	scanPlayedVideoSubInfo.fileDownloader = fileDownloader
	// 检测是否某些参数超出范围
	settings.Get().Check()
	// 初始化 Emby API 接口
	if settings.Get().EmbySettings.Enable == true && settings.Get().EmbySettings.AddressUrl != "" &&
		settings.Get().EmbySettings.APIKey != "" {

		scanPlayedVideoSubInfo.embyHelper = embyHelper.NewEmbyHelper(fileDownloader.MediaInfoDealers)
	}

	// 初始化任务控制
	scanPlayedVideoSubInfo.taskControl, err = task_control.NewTaskControl(settings.Get().CommonSettings.Threads, log)
	if err != nil {
		return nil, err
	}
	// 字幕命名格式解析器
	scanPlayedVideoSubInfo.subFormatter = emby.NewFormatter()
	// 缓存目录的根目录
	shareRootDir, err := pkg.GetShareSubRootFolder()
	if err != nil {
		return nil, err
	}
	scanPlayedVideoSubInfo.shareRootDir = shareRootDir
	// 初始化缓存
	scanPlayedVideoSubInfo.imdbInfoCache = make(map[string]*models.IMDBInfo)

	return &scanPlayedVideoSubInfo, nil
}

func (s *ScanPlayedVideoSubInfo) Cancel() {

	defer func() {
		s.log.Infoln("ScanPlayedVideoSubInfo.Cancel()")
	}()

	s.canceledLock.Lock()
	s.canceled = true
	s.canceledLock.Unlock()

	s.taskControl.Release()
}

func (s *ScanPlayedVideoSubInfo) GetPlayedItemsSubtitle(embySettings *settings.EmbySettings, maxRequestVideoNumber int) (bool, error) {

	var err error
	// 是否是通过 emby_helper api 获取的列表
	if s.embyHelper == nil {
		// 没有填写 emby_helper api 的信息，那么就跳过
		s.log.Infoln("Skip ScanPlayedVideoSubInfo, Emby Settings is null")
		return false, nil
	}

	s.movieSubMap, s.seriesSubMap, err = s.embyHelper.GetPlayedItemsSubtitle(embySettings, maxRequestVideoNumber)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Clear 清理无效的缓存字幕信息
func (s *ScanPlayedVideoSubInfo) Clear() {

	defer func() {
		s.log.Infoln("ScanPlayedVideoSubInfo.Clear Sub End")
		s.log.Infoln("----------------------------------------------")
	}()

	s.log.Infoln("-----------------------------------------------")
	s.log.Infoln("ScanPlayedVideoSubInfo.Clear Sub Start...")

	var imdbInfos []models.IMDBInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().Preload("VideoSubInfos").Find(&imdbInfos)

	// 同时需要把不在数据库记录的字幕给删除，那么就需要把数据库查询出来的给做成 map
	dbSubMap := make(map[string]int)
	for _, info := range imdbInfos {

		for _, oneSubInfo := range info.VideoSubInfos {

			s.log.Infoln("ScanPlayedVideoSubInfo.Clear Sub", oneSubInfo.SubName)

			// 转换到绝对路径
			cacheSubFPath := filepath.Join(s.shareRootDir, oneSubInfo.StoreRPath)
			if pkg.IsFile(cacheSubFPath) == false {
				// 如果文件不存在，那么就删除之前的关联
				// 关联删除了，但是不会删除这些对象，所以后续还需要再次删除
				s.delSubInfo(&info, &oneSubInfo)

				s.log.Debugln("ScanPlayedVideoSubInfo.Clear Sub delSubInfo", oneSubInfo.SubName)

				continue
			}

			dbSubMap[oneSubInfo.StoreRPath] = 0
		}
	}
	// 搜索缓存文件夹所有的字幕出来，对比上面的 map 进行比较
	subFiles, err := sub_parser_hub.SearchMatchedSubFile(s.log, s.shareRootDir)
	if err != nil {
		return
	}

	for _, file := range subFiles {

		subRelPath, err := filepath.Rel(s.shareRootDir, file)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan.Rel", file, err)
			continue
		}

		_, bok := dbSubMap[subRelPath]
		if bok == false {

			err = os.Remove(file)
			s.log.Debugln("ScanPlayedVideoSubInfo.Clear Sub Remove", file)
			if err != nil {
				s.log.Debugln("ScanPlayedVideoSubInfo.Clear Sub Remove", file, err)
				continue
			}
		}
	}
}

func (s *ScanPlayedVideoSubInfo) Scan() error {

	// Emby 观看的列表
	{
		// 从数据库中查询出所有的 IMDBInfo
		// 清空缓存
		s.imdbInfoCache = make(map[string]*models.IMDBInfo)
		// -----------------------------------------------------
		// 并发控制
		s.taskControl.SetCtxProcessFunc("ScanSubPlayedPool", s.scan, common2.ScanPlayedSubTimeOut)
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
	}

	// 使用 Http API 标记的已观看列表
	{
		// TODO 暂时屏蔽掉 http api 提交的已看字幕的接口上传
		if false {
			// 下面需要把给出外部的 HTTP API 提交的视频和字幕信息(ThirdPartSetVideoPlayedInfo)进行判断，存入数据库
			shareRootDir, err := pkg.GetShareSubRootFolder()
			if err != nil {
				return err
			}

			var videoPlayedInfos []models.ThirdPartSetVideoPlayedInfo
			dao.GetDb().Find(&videoPlayedInfos)

			for i, thirdPartSetVideoPlayedInfo := range videoPlayedInfos {
				// 先要判断这个是 Movie 还是 Series
				// 因为设计这个 API 的时候为了简化提交的参数，也假定传入的可能不是正确的分类（电影or连续剧）
				// 所以只能比较傻的，低效率的匹配映射的目录来做到识别是哪个分类的

				bFoundMovie := false
				bFoundSeries := false
				for _, moviePath := range settings.Get().CommonSettings.MoviePaths {
					// 先判断类型是否是 Movie
					if strings.HasPrefix(thirdPartSetVideoPlayedInfo.PhysicalVideoFileFullPath, moviePath) == true {
						bFoundMovie = true
						break
					}
				}
				if bFoundMovie == false {
					for _, seriesPath := range settings.Get().CommonSettings.SeriesPaths {
						// 判断是否是 Series
						if strings.HasPrefix(thirdPartSetVideoPlayedInfo.PhysicalVideoFileFullPath, seriesPath) == true {
							bFoundSeries = true
							break
						}
					}
				}

				if bFoundMovie == false && bFoundSeries == false {
					// 说明提交的这个视频文件无法匹配电影或者连续剧的目录前缀
					s.log.Warningln("Not matched Movie and Series Prefix Path", thirdPartSetVideoPlayedInfo.PhysicalVideoFileFullPath)
					continue
				}

				IsMovie := false
				videoTypes := common2.Movie
				if bFoundMovie == true {
					videoTypes = common2.Movie
					IsMovie = true
				}
				if bFoundSeries == true {
					videoTypes = common2.Series
					IsMovie = false
				}

				tmpSubFPath := filepath.Join(filepath.Dir(thirdPartSetVideoPlayedInfo.PhysicalVideoFileFullPath), thirdPartSetVideoPlayedInfo.SubName)
				s.dealOneVideo(i, thirdPartSetVideoPlayedInfo.PhysicalVideoFileFullPath, tmpSubFPath, videoTypes.String(), shareRootDir, IsMovie, s.imdbInfoCache)
			}
		}
	}

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

	shareRootDir, err := pkg.GetShareSubRootFolder()
	if err != nil {
		return err
	}

	index := 0
	for videoFPath, orgSubFPath := range scanInputData.Videos {

		index++
		stage := make(chan interface{}, 1)
		go func() {
			defer func() {
				close(stage)
			}()
			s.dealOneVideo(index, videoFPath, orgSubFPath, videoTypes, shareRootDir, scanInputData.IsMovie, s.imdbInfoCache)
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

	if pkg.IsFile(orgSubFPath) == false {

		s.log.Errorln("Skip", orgSubFPath, "not exist")
		return
	}

	s.log.Debugln(0)

	// 通过视频的绝对路径，从本地的视频文件对应的 nfo 获取到这个视频的 IMDB ID,
	var err error
	imdbInfoFromVideoFile, err := imdb_helper.GetIMDBInfoFromVideoFile(s.fileDownloader.MediaInfoDealers, videoFPath, isMovie)
	if err != nil {
		s.log.Errorln("GetIMDBInfoFromVideoFile", err)
		return
	}

	s.log.Debugln(1)

	// 使用本程序的 hash 的算法，得到视频的唯一 ID
	fileHash, err := sub_file_hash.Calculate(videoFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".ComputeFileHash", videoFPath, err)
		return
	}
	s.log.Debugln(2)

	var imdbInfo *models.IMDBInfo
	var ok bool
	// 先把 IMDB 信息查询查来，不管是从数据库还是网络（查询出来也得写入到数据库）
	s.cacheImdbInfoCacheLocker.Lock()
	imdbInfo, ok = imdbInfoCache[imdbInfoFromVideoFile.IMDBID]
	s.cacheImdbInfoCacheLocker.Unlock()
	if ok == false {
		s.cacheImdbInfoCacheLocker.Lock()
		imdbInfoCache[imdbInfoFromVideoFile.IMDBID] = imdbInfoFromVideoFile
		imdbInfo = imdbInfoFromVideoFile
		s.cacheImdbInfoCacheLocker.Unlock()
	}
	s.log.Debugln(3)

	// 这里需要判断是否已经获取了 TMDB Info，如果没有则需要去获取
	if imdbInfo.TmdbId == "" {

		s.log.Debugln("3-2")
		videoType := "movie"
		if imdbInfo.IsMovie == false {
			videoType = "series"
		}
		_, err = mix_media_info.GetMediaInfoAndSave(
			s.fileDownloader.MediaInfoDealers,
			imdbInfo,
			imdbInfo.IMDBID, "imdb", videoType)
		if err != nil {
			s.log.Errorln("dealOneVideo.GetMediaInfoAndSave,", imdbInfo.Name, err)
			return
		}
	}

	s.log.Debugln("3-2")

	// 当前扫描到的找个字幕的 sha256 是否已经存在与缓存中了
	tmpSHA256String, err := pkg.GetFileSHA256String(orgSubFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, "orgSubFPath.GetFileSHA256String", videoFPath, err)
		return
	}

	s.log.Debugln(4)

	// 判断找到的关联字幕信息是否已经存在了，不存在则新增关联
	for _, cacheInfo := range imdbInfo.VideoSubInfos {

		if cacheInfo.SHA256 == tmpSHA256String {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, "SHA256 Exist == true, Skip", orgSubFPath)
			return
		}
	}

	s.log.Debugln(5)

	// 新增插入
	// 把现有的字幕 copy 到缓存目录中
	bok, subCacheFPath := sub_share_center.CopySub2Cache(s.log, orgSubFPath, imdbInfo.IMDBID, imdbInfo.Year, false)
	if bok == false {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".CopySub2Cache", orgSubFPath, err)
		return
	}

	s.log.Debugln(6)

	// 不存在，插入，建立关系
	bok, fileInfo, err := s.fileDownloader.SubParserHub.DetermineFileTypeFromFile(subCacheFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".DetermineFileTypeFromFile", imdbInfo.IMDBID, err)
		return
	}
	if bok == false {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".DetermineFileTypeFromFile == false", imdbInfo.IMDBID)
		return
	}

	s.log.Debugln(7)

	// 特指 emby 字幕的情况
	_, _, _, _, extraSubPreName := s.subFormatter.IsMatchThisFormat(filepath.Base(subCacheFPath))
	// 转相对路径存储
	subRelPath, err := filepath.Rel(shareRootDir, subCacheFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".Rel", imdbInfo.IMDBID, err)
		return
	}

	s.log.Debugln(8)

	// 计算需要插入字幕的 sha256
	saveSHA256String, err := pkg.GetFileSHA256String(subCacheFPath)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, "GetFileSHA256String", videoFPath, err)
		return
	}
	// 这个字幕文件是否已经存在了
	var videoSubInfos []models.VideoSubInfo
	dao.GetDb().Where("sha256 = ?", saveSHA256String).Find(&videoSubInfos)
	if len(videoSubInfos) > 0 {
		// 存在，跳过
		s.log.Infoln("ScanPlayedVideoSubInfo.Scan", videoTypes, "SHA256 Exist == true, Skip", orgSubFPath)
		return
	}

	s.log.Debugln(9)

	// 如果不存在，那么就标记这个字幕是未发送
	oneVideoSubInfo := models.NewVideoSubInfo(
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
	oneVideoSubInfo.IsSend = false

	if isMovie == false {
		// 连续剧的时候，如果可能应该获取是 第几季  第几集
		epsVideoNfoInfo, err := decode.GetVideoNfoInfo4OneSeriesEpisode(videoFPath)
		if err != nil {
			s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".GetVideoNfoInfo4OneSeriesEpisode", imdbInfo.Name, err)
			return
		}
		oneVideoSubInfo.Season = epsVideoNfoInfo.Season
		oneVideoSubInfo.Episode = epsVideoNfoInfo.Episode
	}

	s.log.Debugln(10)

	err = dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Append(oneVideoSubInfo)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", videoTypes, ".Append Association", oneVideoSubInfo.SubName, err)
		return
	}
}

// 如果文件不存在，那么就删除之前的关联
// 关联删除了，但是不会删除这些对象，所以后续还需要再次删除
func (s *ScanPlayedVideoSubInfo) delSubInfo(imdbInfo *models.IMDBInfo, cacheInfo *models.VideoSubInfo) bool {

	err := dao.GetDb().Model(imdbInfo).Association("VideoSubInfos").Delete(cacheInfo)
	if err != nil {
		s.log.Warningln("ScanPlayedVideoSubInfo.Scan", ".Delete Association", cacheInfo.SubName, err)
		return false
	}
	// 继续删除这个对象
	dao.GetDb().Delete(cacheInfo)
	s.log.Infoln("Delete Not Exist or SHA256 not the same， Sub Association", cacheInfo.SubName)

	return true
}

type ScanInputData struct {
	Videos  map[string]string
	IsMovie bool
}
