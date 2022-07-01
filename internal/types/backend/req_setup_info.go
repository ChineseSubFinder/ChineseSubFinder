package backend

import (
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
)

type ReqSetupInfo struct {
	Settings settings.Settings `json:"settings" binding:"required"`
}
