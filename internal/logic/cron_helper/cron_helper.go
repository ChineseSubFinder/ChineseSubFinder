package cron_helper

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/mix_media_info"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/jinzhu/now"
	"strconv"
	"sync"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/scan_played_video_subinfo"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/task_queue"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/video_scan_and_refresh_helper"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type CronHelper struct {
	stopping                      bool                                                     // 正在停止
	cronHelperRunning             bool                                                     // 这个是定时器启动的状态，它为true，不代表核心函数在执行
	scanPlayedVideoSubInfo        *scan_played_video_subinfo.ScanPlayedVideoSubInfo        // 扫描已经播放过的视频的字幕信息
	FileDownloader                *file_downloader.FileDownloader                          // 文件下载器
	DownloadQueue                 *task_queue.TaskQueue                                    // 需要下载的视频的队列
	Downloader                    *downloader.Downloader                                   // 下载者线程
	videoScanAndRefreshHelper     *video_scan_and_refresh_helper.VideoScanAndRefreshHelper // 视频扫描和刷新的帮助类
	cronLock                      sync.Mutex                                               // 锁
	c                             *cron.Cron                                               // 定时器实例
	Settings                      *settings.Settings                                       // 设置实例
	log                           *logrus.Logger                                           // 日志实例
	entryIDScanVideoProcess       cron.EntryID
	entryIDSupplierCheck          cron.EntryID
	entryIDQueueDownloader        cron.EntryID
	entryIDScanPlayedVideoSubInfo cron.EntryID
	entryIDUploadPlayedVideoSub   cron.EntryID
}

func NewCronHelper(fileDownloader *file_downloader.FileDownloader) *CronHelper {

	ch := CronHelper{
		FileDownloader: fileDownloader,
		log:            fileDownloader.Log,
		Settings:       fileDownloader.Settings,
		// 实例化下载队列
		DownloadQueue: task_queue.NewTaskQueue(fileDownloader.CacheCenter),
	}

	var err error
	// ----------------------------------------------
	// 扫描已播放
	ch.scanPlayedVideoSubInfo, err = scan_played_video_subinfo.NewScanPlayedVideoSubInfo(ch.log, ch.Settings, fileDownloader)
	if err != nil {
		ch.log.Panicln(err)
	}
	// ----------------------------------------------
	// 字幕扫描器
	ch.videoScanAndRefreshHelper = video_scan_and_refresh_helper.NewVideoScanAndRefreshHelper(
		ch.FileDownloader,
		ch.DownloadQueue)

	// ----------------------------------------------
	// 初始化下载者，里面的两个 func 需要使用定时器启动 SupplierCheck QueueDownloader
	ch.Downloader = downloader.NewDownloader(
		sub_formatter.GetSubFormatter(ch.log, ch.Settings.AdvancedSettings.SubNameFormatter),
		ch.FileDownloader, ch.DownloadQueue)

	// 强制进行一次字幕源有效性检查
	ch.Downloader.SupplierCheck()

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
	// 判断扫描任务的时间间隔是否符合要求，不符合则重写默认值
	_, err := cron.ParseStandard(ch.Settings.CommonSettings.ScanInterval)
	if err != nil {
		ch.log.Warningln("CommonSettings.ScanInterval format error, after v0.25.x , need reset this at WebUI")
		// 如果解析错误了，就需要重新赋值默认值过来，然后保存
		nowSettings := ch.Settings
		nowSettings.CommonSettings.ScanInterval = settings.NewCommonSettings().ScanInterval
		err = settings.SetFullNewSettings(nowSettings)
		if err != nil {
			ch.log.Panicln("CronHelper.SetFullNewSettings:", err)
			return
		}
	}
	// ----------------------------------------------
	ch.c = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	{
		// 测试部分定时器代码，提前运行
		if ch.Settings.SpeedDevMode == true {

			//ch.scanPlayedVideoSub()
			ch.uploadPlayedVideoSub()
		}
	}

	// 定时器
	// 这个暂时无法被取消执行
	ch.entryIDScanVideoProcess, err = ch.c.AddFunc(ch.Settings.CommonSettings.ScanInterval, ch.scanVideoProcessAdd2DownloadQueue)
	if err != nil {
		ch.log.Panicln("CronHelper scanVideoProcessAdd2DownloadQueue, scanVideoProcessAdd2DownloadQueue Cron entryID:", ch.entryIDScanVideoProcess, "Error:", err)
	}
	// 这个可以由 ch.Downloader.Cancel() 取消执行
	ch.entryIDSupplierCheck, err = ch.c.AddFunc("@every 1h", ch.Downloader.SupplierCheck)
	if err != nil {
		ch.log.Panicln("CronHelper SupplierCheck, SupplierCheck Cron entryID:", ch.entryIDSupplierCheck, "Error:", err)
	}
	// 这个可以由 ch.Downloader.Cancel() 取消执行
	ch.entryIDQueueDownloader, err = ch.c.AddFunc("@every 15s", ch.Downloader.QueueDownloader)
	if err != nil {
		ch.log.Panicln("CronHelper QueueDownloader, QueueDownloader Cron entryID:", ch.entryIDQueueDownloader, "Error:", err)
	}
	// 这个可以由 ch.scanPlayedVideoSubInfo.Cancel() 取消执行
	ch.entryIDScanPlayedVideoSubInfo, err = ch.c.AddFunc("@every 24h", ch.scanPlayedVideoSub)
	if err != nil {
		ch.log.Panicln("CronHelper QueueDownloader, scanPlayedVideoSub Cron entryID:", ch.entryIDScanPlayedVideoSubInfo, "Error:", err)
	}
	// 字幕的上传逻辑
	if ch.Settings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled == true {

		intervalNowTask := "@every 1m"
		if ch.Settings.SpeedDevMode == true {
			intervalNowTask = "@every 10s"
		}
		ch.entryIDUploadPlayedVideoSub, err = ch.c.AddFunc(intervalNowTask, ch.uploadPlayedVideoSub)
		if err != nil {
			ch.log.Panicln("CronHelper QueueDownloader, uploadPlayedVideoSub Cron entryID:", ch.entryIDUploadPlayedVideoSub, "Error:", err)
		}
	}

	// ----------------------------------------------
	if runImmediately == true {
		// 是否在定时器开启前先执行一次视频扫描任务
		ch.cronLock.Lock()
		if ch.cronHelperRunning == true && ch.stopping == false {
			ch.cronLock.Unlock()
			//----------------------------------------------
			// 没有停止，那么继续扫描
			ch.log.Infoln("First Time scanVideoProcessAdd2DownloadQueue Start")
			if ch.Settings.SpeedDevMode == false {
				ch.scanVideoProcessAdd2DownloadQueue()
			}
			ch.log.Infoln("First Time scanVideoProcessAdd2DownloadQueue End")
			//----------------------------------------------
		} else {
			ch.cronLock.Unlock()
			ch.log.Infoln("CronHelper is stopping, not start scanVideoProcessAdd2DownloadQueue")
			return
		}
	} else {
		ch.log.Infoln("RunAtStartup: false, so will not Run At Startup")
	}
	// ----------------------------------------------
	// 如果不是立即执行，那么就等待定时器开启
	ch.cronLock.Lock()
	if ch.cronHelperRunning == true && ch.stopping == false {
		ch.cronLock.Unlock()
		//----------------------------------------------
		ch.log.Infoln("CronHelper Start...")
		ch.c.Start()
		//----------------------------------------------
		// 只有定时任务 start 之后才能拿到信息
		if len(ch.c.Entries()) > 0 {
			// 不会马上启动扫描，那么就需要设置当前的时间，且为 waiting
			tttt := ch.c.Entry(ch.entryIDScanVideoProcess).Next.Format("2006-01-02 15:04:05")
			ch.log.Infoln("Next Sub Scan Will Process At:", tttt)
		} else {
			ch.log.Errorln("Can't get cron jobs, will not send SubScanJobStatus")
		}
		//----------------------------------------------
	} else {
		ch.cronLock.Unlock()
		ch.log.Infoln("CronHelper is stopping, not start CronHelper")
	}
	//----------------------------------------------

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

	ch.videoScanAndRefreshHelper.Cancel()
	ch.Downloader.Cancel()
	ch.scanPlayedVideoSubInfo.Cancel()
	// Stop stops the cron scheduler if it is running; otherwise it does nothing.
	// A context is returned so the caller can wait for running jobs to complete.
	nowContext := ch.c.Stop()
	select {
	case <-time.After(5 * time.Minute):
		ch.log.Warningln("Wait over 5 min, CronHelper is timeout")
	case <-nowContext.Done():
		ch.log.Infoln("CronHelper.Stop() context<-Done.")
	}

	ch.cronLock.Lock()
	ch.cronHelperRunning = false
	ch.stopping = false
	ch.cronLock.Unlock()

	ch.log.Infoln("CronHelper.Stop() Done.")
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

// uploadPlayedVideoSub  上传字幕的定时器
func (ch *CronHelper) uploadPlayedVideoSub() {

	// 找出没有上传过的字幕列表
	var notUploadedVideoSubInfos []models.VideoSubInfo
	dao.GetDb().Where("is_send = ?", false).Limit(1).Find(&notUploadedVideoSubInfos)

	if len(notUploadedVideoSubInfos) < 1 {
		ch.log.Debugln("No notUploadedVideoSubInfos")
		return
	}

	var imdbInfos []models.IMDBInfo
	dao.GetDb().Where("imdb_id = ?", notUploadedVideoSubInfos[0].IMDBInfoID).Find(&imdbInfos)
	if len(imdbInfos) < 1 {
		// 如果没有找到，那么就没有办法推断出 IMDB ID 的相关信息和 TMDB ID 信息，要来何用，删除即可
		ch.log.Infoln("No imdbInfos, will delete this VideoSubInfo,", notUploadedVideoSubInfos[0].SubName)
		dao.GetDb().Delete(&notUploadedVideoSubInfos[0])
		return
	}
	videoType := ""
	if imdbInfos[0].IsMovie == true {
		videoType = "movie"
	} else {
		videoType = "series"
	}
	var err error
	var finalQueryIMDBInfo *models.MediaInfo
	if imdbInfos[0].TmdbId == "" {

		// 需要先对这个字幕的 IMDB ID 转 TMDB ID 信息进行查询，得到 TMDB ID 和 Year (2019 2022)
		finalQueryIMDBInfo, err = mix_media_info.GetMediaInfoAndSave(ch.log, ch.FileDownloader.SubtitleBestApi, &imdbInfos[0], imdbInfos[0].IMDBID, "imdb", videoType)
		if err != nil {
			ch.log.Errorln(errors.New("GetMediaInfoAndSave error:" + err.Error()))
			return
		}
	} else {

		var mediaInfos []models.MediaInfo
		dao.GetDb().Where("tmdb_id = ?", imdbInfos[0].TmdbId).Find(&mediaInfos)
		if len(mediaInfos) < 1 {
			finalQueryIMDBInfo, err = mix_media_info.GetMediaInfoAndSave(ch.log, ch.FileDownloader.SubtitleBestApi, &imdbInfos[0], imdbInfos[0].IMDBID, "imdb", videoType)
			if err != nil {
				ch.log.Errorln(errors.New("GetMediaInfoAndSave error:" + err.Error()))
				return
			}
		} else {
			finalQueryIMDBInfo = &mediaInfos[0]
		}
	}
	// 问询这个字幕是否上传过了，如果没有就需要进入上传的队列
	askForUploadReply, err := ch.FileDownloader.SubtitleBestApi.AskFroUpload(notUploadedVideoSubInfos[0].SHA256)
	if err != nil {
		ch.log.Errorln(fmt.Errorf("AskFroUpload err: %v", err))
		return
	}
	if askForUploadReply.Status == 3 {
		// 上传过了，直接标记本地的 is_send 字段为 true
		notUploadedVideoSubInfos[0].IsSend = true
		dao.GetDb().Save(&notUploadedVideoSubInfos[0])
		ch.log.Infoln("Subtitle has been uploaded, so will not upload again")
		return
	} else if askForUploadReply.Status == 4 {
		// 上传队列满了，等待下次定时器触发
		ch.log.Infoln("Subtitle upload queue is full, will try ask upload again")
		return
	} else if askForUploadReply.Status == 2 {
		// 这个上传任务已经在队列中了，也许有其他人也需要上传这个字幕，或者本机排队的时候故障了，重启也可能遇到这个故障
		ch.log.Infoln("Subtitle is int the queue")
		return
	} else if askForUploadReply.Status == 1 {
		// 正确放入了队列，然后需要按规划的时间进行上传操作
		// 这里可能需要执行耗时操作来等待到安排的时间点进行字幕的上传，不能直接长时间的 Sleep 操作
		// 每次 Sleep 1s 然后就判断一次定时器是否还允许允许，如果不运行了，那么也就需要退出循环

		// 得到目标时间与当前时间的差值，单位是s
		waitTime := askForUploadReply.ScheduledUnixTime - time.Now().Unix()
		if waitTime <= 0 {
			waitTime = 5
		}
		ch.log.Infoln("will wait", waitTime, "s 2 upload sub 2 server")
		var sleepCounter int64
		sleepCounter = 0
		normalStatus := false
		for ch.cronHelperRunning == true {
			if sleepCounter > waitTime {
				normalStatus = true
				break
			}
			if sleepCounter%30 == 0 {
				ch.log.Infoln("wait 2 upload sub")
			}
			time.Sleep(1 * time.Second)
			sleepCounter++
		}
		if normalStatus == false || ch.cronHelperRunning == false {
			// 说明不是正常跳出来的，是结束定时器来执行的
			ch.log.Infoln("uploadPlayedVideoSub early termination")
			return
		}
		// 发送字幕
		shareRootDir, err := my_folder.GetShareSubRootFolder()
		if err != nil {
			ch.log.Errorln("GetShareSubRootFolder error:", err.Error())
			return
		}

		releaseTime, err := now.Parse(finalQueryIMDBInfo.Year)
		if err != nil {
			ch.log.Errorln("now.Parse error:", err.Error())
			return
		}

		uploadSubReply, err := ch.FileDownloader.SubtitleBestApi.UploadSub(&notUploadedVideoSubInfos[0], shareRootDir, finalQueryIMDBInfo.TmdbId, strconv.Itoa(releaseTime.Year()), ch.Settings.AdvancedSettings.ProxySettings)
		if err != nil {
			ch.log.Errorln("UploadSub error:", err.Error())
			return
		}
		if uploadSubReply.Status == 1 {
			// 成功，其他情况就等待 Ask for Upload
			notUploadedVideoSubInfos[0].IsSend = true
			dao.GetDb().Save(&notUploadedVideoSubInfos[0])
			ch.log.Infoln("Add subtitle in upload queue")
			return
		} else {
			ch.log.Warningln("UploadSub Message:", uploadSubReply.Message)
			return
		}

	} else {
		// 不是预期的返回值，需要报警
		ch.log.Errorln(fmt.Errorf("AskFroUpload Not the expected return value, Status: %d, Message: %v", askForUploadReply.Status, askForUploadReply.Message))
		return
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

	// 扫描字幕任务开始，先是扫描阶段，那么是拿不到有多少视频需要扫描的数量的
	ch.log.Infoln("scanVideoProcessAdd2DownloadQueue Start:", time.Now().Format("2006-01-02 15:04:05"))
	// ----------------------------------------------------------------------------------------
	// ----------------------------------------------------------------------------------------
	// 扫描有那些视频需要下载字幕，放入队列中，然后会有下载者去这个队列取出来进行下载
	err := ch.videoScanAndRefreshHelper.Start()
	if err != nil {
		ch.log.Errorln(err)
		return
	}
}

const (
	Stopped  = "stopped"
	Running  = "running"
	Stopping = "stopping"
)
