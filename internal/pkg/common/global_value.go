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

// SetSubScanJobStatusRunning 设置扫描字幕任务的状态为运行
func SetSubScanJobStatusRunning(WorkingUnitIndex, UnitCount int, WorkingUnitName string,
	WorkingVideoIndex, VideoCount int, WorkingVideoName string) {

	subDownloadJobInfoLock.Lock()
	if subDownloadJobInfo == nil {
		subDownloadJobInfo = &ws.SubDownloadJobInfo{}
	}
	subDownloadJobInfo.Status = ws.Running
	subDownloadJobInfo.WorkingUnitIndex = WorkingUnitIndex
	subDownloadJobInfo.UnitCount = UnitCount
	subDownloadJobInfo.WorkingUnitName = WorkingUnitName

	subDownloadJobInfo.WorkingVideoIndex = WorkingVideoIndex
	subDownloadJobInfo.VideoCount = VideoCount
	subDownloadJobInfo.WorkingVideoName = WorkingVideoName

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
