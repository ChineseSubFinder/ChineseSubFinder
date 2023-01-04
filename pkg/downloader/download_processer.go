package downloader

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	taskQueue2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/series_helper"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/task_queue"
	"golang.org/x/net/context"
)

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
	if settings.Get().EmbySettings.Enable == true && d.embyHelper != nil && job.MediaServerInsideVideoID != "" {

		d.log.Infoln("字幕下载完毕，尝试刷新 Emby 中对应字幕", job.VideoFPath, job.MediaServerInsideVideoID)
		err = d.embyHelper.EmbyApi.UpdateVideoSubList(settings.Get().EmbySettings, job.MediaServerInsideVideoID)
		if err != nil {
			d.log.Errorln("UpdateVideoSubList", job.VideoFPath, job.MediaServerInsideVideoID, "Error:", err)
			return err
		}
	} else {
		if settings.Get().EmbySettings.Enable == false {
			d.log.Infoln("字幕下载完毕，尝试刷新 Emby 中对应字幕", job.VideoFPath, "Skip, because Emby enable is false")
		} else if d.embyHelper == nil {
			d.log.Infoln("字幕下载完毕，尝试刷新 Emby 中对应字幕", job.VideoFPath, "Skip, because EmbyHelper is nil")
		} else if job.MediaServerInsideVideoID == "" {
			d.log.Infoln("字幕下载完毕，尝试刷新 Emby 中对应字幕", job.VideoFPath, "Skip, because MediaServerInsideVideoID is empty")
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
	// 这里拿到了这一部连续剧的所有的剧集信息，以及所有下载到的字幕信息
	seriesInfo, err := series_helper.ReadSeriesInfoFromDir(
		d.fileDownloader.MediaInfoDealers, job.SeriesRootDirPath,
		settings.Get().AdvancedSettings.TaskQueue.ExpirationTime,
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
			seasonEpsKey := pkg.GetEpisodeKeyName(episodeInfo.Season, episodeInfo.Episode)
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
	if settings.Get().AdvancedSettings.SaveFullSeasonTmpSubtitles == false {
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
	if settings.Get().EmbySettings.Enable == true && d.embyHelper != nil {

		if job.MediaServerInsideVideoID != "" {
			d.log.Infoln("字幕下载完毕，尝试刷新 Emby 中对应字幕", job.SeriesRootDirPath, job.MediaServerInsideVideoID, job.Season, job.Episode)
			err = d.embyHelper.EmbyApi.UpdateVideoSubList(settings.Get().EmbySettings, job.MediaServerInsideVideoID)
			if err != nil {
				d.log.Errorln("UpdateVideoSubList", job.SeriesRootDirPath, job.MediaServerInsideVideoID, job.Season, job.Episode, "Error:", err)
				return err
			}
		} else {
			d.log.Warningln("字幕下载完毕，尝试刷新 Emby 中对应字幕，跳过，因为 MediaServerInsideVideoID 为空", job.SeriesRootDirPath, job.Season, job.Episode)
		}
	}

	return nil
}
