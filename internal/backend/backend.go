package backend

import (
	"github.com/google/uuid"
)

// GenerateAccessToken 生成随机的 AccessToken
func GenerateAccessToken() string {
	u4 := uuid.New()
	return u4.String()
}
