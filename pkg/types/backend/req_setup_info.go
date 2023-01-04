package backend

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
)

type ReqSetupInfo struct {
	Settings settings.Settings `json:"settings" binding:"required"`
}
