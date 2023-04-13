package manual_upload_sub_2_local

import (
	"sync"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/scan_logic"
	"github.com/pkg/errors"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/save_sub_helper"
	subCommon "github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"

	"github.com/sirupsen/logrus"

	llq "github.com/emirpasic/gods/queues/linkedlistqueue"
	"github.com/emirpasic/gods/sets/hashset"
)

type ManualUploadSub2Local struct {
	log              *logrus.Logger
	saveSubHelper    *save_sub_helper.SaveSubHelper // 保存字幕的逻辑
	scanLogic        *scan_logic.ScanLogic          // 是否扫描逻辑
	subNameFormatter subCommon.FormatterName        // 从 inSubFormatter 推断出来
	processQueue     *llq.Queue
	jobSet           *hashset.Set
	jobResultMap     sync.Map
	addOneSignal     chan interface{}
	addLocker        sync.Mutex
	subParserHub     *sub_parser_hub.SubParserHub
	workingJob       *Job // 正在操作的任务的路径
}

func NewManualUploadSub2Local(log *logrus.Logger, saveSubHelper *save_sub_helper.SaveSubHelper, scanLogic *scan_logic.ScanLogic) *ManualUploadSub2Local {

	m := &ManualUploadSub2Local{
		log:           log,
		saveSubHelper: saveSubHelper,
		scanLogic:     scanLogic,
		processQueue:  llq.New(),
		jobSet:        hashset.New(),
		jobResultMap:  sync.Map{},
		addOneSignal:  make(chan interface{}, 1),
		subParserHub:  sub_parser_hub.NewSubParserHub(log, ass.NewParser(log), srt.NewParser(log)),
		workingJob:    nil,
	}

	// 这里就不单独弄一个 settings.SubNameFormatter 字段来传递值了，因为 inSubFormatter 就已经知道是什么 formatter 了
	m.subNameFormatter = subCommon.FormatterName(saveSubHelper.SubFormatter.GetFormatterFormatterName())

	go func(mu *ManualUploadSub2Local) {
		for {
			select {
			case _ = <-mu.addOneSignal:
				// 有新任务了
				m.dealers()
			}
		}
	}(m)

	return m
}

// IsJobInQueue 是否正在队列中排队，或者正在被处理
func (m *ManualUploadSub2Local) IsJobInQueue(job *Job) bool {
	m.addLocker.Lock()
	defer func() {
		m.addLocker.Unlock()
	}()

	if job == nil || job.VideoFPath == "" {
		return false
	}
	if m.jobSet.Contains(job.VideoFPath) == true {
		// 已经在队列中了
		return true
	} else {

		if m.workingJob == nil {
			return false
		}
		// 还有一种可能，任务从队列拿出来了，正在处理，那么在外部开来也还是在队列中的
		if m.workingJob.VideoFPath == job.VideoFPath {
			return true
		}
	}
	return false
}

// Add 添加任务
func (m *ManualUploadSub2Local) Add(job *Job) {

	m.addLocker.Lock()
	defer func() {
		m.addLocker.Unlock()
	}()

	if m.jobSet.Contains(job.VideoFPath) == true {
		// 已经在队列中了
		return
	}
	m.processQueue.Enqueue(job)
	m.jobSet.Add(job.VideoFPath)
	// 通知有新任务了
	m.addOneSignal <- struct{}{}

	return
}

// JobResult 任务结果，如果成功 ok，如果没有就是空，其他就是错误信息
func (m *ManualUploadSub2Local) JobResult(job *Job) string {

	value, found := m.jobResultMap.LoadAndDelete(job.VideoFPath)
	if found == false {
		return ""
	}

	return value.(string)
}

// ListJob 任务列表
func (m *ManualUploadSub2Local) ListJob() []*Job {

	m.addLocker.Lock()
	defer func() {
		m.addLocker.Unlock()
	}()
	ret := make([]*Job, 0)
	for _, v := range m.processQueue.Values() {
		ret = append(ret, v.(*Job))
	}
	if m.workingJob != nil {
		ret = append(ret, m.workingJob)
	}
	return ret
}

func (m *ManualUploadSub2Local) dealers() {

	m.addLocker.Lock()
	if m.processQueue.Empty() == true {
		// 没有任务了
		m.addLocker.Unlock()
		return
	}
	job, ok := m.processQueue.Dequeue()
	if ok == false {
		// 没有任务了
		m.addLocker.Unlock()
		return
	}
	// 移除这个任务
	m.jobSet.Remove(job.(*Job).VideoFPath)
	// 标记这个正在处理
	m.workingJob = job.(*Job)
	m.addLocker.Unlock()
	// 具体处理这个任务
	err := m.processSub(job.(*Job))
	if err != nil {
		m.log.Error(err)
	}
}

func (m *ManualUploadSub2Local) processSub(job *Job) error {

	var err error
	defer func() {
		// 任务处理完了
		m.addLocker.Lock()
		m.workingJob = nil
		m.addLocker.Unlock()

		if err != nil {
			m.jobResultMap.Store(job.VideoFPath, err.Error())
		} else {
			m.jobResultMap.Store(job.VideoFPath, "ok")
		}
	}()

	// 不管是不是保存多个字幕，都要先扫描本地的字幕，进行 .Default .Forced 去除
	// 这个视频的所有字幕，去除 .default .Forced 标记
	err = sub_helper.SearchVideoMatchSubFileAndRemoveExtMark(m.log, job.VideoFPath)
	if err != nil {
		// 找个错误可以忍
		m.log.Errorln("SearchVideoMatchSubFileAndRemoveExtMark,", job.VideoFPath, err)
	}

	bFind, subFileInfo, err := m.subParserHub.DetermineFileTypeFromFile(job.SubFPath)
	if err != nil {
		err = errors.New("DetermineFileTypeFromFile," + job.SubFPath + "," + err.Error())
		return err
	}
	if bFind == false {
		err = errors.New("DetermineFileTypeFromFile," + job.SubFPath + ",not support SubType")
		return err
	}

	var skipInfo *models.SkipScanInfo
	if m.subNameFormatter == subCommon.Emby {
		err = m.saveSubHelper.WriteSubFile2VideoPath(job.VideoFPath, *subFileInfo, "manual", true, false)
		if err != nil {
			err = errors.New("WriteSubFile2VideoPath," + job.VideoFPath + "," + err.Error())
			return err
		}
		// 默认设置这个视频“跳过”（跳过扫描和下载字幕）属性
		skipInfo = models.NewSkipScanInfoByMovie(job.VideoFPath, true)
	} else {
		err = m.saveSubHelper.WriteSubFile2VideoPath(job.VideoFPath, *subFileInfo, "manual", false, false)
		if err != nil {
			err = errors.New("WriteSubFile2VideoPath," + job.VideoFPath + "," + err.Error())
			return err
		}
		// 默认设置这个视频“跳过”（跳过扫描和下载字幕）属性
		skipInfo = models.NewSkipScanInfoBySeriesEx(job.VideoFPath, true)
	}

	m.scanLogic.Set(skipInfo)

	return nil
}

type Job struct {
	VideoFPath string `json:"video_f_path"`
	SubFPath   string `json:"sub_f_path"`
}

type Reply struct {
	Jobs []*Job `json:"jobs"`
}
