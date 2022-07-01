package common

import (
	"sync"
)

// SetAccessToken 设置 Web UI 访问的 Token
func SetAccessToken(newToken string) {

	defer mutexAccessToken.Unlock()
	mutexAccessToken.Lock()
	accessToken = newToken
}

// GetAccessToken 获取 Web UI 访问的 Token
func GetAccessToken() string {

	defer mutexAccessToken.Unlock()
	mutexAccessToken.Lock()
	return accessToken
}

// SetApiToken 设置 API 接口访问的 Token
func SetApiToken(newToken string) {

	defer mutexAccessToken.Unlock()
	mutexAccessToken.Lock()
	apiToken = newToken
}

// GetApiToken 获取 API 接口访问的 Token
func GetApiToken() string {

	defer mutexAccessToken.Unlock()
	mutexAccessToken.Lock()
	return apiToken
}

var (
	accessToken      = ""
	apiToken         = ""
	mutexAccessToken sync.Mutex
)
