package preview_queue

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ffmpeg_helper"
	llq "github.com/emirpasic/gods/queues/linkedlistqueue"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/sirupsen/logrus"
)

type PreviewQueue struct {
	log          *logrus.Logger
	ffmpegHelper *ffmpeg_helper.FFMPEGHelper
	processQueue *llq.Queue
	jobSet       *hashset.Set
	jobResultMap sync.Map
	addOneSignal chan interface{}
	addLocker    sync.Mutex
	workingJob   *Job // 正在操作的任务的路径
}

func NewPreviewQueue(log *logrus.Logger) *PreviewQueue {

	p := &PreviewQueue{
		log:          log,
		ffmpegHelper: ffmpeg_helper.NewFFMPEGHelper(log),
		processQueue: llq.New(),
		jobSet:       hashset.New(),
		jobResultMap: sync.Map{},
		addOneSignal: make(chan interface{}, 1),
		workingJob:   nil,
	}
	go func(pu *PreviewQueue) {
		for {
			select {
			case <-pu.addOneSignal:
				// 有新任务了
				pu.dealers()
			}
		}
	}(p)

	return p
}

// GetVideoHLSAndSubByTimeRangeExportPathInfo 获取视频的HLS和字幕的导出路径信息
func (p *PreviewQueue) GetVideoHLSAndSubByTimeRangeExportPathInfo(videoFullPath string, subFullPaths []string, startTimeString, timeLength string) (string, []string, error) {
	// 导出视频
	if pkg.IsFile(videoFullPath) == false {
		return "", nil, errors.New("video file not exist, maybe is bluray file, so not support yet")
	}

	for _, subFullPath := range subFullPaths {
		if pkg.IsFile(subFullPath) == false {
			return "", nil, errors.New("sub file not exist:" + subFullPath)
		}
	}

	outDirPath, err := pkg.GetVideoAndSubPreviewCacheFolder()
	if err != nil {
		return "", nil, err
	}
	fileName := filepath.Base(videoFullPath)
	frontName := strings.ReplaceAll(fileName, filepath.Ext(fileName), "")
	outDirSubPath := filepath.Join(outDirPath, frontName, startTimeString+"-"+timeLength)

	// 字幕的相对位置
	outSubFPaths := make([]string, 0)
	for i := 0; i < len(subFullPaths); i++ {

		outSubFileFPath := filepath.Join(outDirSubPath, fmt.Sprintf(frontName+"_%d"+common.SubExtSRT, i))

		var subRelPath string
		subRelPath, err = filepath.Rel(outDirPath, outSubFileFPath)
		if err != nil {
			return "", nil, err
		}
		outSubFPaths = append(outSubFPaths, subRelPath)
	}

	// outputlist.m3u8 的相对位置
	outputListRelPath, err := filepath.Rel(outDirPath, filepath.Join(outDirSubPath, "outputlist.m3u8"))
	if err != nil {
		return "", nil, err
	}

	return outputListRelPath, outSubFPaths, nil
}

// IsJobInQueue 是否正在队列中排队，或者正在被处理
func (p *PreviewQueue) IsJobInQueue(job *Job) bool {
	p.addLocker.Lock()
	defer func() {
		p.addLocker.Unlock()
	}()

	if job == nil || job.VideoFPath == "" {
		return false
	}
	if p.jobSet.Contains(job.VideoFPath) == true {
		// 已经在队列中了
		return true
	} else {

		if p.workingJob == nil {
			return false
		}
		// 还有一种可能，任务从队列拿出来了，正在处理，那么在外部开来也还是在队列中的
		if p.workingJob.VideoFPath == job.VideoFPath {
			return true
		}
	}
	return false
}

// Add 添加任务
func (p *PreviewQueue) Add(job *Job) {

	p.addLocker.Lock()
	defer func() {
		p.addLocker.Unlock()
	}()

	if p.jobSet.Contains(job.VideoFPath) == true {
		// 已经在队列中了
		return
	}
	p.processQueue.Enqueue(job)
	p.jobSet.Add(job.VideoFPath)
	// 通知有新任务了
	p.addOneSignal <- struct{}{}

	return
}

// ListJob 任务列表
func (p *PreviewQueue) ListJob() []*Job {

	p.addLocker.Lock()
	defer func() {
		p.addLocker.Unlock()
	}()
	ret := make([]*Job, 0)
	for _, v := range p.processQueue.Values() {
		ret = append(ret, v.(*Job))
	}
	if p.workingJob != nil {
		ret = append(ret, p.workingJob)
	}
	return ret
}

// JobResult 任务结果，如果成功 ok，如果没有就是空，其他就是错误信息
func (p *PreviewQueue) JobResult(job *Job) string {

	value, found := p.jobResultMap.LoadAndDelete(job.VideoFPath)
	if found == false {
		return ""
	}

	return value.(string)
}

func (p *PreviewQueue) dealers() {

	p.addLocker.Lock()
	if p.processQueue.Empty() == true {
		// 没有任务了
		p.addLocker.Unlock()
		return
	}
	job, ok := p.processQueue.Dequeue()
	if ok == false {
		// 没有任务了
		p.addLocker.Unlock()
		return
	}
	// 移除这个任务
	p.jobSet.Remove(job.(*Job).VideoFPath)
	// 标记这个正在处理
	p.workingJob = job.(*Job)
	p.addLocker.Unlock()
	// 具体处理这个任务
	err := p.processSub(job.(*Job))
	if err != nil {
		p.log.Error(err)
	}
}

func (p *PreviewQueue) processSub(job *Job) error {

	var err error
	defer func() {
		// 任务处理完了
		p.addLocker.Lock()
		p.workingJob = nil
		p.addLocker.Unlock()

		if err != nil {
			p.jobResultMap.Store(job.VideoFPath, err.Error())
		} else {
			p.jobResultMap.Store(job.VideoFPath, "ok")
		}
	}()

	const segmentTime = "5.000"
	nowOutRootDirPath, err := pkg.GetVideoAndSubPreviewCacheFolder()
	if err != nil {
		return err
	}
	// 具体处理这个任务，这个任务在加入队列之前就可以预测将要存放在哪，以及名称是什么
	m3u8FPath, subFPath, err := p.ffmpegHelper.ExportVideoHLSAndSubByTimeRange(job.VideoFPath, job.SubFPaths, job.StartTime, job.EndTime, segmentTime, nowOutRootDirPath)
	if err != nil {
		return err
	}
	p.log.Infoln("preview m3u8FPath:", m3u8FPath)
	p.log.Infoln("preview subFPath:", subFPath)

	return nil
}

type Job struct {
	VideoFPath string   `json:"video_f_path"`
	SubFPaths  []string `json:"sub_f_paths"`
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
}

type Reply struct {
	Jobs []*Job `json:"jobs"`
}
