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

var (
	accessToken      = ""
	mutexAccessToken sync.Mutex
)

var (
	IsSubDownloadJobInfoRunning = false
	subDownloadJobInfo          *ws.SubDownloadJobInfo
	subDownloadJobInfoLock      sync.Mutex
)
