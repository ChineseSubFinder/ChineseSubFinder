package downloader

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/csf"

	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	markSystem "github.com/allanpk716/ChineseSubFinder/internal/logic/mark_system"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/pre_download_process"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	subCommon "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/task_queue"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	taskQueue2 "github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Downloader 实例化一次用一次，不要反复的使用，很多临时标志位需要清理。
type Downloader struct {
	settings                 *settings.Settings
	log                      *logrus.Logger
	fileDownloader           *file_downloader.FileDownloader
	ctx                      context.Context
	cancel                   context.CancelFunc
	subSupplierHub           *subSupplier.SubSupplierHub                  // 字幕提供源的集合，这个需要定时进行扫描，这些字幕源是否有效，以及下载验证码信息
	mk                       *markSystem.MarkingSystem                    // MarkingSystem，字幕的评价系统
	subFormatter             ifaces.ISubFormatter                         // 字幕格式化命名的实现
	subNameFormatter         subCommon.FormatterName                      // 从 inSubFormatter 推断出来
	subTimelineFixerHelperEx *sub_timeline_fixer.SubTimelineFixerHelperEx // 字幕时间轴校正
	downloaderLock           sync.Mutex                                   // 取消执行 task control 的 Lock
	downloadQueue            *task_queue.TaskQueue                        // 需要下载的视频的队列
	embyHelper               *embyHelper.EmbyHelper                       // Emby 的实例

	cacheLocker   sync.Mutex
	movieInfoMap  map[string]MovieInfo  // 给 Web 界面使用的，Key: VideoFPath
	seasonInfoMap map[string]SeasonInfo // 给 Web 界面使用的,Key: RootDirPath
}

func NewDownloader(inSubFormatter ifaces.ISubFormatter, fileDownloader *file_downloader.FileDownloader, downloadQueue *task_queue.TaskQueue) *Downloader {

	var downloader Downloader
	downloader.fileDownloader = fileDownloader
	downloader.subFormatter = inSubFormatter
	downloader.fileDownloader = fileDownloader
	downloader.log = fileDownloader.Log
	// 参入设置信息
	downloader.settings = fileDownloader.Settings
	// 检测是否某些参数超出范围
	downloader.settings.Check()
	// 这里就不单独弄一个 settings.SubNameFormatter 字段来传递值了，因为 inSubFormatter 就已经知道是什么 formatter 了
	downloader.subNameFormatter = subCommon.FormatterName(downloader.subFormatter.GetFormatterFormatterName())

	var sitesSequence = make([]string, 0)
	// TODO 这里写固定了抉择字幕的顺序
	sitesSequence = append(sitesSequence, common.SubSiteZiMuKu)
	sitesSequence = append(sitesSequence, common.SubSiteSubHd)
	sitesSequence = append(sitesSequence, common.SubSiteChineseSubFinder)
	sitesSequence = append(sitesSequence, common.SubSiteAssrt)
	sitesSequence = append(sitesSequence, common.SubSiteA4K)
	sitesSequence = append(sitesSequence, common.SubSiteShooter)
	sitesSequence = append(sitesSequence, common.SubSiteXunLei)
	downloader.mk = markSystem.NewMarkingSystem(downloader.log, sitesSequence, downloader.settings.AdvancedSettings.SubTypePriority)

	// 初始化，字幕校正的实例
	downloader.subTimelineFixerHelperEx = sub_timeline_fixer.NewSubTimelineFixerHelperEx(downloader.log, *downloader.settings.TimelineFixerSettings)

	if downloader.settings.AdvancedSettings.FixTimeLine == true {
		downloader.subTimelineFixerHelperEx.Check()
	}
	// 任务队列
	downloader.downloadQueue = downloadQueue
	// 单个任务的超时设置
	downloader.ctx, downloader.cancel = context.WithCancel(context.Background())
	// 用于字幕下载后的刷新
	if downloader.settings.EmbySettings.Enable == true {
		downloader.embyHelper = embyHelper.NewEmbyHelper(downloader.log, downloader.settings)
	}

	downloader.movieInfoMap = make(map[string]MovieInfo)
	downloader.seasonInfoMap = make(map[string]SeasonInfo)

	err := downloader.loadVideoListCache()
	if err != nil {
		downloader.log.Errorln("loadVideoListCache error:", err)
	}

	return &downloader
}

// SupplierCheck 检查字幕源是否有效，会影响后续的字幕源是否参与下载
func (d *Downloader) SupplierCheck() {

	defer func() {
		if p := recover(); p != nil {
			d.log.Errorln("Downloader.SupplierCheck() panic")
			my_util.PrintPanicStack(d.log)
		}
		d.downloaderLock.Unlock()

		d.log.Infoln("Download.SupplierCheck() End")
	}()

	d.downloaderLock.Lock()
	d.log.Infoln("Download.SupplierCheck() Start ...")

	//// 创建一个 chan 用于任务的中断和超时
	//done := make(chan interface{}, 1)
	//// 接收内部任务的 panic
	//panicChan := make(chan interface{}, 1)
	//
	//go func() {
	//	defer func() {
	//		if p := recover(); p != nil {
	//			panicChan <- p
	//		}
	//
	//		close(done)
	//		close(panicChan)
	//	}()
	// 下载前的初始化
	d.log.Infoln("PreDownloadProcess.Init().Check().Wait()...")

	if d.settings.SpeedDevMode == true {
		// 这里是调试使用的，指定了只用一个字幕源
		subSupplierHub := subSupplier.NewSubSupplierHub(csf.NewSupplier(d.fileDownloader))
		d.subSupplierHub = subSupplierHub
	} else {

		preDownloadProcess := pre_download_process.NewPreDownloadProcess(d.fileDownloader)
		err := preDownloadProcess.Init().Check().Wait()
		if err != nil {
			//done <- errors.New(fmt.Sprintf("NewPreDownloadProcess Error: %v", err))
			d.log.Errorln(errors.New(fmt.Sprintf("NewPreDownloadProcess Error: %v", err)))
		} else {
			// 更新 SubSupplierHub 实例
			d.subSupplierHub = preDownloadProcess.SubSupplierHub
			//done <- nil
		}
	}

	//	done <- nil
	//}()
	//
	//select {
	//case err := <-done:
	//	if err != nil {
	//		d.log.Errorln(err)
	//	}
	//	break
	//case p := <-panicChan:
	//	// 遇到内部的 panic，向外抛出
	//	panic(p)
	//case <-d.ctx.Done():
	//	{
	//		d.log.Errorln("cancel SupplierCheck")
	//		return
	//	}
	//}
}

// QueueDownloader 从字幕队列中取一个视频的字幕下载任务出来，并且开始下载
func (d *Downloader) QueueDownloader() {

	defer func() {
		if p := recover(); p != nil {
			d.log.Errorln("Downloader.QueueDownloader() panic")
			my_util.PrintPanicStack(d.log)
		}

		d.downloaderLock.Unlock()
		d.log.Debugln("Download.QueueDownloader() End")
	}()

	d.log.Debugln("Download.QueueDownloader() Try Start ...")
	d.downloaderLock.Lock()
	d.log.Debugln("Download.QueueDownloader() Start ...")

	var downloadCounter int64
	downloadCounter = 0
	// 移除查过三个月的 Done 任务
	d.downloadQueue.BeforeGetOneJob()
	// 从队列取数据出来，见《任务生命周期》
	bok, oneJob, err := d.downloadQueue.GetOneJob()
	if err != nil {
		d.log.Errorln("d.downloadQueue.GetOneWaitingJob()", err)
		return
	}
	if bok == false {
		d.log.Debugln("Download Queue Is Empty, Skip This Time")
		return
	}
	// --------------------------------------------------
	// 判断是否看过
	isPlayed := false
	if d.embyHelper != nil {
		// 在拿出来后，如果是有内部媒体服务器媒体 ID 的，那么就去查询是否已经观看过了
		isPlayed, err = d.embyHelper.IsVideoPlayed(oneJob.MediaServerInsideVideoID)
		if err != nil {
			d.log.Errorln("d.embyHelper.IsVideoPlayed()", oneJob.VideoFPath, err)
			return
		}
	}
	// 不管如何，只要是发现数据库中有 HTTP API 提交的信息，就认为是看过
	var videoPlayedInfos []models.ThirdPartSetVideoPlayedInfo
	dao.GetDb().Where("physical_video_file_full_path = ?", oneJob.VideoFPath).Find(&videoPlayedInfos)
	if len(videoPlayedInfos) > 0 {
		isPlayed = true
	}
	// --------------------------------------------------
	// 如果已经播放过 且 这个任务的优先级 > 3 ，不是很急的那种，说明是可以设置忽略继续下载的
	if isPlayed == true && oneJob.TaskPriority > task_queue.HighTaskPriorityLevel {
		// 播放过了，那么就标记 ignore
		oneJob.JobStatus = taskQueue2.Ignore
		bok, err = d.downloadQueue.Update(oneJob)
		if err != nil {
			d.log.Errorln("d.downloadQueue.Update()", err)
			return
		}
		if bok == false {
			d.log.Errorln("d.downloadQueue.Update() Failed")
			return
		}
		d.log.Infoln("Is Played, Ignore This Job")
		return
	}

	// 取出来后，需要标记为正在下载
	oneJob.JobStatus = taskQueue2.Downloading
	bok, err = d.downloadQueue.Update(oneJob)
	if err != nil {
		d.log.Errorln("d.downloadQueue.Update()", err)
		return
	}
	if bok == false {
		d.log.Errorln("d.downloadQueue.Update() Failed")
		return
	}
	// ------------------------------------------------------------------------
	// 开始标记，这个是单次扫描的开始，要注意格式，在日志的内部解析识别单个日志开头的时候需要特殊的格式
	d.log.Infoln("------------------------------------------")
	d.log.Infoln(log_helper.OnceSubsScanStart + "#" + oneJob.Id)
	// ------------------------------------------------------------------------
	defer func() {
		d.log.Infoln(log_helper.OnceSubsScanEnd)
		d.log.Infoln("------------------------------------------")
	}()

	downloadCounter++
	// 创建一个 chan 用于任务的中断和超时
	done := make(chan interface{}, 1)
	// 接收内部任务的 panic
	panicChan := make(chan interface{}, 1)

	go func() {
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
			close(done)
			close(panicChan)
			// 没下载完毕一次，进行一次缓存和 Chrome 的清理
			err = my_folder.ClearRootTmpFolder()
			if err != nil {
				d.log.Error("ClearRootTmpFolder", err)
			}
			my_util.CloseChrome(d.log)
		}()

		if oneJob.VideoType == common.Movie {
			// 电影
			// 具体的下载逻辑 func()
			done <- d.movieDlFunc(d.ctx, oneJob, downloadCounter)
		} else if oneJob.VideoType == common.Series {
			// 连续剧
			// 具体的下载逻辑 func()
			done <- d.seriesDlFunc(d.ctx, oneJob, downloadCounter)
		} else {
			d.log.Errorln("oneJob.VideoType not support, oneJob.VideoType = ", oneJob.VideoType)
			done <- nil
		}
	}()

	select {
	case err := <-done:
		// 跳出 select，可以外层继续，不会阻塞在这里
		if err != nil {
			d.log.Errorln(err)
		}
		// 刷新视频的缓存结构
		d.UpdateInfo(oneJob)

		break
	case p := <-panicChan:
		// 遇到内部的 panic，向外抛出
		panic(p)
	case <-d.ctx.Done():
		{
			// 取消这个 context
			d.log.Warningln("cancel Downloader.QueueDownloader()")
			return
		}
	}
}

func (d *Downloader) Cancel() {
	if d == nil {
		return
	}
	d.cancel()
	d.log.Infoln("Downloader.Cancel()")
}

func (d *Downloader) movieDlFunc(ctx context.Context, job taskQueue2.OneJob, downloadIndex int64) error {

	nowSubSupplierHub := d.subSupplierHub
	if nowSubSupplierHub.Suppliers == nil || len(nowSubSupplierHub.Suppliers) < 1 {
		d.log.Infoln("Wait SupplierCheck Update *subSupplierHub, movieDlFunc Skip this time")
		return nil
	}

	// 字幕都下载缓存好了，需要抉择存哪一个，优先选择中文双语的，然后到中文
	organizeSubFiles, err := nowSubSupplierHub.DownloadSub4Movie(job.VideoFPath, downloadIndex)
	if err != nil {
		err = errors.New(fmt.Sprintf("subSupplierHub.DownloadSub4Movie: %v, %v", job.VideoFPath, err))
		d.downloadQueue.AutoDetectUpdateJobStatus(job, err)
		return err
	}
	// 返回的两个值都是 nil 的时候，就是没有下载到字幕
	if organizeSubFiles == nil || len(organizeSubFiles) < 1 {
		d.log.Infoln(task_queue.ErrNoSubFound.Error(), filepath.Base(job.VideoFPath))
		d.downloadQueue.AutoDetectUpdateJobStatus(job, task_queue.ErrNoSubFound)
		return nil
	}

	err = d.oneVideoSelectBestSub(job.VideoFPath, organizeSubFiles)
	if err != nil {
		d.downloadQueue.AutoDetectUpdateJobStatus(job, err)
		return err
	}

	d.downloadQueue.AutoDetectUpdateJobStatus(job, nil)

	// TODO 刷新字幕，这里是 Emby 的，如果是其他的，需要再对接对应的媒体服务器
	if d.settings.EmbySettings.Enable == true && d.embyHelper != nil && job.MediaServerInsideVideoID != "" {

		d.log.Infoln("Refresh Emby Subtitle with id:", job.MediaServerInsideVideoID)
		err = d.embyHelper.EmbyApi.UpdateVideoSubList(job.MediaServerInsideVideoID)
		if err != nil {
			d.log.Errorln("UpdateVideoSubList", job.VideoFPath, job.MediaServerInsideVideoID, "Error:", err)
			return err
		}
	} else {
		if d.settings.EmbySettings.Enable == false {
			d.log.Infoln("UpdateVideoSubList", job.VideoFPath, "Skip, because Emby enable is false")
		} else if d.embyHelper == nil {
			d.log.Infoln("UpdateVideoSubList", job.VideoFPath, "Skip, because EmbyHelper is nil")
		} else if job.MediaServerInsideVideoID == "" {
			d.log.Infoln("UpdateVideoSubList", job.VideoFPath, "Skip, because MediaServerInsideVideoID is empty")
		}
	}

	return nil
}

func (d *Downloader) seriesDlFunc(ctx context.Context, job taskQueue2.OneJob, downloadIndex int64) error {

	nowSubSupplierHub := d.subSupplierHub
	if nowSubSupplierHub == nil || nowSubSupplierHub.Suppliers == nil || len(nowSubSupplierHub.Suppliers) < 1 {
		d.log.Infoln("Wait SupplierCheck Update *subSupplierHub, movieDlFunc Skip this time")
		return nil
	}
	var err error
	// 设置只有一集需要下载
	epsMap := make(map[int][]int, 0)
	epsMap[job.Season] = []int{job.Episode}

	if job.VideoType == common.Series && (job.SeriesRootDirPath == "" || job.Season <= 0 || job.Episode <= 0) {

		// 连续剧的时候需要额外提交信息
		epsVideoNfoInfo, err := decode.GetVideoNfoInfo4OneSeriesEpisode(job.VideoFPath)
		if err != nil {
			return err
		}
		seriesInfoDirPath := decode.GetSeriesDirRootFPath(job.VideoFPath)
		if seriesInfoDirPath == "" {
			err = errors.New(fmt.Sprintf("decode.GetSeriesDirRootFPath == Empty, %s", job.VideoFPath))
			return err
		}

		job.Season = epsVideoNfoInfo.Season
		job.Episode = epsVideoNfoInfo.Episode
		job.SeriesRootDirPath = seriesInfoDirPath
		// 如果进来的时候是 Season 和 Eps 是 -1， 那么就需要重新赋值那些集是需要下载的，否则后面会跳过
		epsMap = make(map[int][]int, 0)
		epsMap[job.Season] = []int{job.Episode}
	}

	// 这里拿到了这一部连续剧的所有的剧集信息，以及所有下载到的字幕信息
	seriesInfo, err := series_helper.ReadSeriesInfoFromDir(
		d.log, job.SeriesRootDirPath,
		d.settings.AdvancedSettings.TaskQueue.ExpirationTime,
		false,
		false,
		epsMap)
	if err != nil {
		err = errors.New(fmt.Sprintf("seriesDlFunc.ReadSeriesInfoFromDir, Error: %v", err))
		d.downloadQueue.AutoDetectUpdateJobStatus(job, err)
		return err
	}
	// 下载好的字幕文件
	var organizeSubFiles map[string][]string
	// 下载的接口是统一的
	organizeSubFiles, err = nowSubSupplierHub.DownloadSub4Series(job.SeriesRootDirPath,
		seriesInfo,
		downloadIndex)
	if err != nil {
		err = errors.New(fmt.Sprintf("seriesDlFunc.DownloadSub4Series %v S%vE%v %v", filepath.Base(job.SeriesRootDirPath), job.Season, job.Episode, err))
		d.downloadQueue.AutoDetectUpdateJobStatus(job, err)
		return err
	}
	// 是否下载到字幕了
	if organizeSubFiles == nil || len(organizeSubFiles) < 1 {
		d.log.Infoln(task_queue.ErrNoSubFound.Error(), filepath.Base(job.VideoFPath), job.Season, job.Episode)
		d.downloadQueue.AutoDetectUpdateJobStatus(job, task_queue.ErrNoSubFound)
		return nil
	}

	var errSave2Local error
	save2LocalSubCount := 0
	// 只针对需要下载字幕的视频进行字幕的选择保存
	subVideoCount := 0
	for epsKey, episodeInfo := range seriesInfo.NeedDlEpsKeyList {

		// 创建一个 chan 用于任务的中断和超时
		done := make(chan interface{}, 1)
		// 接收内部任务的 panic
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}

				close(done)
				close(panicChan)
			}()
			// 匹配对应的 Eps 去处理
			done <- d.oneVideoSelectBestSub(episodeInfo.FileFullPath, organizeSubFiles[epsKey])
		}()

		select {
		case errInterface := <-done:
			if errInterface != nil {
				errSave2Local = errInterface.(error)
				d.log.Errorln(errInterface.(error))
			} else {
				save2LocalSubCount++
			}
			break
		case p := <-panicChan:
			// 遇到内部的 panic，向外抛出
			d.log.Errorln("seriesDlFunc.oneVideoSelectBestSub panicChan", p)
			break
		case <-ctx.Done():
			{
				err = errors.New(fmt.Sprintf("cancel at NeedDlEpsKeyList.oneVideoSelectBestSub, %v S%dE%d", seriesInfo.Name, episodeInfo.Season, episodeInfo.Episode))
				d.downloadQueue.AutoDetectUpdateJobStatus(job, err)
				return err
			}
		}

		subVideoCount++
	}
	// 这里会拿到一份季度字幕的列表比如，Key 是 S1E0 S2E0 S3E0，value 是新的存储位置
	fullSeasonSubDict := d.saveFullSeasonSub(seriesInfo, organizeSubFiles)
	// TODO 季度的字幕包，应该优先于零散的字幕吧，暂定就这样了，注意是全部都替换
	// 需要与有下载需求的季交叉
	for _, episodeInfo := range seriesInfo.EpList {

		_, ok := seriesInfo.NeedDlSeasonDict[episodeInfo.Season]
		if ok == false {
			continue
		}

		// 创建一个 chan 用于任务的中断和超时
		done := make(chan interface{}, 1)
		// 接收内部任务的 panic
		panicChan := make(chan interface{}, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
				close(done)
				close(panicChan)
			}()
			// 匹配对应的 Eps 去处理
			seasonEpsKey := my_util.GetEpisodeKeyName(episodeInfo.Season, episodeInfo.Episode)
			if fullSeasonSubDict[seasonEpsKey] == nil || len(fullSeasonSubDict[seasonEpsKey]) < 1 {
				d.log.Infoln("seriesDlFunc.saveFullSeasonSub, no sub found, Skip", seasonEpsKey)
				done <- nil
			}

			done <- d.oneVideoSelectBestSub(episodeInfo.FileFullPath, fullSeasonSubDict[seasonEpsKey])
		}()

		select {
		case errInterface := <-done:
			if errInterface != nil {
				errSave2Local = errInterface.(error)
				d.log.Errorln(errInterface.(error))
			} else {
				save2LocalSubCount++
			}

			break
		case p := <-panicChan:
			// 遇到内部的 panic，向外抛出
			d.log.Errorln("seriesDlFunc.oneVideoSelectBestSub panicChan", p)
			break
		case <-ctx.Done():
			{
				err = errors.New(fmt.Sprintf("cancel at NeedDlEpsKeyList.oneVideoSelectBestSub, %v S%dE%d", seriesInfo.Name, episodeInfo.Season, episodeInfo.Episode))
				d.downloadQueue.AutoDetectUpdateJobStatus(job, err)
				return err
			}
		}
	}
	// 是否清理全季的缓存字幕文件夹
	if d.settings.AdvancedSettings.SaveFullSeasonTmpSubtitles == false {
		err = sub_helper.DeleteOneSeasonSubCacheFolder(seriesInfo.DirPath)
		if err != nil {
			d.log.Errorln("seriesDlFunc.DeleteOneSeasonSubCacheFolder", err)
		}
	}

	if save2LocalSubCount < 1 {
		// 下载的字幕都没有一个能够写入到本地的，那么就有问题了
		d.downloadQueue.AutoDetectUpdateJobStatus(job, errSave2Local)
		return errSave2Local
	}
	// 哪怕有一个写入到本地成功了，也无需对本次任务报错
	d.downloadQueue.AutoDetectUpdateJobStatus(job, nil)
	// TODO 刷新字幕，这里是 Emby 的，如果是其他的，需要再对接对应的媒体服务器
	if d.settings.EmbySettings.Enable == true && d.embyHelper != nil && job.MediaServerInsideVideoID != "" {
		err = d.embyHelper.EmbyApi.UpdateVideoSubList(job.MediaServerInsideVideoID)
		if err != nil {
			d.log.Errorln("UpdateVideoSubList", job.SeriesRootDirPath, job.MediaServerInsideVideoID, job.Season, job.Episode, "Error:", err)
			return err
		}
	}

	return nil
}
