package backend

import "github.com/allanpk716/ChineseSubFinder/internal/models"

type ReqSetupInfo struct {
	UserInfo models.UserInfo `json:"user_info" binding:"required"`
}
