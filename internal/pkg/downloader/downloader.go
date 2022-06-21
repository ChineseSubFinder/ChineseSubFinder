package downloader

import (
	"errors"
	"fmt"
	"sync"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/csf"

	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	markSystem "github.com/allanpk716/ChineseSubFinder/internal/logic/mark_system"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/pre_download_process"
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	subCommon "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
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

	d.log.Debugln("Download.QueueDownloader() Try Start ...")
	d.downloaderLock.Lock()
	d.log.Debugln("Download.QueueDownloader() Start ...")

	defer func() {
		if p := recover(); p != nil {
			d.log.Errorln("Downloader.QueueDownloader() panic")
			my_util.PrintPanicStack(d.log)
		}
		d.downloaderLock.Unlock()
		d.log.Debugln("Download.QueueDownloader() End")
	}()

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
	// 这个任务如果是 series 那么需要考虑是否原始存入的信息是缺失的，需要补全
	{
		if oneJob.VideoType == common.Series && (oneJob.SeriesRootDirPath == "" || oneJob.Season <= 0 || oneJob.Episode <= 0) {
			// 连续剧的时候需要额外提交信息
			epsVideoNfoInfo, err := decode.GetVideoNfoInfo4OneSeriesEpisode(oneJob.VideoFPath)
			if err != nil {
				d.log.Errorln("decode.GetVideoNfoInfo4OneSeriesEpisode()", err)
				return
			}
			seriesInfoDirPath := decode.GetSeriesDirRootFPath(oneJob.VideoFPath)
			if seriesInfoDirPath == "" {
				d.log.Errorln(fmt.Sprintf("decode.GetSeriesDirRootFPath == Empty, %s", oneJob.VideoFPath))
				return
			}
			oneJob.Season = epsVideoNfoInfo.Season
			oneJob.Episode = epsVideoNfoInfo.Episode
			oneJob.SeriesRootDirPath = seriesInfoDirPath
		}
	}
	// --------------------------------------------------
	// 判断是否看过，这个只有 Emby 情况下才会生效
	{
		isPlayed := false
		if d.embyHelper != nil {
			// 在拿出来后，如果是有内部媒体服务器媒体 ID 的，那么就去查询是否已经观看过了
			isPlayed, err = d.embyHelper.IsVideoPlayed(oneJob.MediaServerInsideVideoID)
			if err != nil {
				d.log.Errorln("d.embyHelper.IsVideoPlayed()", oneJob.VideoFPath, err)
				return
			}
		}
		// TODO 暂时屏蔽掉 http api 提交的已看字幕的接口上传
		// 不管如何，只要是发现数据库中有 HTTP API 提交的信息，就认为是看过
		//var videoPlayedInfos []models.ThirdPartSetVideoPlayedInfo
		//dao.GetDb().Where("physical_video_file_full_path = ?", oneJob.VideoFPath).Find(&videoPlayedInfos)
		//if len(videoPlayedInfos) > 0 {
		//	isPlayed = true
		//}
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
	}
	// --------------------------------------------------
	// 判断是否需要跳过，因为如果是 Normal 扫描出来的，那么可能因为视频时间久远，下载一次即可
	{
		if oneJob.TaskPriority > task_queue.HighTaskPriorityLevel {
			// 优先级大于 3，那么就不是很急的任务，才需要判断
			if oneJob.VideoType == common.Movie {
				if d.subSupplierHub.MovieNeedDlSub(oneJob.VideoFPath, false) == false {
					// 需要标记忽略
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
					d.log.Infoln("MovieNeedDlSub == false, Ignore This Job")
					return
				}
			} else if oneJob.VideoType == common.Series {

				bNeedDlSub, seriesInfo, err := d.subSupplierHub.SeriesNeedDlSub(oneJob.SeriesRootDirPath,
					false, false)
				if err != nil {
					d.log.Errorln("SeriesNeedDlSub", err)
					return
				}
				// 需要下载的 Eps 是否与 Normal 判断这个连续剧中有那些剧集需要下载的，情况符合。通过下载的时间来判断
				epsKey := my_util.GetEpisodeKeyName(oneJob.Season, oneJob.Episode)
				_, found := seriesInfo.NeedDlEpsKeyList[epsKey]
				if bNeedDlSub == false || found == false {
					// 需要标记忽略
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
					d.log.Infoln("SeriesNeedDlSub == false, Ignore This Job")
					return
				}
			}
		}
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

			if global_value.LiteMode() == false {
				my_util.CloseChrome(d.log)
			}
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
