package common

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend/ws"
	"github.com/huandu/go-clone"
	"sync"
)

func SetAccessToken(newToken string) {

	defer mutexAccessToken.Unlock()
	mutexAccessToken.Lock()
	accessToken = newToken
}

func GetAccessToken() string {

	defer mutexAccessToken.Unlock()
	mutexAccessToken.Lock()
	return accessToken
}

// GetSubScanJobStatus 只用于获取，不是设置值，因为创建了实例的副本
func GetSubScanJobStatus() *ws.SubDownloadJobInfo {

	var tmpSubDownloadJobInfoLock *ws.SubDownloadJobInfo
	subDownloadJobInfoLock.Lock()
	tmpSubDownloadJobInfoLock = clone.Clone(subDownloadJobInfo).(*ws.SubDownloadJobInfo)
	subDownloadJobInfoLock.Unlock()
	return tmpSubDownloadJobInfoLock
}

// SetSubScanJobStatusPreparing 设置扫描字幕任务的状态为准备
func SetSubScanJobStatusPreparing(startedTime string) {

	subDownloadJobInfoLock.Lock()
	if subDownloadJobInfo == nil {
		subDownloadJobInfo = &ws.SubDownloadJobInfo{}
	}
	subDownloadJobInfo.Status = ws.Preparing
	subDownloadJobInfo.StartedTime = startedTime
	subDownloadJobInfoLock.Unlock()
}

// SetSubScanJobStatusScanMovie 设置扫描字幕任务的状态为运行
func SetSubScanJobStatusScanMovie(WorkingUnitIndex, UnitCount int, WorkingUnitName string) {

	subDownloadJobInfoLock.Lock()

	if subDownloadJobInfo == nil {
		subDownloadJobInfo = &ws.SubDownloadJobInfo{}
	}
	subDownloadJobInfo.Status = ws.ScanMovie
	subDownloadJobInfo.WorkingUnitIndex = WorkingUnitIndex
	subDownloadJobInfo.UnitCount = UnitCount
	subDownloadJobInfo.WorkingUnitName = WorkingUnitName

	subDownloadJobInfoLock.Unlock()
}

// SetSubScanJobStatusScanSeriesMain 设置扫描字幕任务的状态为运行
func SetSubScanJobStatusScanSeriesMain(WorkingUnitIndex, UnitCount int, WorkingUnitName string) {

	subDownloadJobInfoLock.Lock()

	if subDownloadJobInfo == nil {
		subDownloadJobInfo = &ws.SubDownloadJobInfo{}
	}
	subDownloadJobInfo.Status = ws.ScanSeries
	subDownloadJobInfo.WorkingUnitIndex = WorkingUnitIndex
	subDownloadJobInfo.UnitCount = UnitCount
	subDownloadJobInfo.WorkingUnitName = WorkingUnitName

	subDownloadJobInfo.WorkingVideoIndex = 0
	subDownloadJobInfo.VideoCount = 0
	subDownloadJobInfo.WorkingVideoName = ""

	subDownloadJobInfoLock.Unlock()
}

// SetSubScanJobStatusScanSeriesSub 设置扫描字幕任务的状态为运行
func SetSubScanJobStatusScanSeriesSub(WorkingVideoIndex, VideoCount int, WorkingVideoName string) {

	subDownloadJobInfoLock.Lock()

	if subDownloadJobInfo == nil {
		subDownloadJobInfo = &ws.SubDownloadJobInfo{}
	}
	subDownloadJobInfo.Status = ws.ScanSeries
	update := false
	// 因为这里不同于movie的逻辑，movie 是在字幕下载者之前进行的统计更新
	// 而对应 series 则是进入到具体的一个下载者中进行的任务进度更新，所以需要考虑并发情况，以最大值来更新
	if subDownloadJobInfo.WorkingVideoIndex < WorkingVideoIndex {
		subDownloadJobInfo.WorkingVideoIndex = WorkingVideoIndex
		update = true
	}

	if subDownloadJobInfo.VideoCount < VideoCount {
		subDownloadJobInfo.VideoCount = VideoCount
		update = true
	}

	if update == true {
		subDownloadJobInfo.WorkingVideoName = WorkingVideoName
	}

	subDownloadJobInfoLock.Unlock()
}

// SetSubScanJobStatusWaiting 设置扫描字幕任务的状态为等待
func SetSubScanJobStatusWaiting(startedTime string) {

	subDownloadJobInfoLock.Lock()
	if subDownloadJobInfo == nil {
		subDownloadJobInfo = &ws.SubDownloadJobInfo{}
	}
	subDownloadJobInfo.Status = ws.Waiting
	subDownloadJobInfo.StartedTime = startedTime
	subDownloadJobInfoLock.Unlock()
}

// SetSubScanJobStatusNil 如果总任务停止了，那么就需要设置为 nil，这样定时器发送的时候就会判断是否为 nil，是就不会继续触发
// 如果总任务开始了，是否是立即开始都会由实例化操作
func SetSubScanJobStatusNil() {

	subDownloadJobInfoLock.Lock()
	subDownloadJobInfo = nil
	subDownloadJobInfoLock.Unlock()
}

var (
	accessToken      = ""
	mutexAccessToken sync.Mutex
)

var (
	subDownloadJobInfo     *ws.SubDownloadJobInfo
	subDownloadJobInfoLock sync.Mutex
)
