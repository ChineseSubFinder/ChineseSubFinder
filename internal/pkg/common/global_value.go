package common

import (
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

var (
	accessToken      = ""
	mutexAccessToken sync.Mutex
)
