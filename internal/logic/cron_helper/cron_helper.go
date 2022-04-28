package cron_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/scan_played_video_subinfo"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/task_queue"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/video_scan_and_refresh_helper"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type CronHelper struct {
	stopping                      bool // 正在停止
	cronHelperRunning             bool // 这个是定时器启动的状态，它为true，不代表核心函数在执行
	scanPlayedVideoSubInfo        *scan_played_video_subinfo.ScanPlayedVideoSubInfo
	fileDownloader                *file_downloader.FileDownloader
	downloadQueue                 *task_queue.TaskQueue  // 需要下载的视频的队列
	downloader                    *downloader.Downloader // 下载者线程
	cronLock                      sync.Mutex             // 锁
	c                             *cron.Cron             // 定时器实例
	settings                      *settings.Settings     // 设置实例
	log                           *logrus.Logger         // 日志实例
	entryIDScanVideoProcess       cron.EntryID
	entryIDSupplierCheck          cron.EntryID
	entryIDQueueDownloader        cron.EntryID
	entryIDScanPlayedVideoSubInfo cron.EntryID
}

func NewCronHelper(fileDownloader *file_downloader.FileDownloader) *CronHelper {

	ch := CronHelper{
		fileDownloader: fileDownloader,
		log:            fileDownloader.Log,
		settings:       fileDownloader.Settings,
		// 实例化下载队列
		downloadQueue: task_queue.NewTaskQueue("LocalSubDownloadQueue", fileDownloader.Settings, fileDownloader.Log),
	}

	var err error
	ch.scanPlayedVideoSubInfo, err = scan_played_video_subinfo.NewScanPlayedVideoSubInfo(ch.log, ch.settings)
	if err != nil {
		ch.log.Panicln(err)
	}
	return &ch
}

// Start 开启定时器任务，这个任务是非阻塞的，scanVideoProcessAdd2DownloadQueue 仅仅可能是这个函数执行耗时而已
// runImmediately == false 那么 ch.c.Start() 是不会阻塞的
func (ch *CronHelper) Start(runImmediately bool) {

	ch.cronLock.Lock()
	if ch.cronHelperRunning == true {
		ch.cronLock.Unlock()
		return
	}
	ch.cronLock.Unlock()

	ch.cronLock.Lock()
	ch.cronHelperRunning = true
	ch.stopping = false
	ch.cronLock.Unlock()
	// ----------------------------------------------
	// 初始化下载者，里面的两个 func 需要使用定时器启动 SupplierCheck QueueDownloader
	ch.downloader = downloader.NewDownloader(
		sub_formatter.GetSubFormatter(ch.log, ch.settings.AdvancedSettings.SubNameFormatter),
		ch.fileDownloader, ch.downloadQueue)
	// ----------------------------------------------
	// 判断扫描任务的时间间隔是否符合要求，不符合则重写默认值
	_, err := cron.ParseStandard(ch.settings.CommonSettings.ScanInterval)
	if err != nil {
		ch.log.Warningln("CommonSettings.ScanInterval format error, after v0.25.x , need reset this at WebUI")
		// 如果解析错误了，就需要重新赋值默认值过来，然后保存
		nowSettings := ch.settings
		nowSettings.CommonSettings.ScanInterval = settings.NewCommonSettings().ScanInterval
		err = settings.SetFullNewSettings(nowSettings)
		if err != nil {
			ch.log.Panicln("CronHelper.SetFullNewSettings:", err)
			return
		}
	}
	// ----------------------------------------------
	ch.c = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	// 定时器
	// 这个暂时无法被取消执行
	ch.entryIDScanVideoProcess, err = ch.c.AddFunc(ch.settings.CommonSettings.ScanInterval, ch.scanVideoProcessAdd2DownloadQueue)
	if err != nil {
		ch.log.Panicln("CronHelper scanVideoProcessAdd2DownloadQueue, Cron entryID:", ch.entryIDScanVideoProcess, "Error:", err)
	}
	// 这个可以由 ch.downloader.Cancel() 取消执行
	ch.entryIDSupplierCheck, err = ch.c.AddFunc("@every 1h", ch.downloader.SupplierCheck)
	if err != nil {
		ch.log.Panicln("CronHelper SupplierCheck, Cron entryID:", ch.entryIDSupplierCheck, "Error:", err)
	}
	// 这个可以由 ch.downloader.Cancel() 取消执行
	ch.entryIDQueueDownloader, err = ch.c.AddFunc("@every 15s", ch.downloader.QueueDownloader)
	if err != nil {
		ch.log.Panicln("CronHelper QueueDownloader, Cron entryID:", ch.entryIDQueueDownloader, "Error:", err)
	}
	// 这个可以由 ch.scanPlayedVideoSubInfo.Cancel() 取消执行
	ch.entryIDScanPlayedVideoSubInfo, err = ch.c.AddFunc("@every 24h", ch.scanPlayedVideoSub)
	if err != nil {
		ch.log.Panicln("CronHelper QueueDownloader, Cron entryID:", ch.entryIDScanPlayedVideoSubInfo, "Error:", err)
	}

	// 是否在定时器开启前先执行一次任务
	if runImmediately == true {

		ch.log.Infoln("First Time scanVideoProcessAdd2DownloadQueue Start")

		//if ch.settings.SpeedDevMode == false {
		ch.scanVideoProcessAdd2DownloadQueue()
		//}

		ch.downloader.SupplierCheck()

		ch.log.Infoln("First Time scanVideoProcessAdd2DownloadQueue End")

	} else {
		ch.log.Infoln("RunAtStartup: false, so will not Run At Startup")
	}

	ch.log.Infoln("CronHelper Start...")
	ch.c.Start()

	// 只有定时任务 start 之后才能拿到信息
	if len(ch.c.Entries()) > 0 {

		// 不会马上启动扫描，那么就需要设置当前的时间，且为 waiting
		tttt := ch.c.Entry(ch.entryIDScanVideoProcess).Next.Format("2006-01-02 15:04:05")
		common.SetSubScanJobStatusWaiting(tttt)

		ch.log.Infoln("Next Sub Scan Will Process At:", tttt)
	} else {
		ch.log.Errorln("Can't get cron jobs, will not send SubScanJobStatus")
	}
}

// Stop 会阻塞等待任务完成
func (ch *CronHelper) Stop() {

	cronHelperRunning := false
	ch.cronLock.Lock()
	cronHelperRunning = ch.cronHelperRunning
	ch.cronLock.Unlock()

	if cronHelperRunning == false {
		return
	}

	ch.cronLock.Lock()
	if ch.stopping == true {
		ch.cronLock.Unlock()
		return
	}
	ch.stopping = true
	ch.cronLock.Unlock()

	ch.downloader.Cancel()
	ch.scanPlayedVideoSubInfo.Cancel()
	// Stop stops the cron scheduler if it is running; otherwise it does nothing.
	// A context is returned so the caller can wait for running jobs to complete.
	nowContext := ch.c.Stop()
	select {
	case <-time.After(5 * time.Minute):
		ch.log.Warningln("Wait over 5 min, CronHelper is timeout")
	case <-nowContext.Done():
		ch.log.Infoln("CronHelper.Stop() Done.")
	}

	ch.cronLock.Lock()
	ch.cronHelperRunning = false
	ch.stopping = false
	ch.cronLock.Unlock()

	common.SetSubScanJobStatusNil()
}

func (ch *CronHelper) scanPlayedVideoSub() {

	bok, err := ch.scanPlayedVideoSubInfo.GetPlayedItemsSubtitle()
	if err != nil {
		ch.log.Errorln(err)
	}
	if bok == true {
		ch.scanPlayedVideoSubInfo.Clear()
		err = ch.scanPlayedVideoSubInfo.Scan()
		if err != nil {
			ch.log.Errorln(err)
		}
	}
}

func (ch *CronHelper) CronHelperRunning() bool {

	defer func() {
		ch.cronLock.Unlock()
	}()
	ch.cronLock.Lock()
	return ch.cronHelperRunning
}

func (ch *CronHelper) CronHelperStopping() bool {

	defer func() {
		ch.cronLock.Unlock()
	}()
	ch.cronLock.Lock()
	return ch.stopping
}

func (ch *CronHelper) CronRunningStatusString() string {

	if ch.CronHelperRunning() == true {
		if ch.CronHelperStopping() == true {
			return Stopping
		}
		return Running
	} else {
		return Stopped
	}
}

// scanVideoProcessAdd2DownloadQueue 定时执行的视频扫描任务，提交给任务队列，然后由额外的下载者线程去取队列中的任务下载
func (ch *CronHelper) scanVideoProcessAdd2DownloadQueue() {

	defer func() {
		ch.cronLock.Lock()
		ch.cronLock.Unlock()

		// 下载完后，应该继续是等待
		tttt := ch.c.Entry(ch.entryIDScanVideoProcess).Next.Format("2006-01-02 15:04:05")
		common.SetSubScanJobStatusWaiting(tttt)
	}()

	// 扫描字幕任务开始，先是扫描阶段，那么是拿不到有多少视频需要扫描的数量的
	common.SetSubScanJobStatusPreparing(time.Now().Format("2006-01-02 15:04:05"))
	// ----------------------------------------------------------------------------------------
	// ----------------------------------------------------------------------------------------
	// 扫描有那些视频需要下载字幕，放入队列中，然后会有下载者去这个队列取出来进行下载
	videoScanAndRefreshHelper := video_scan_and_refresh_helper.NewVideoScanAndRefreshHelper(
		ch.fileDownloader,
		ch.downloadQueue)

	ch.log.Infoln("Video Scan Started...")
	// 先进行扫描
	scanResult, err := videoScanAndRefreshHelper.ScanNormalMovieAndSeries()
	if err != nil {
		ch.log.Errorln("ScanNormalMovieAndSeries", err)
		return
	}
	err = videoScanAndRefreshHelper.ScanEmbyMovieAndSeries(scanResult)
	if err != nil {
		ch.log.Errorln("ScanEmbyMovieAndSeries", err)
		return
	}
	// 过滤出需要下载的视频有那些，并放入队列中
	err = videoScanAndRefreshHelper.FilterMovieAndSeriesNeedDownload(scanResult)
	if err != nil {
		ch.log.Errorln("FilterMovieAndSeriesNeedDownload", err)
		return
	}
}

const (
	Stopped  = "stopped"
	Running  = "running"
	Stopping = "stopping"
)
