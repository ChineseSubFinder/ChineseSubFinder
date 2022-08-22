//go:build windows

package extend_log

import (
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
	"github.com/sirupsen/logrus"
)

type ExtendLog struct {
}

func (e *ExtendLog) AddHook(log *logrus.Logger, extendLog settings.ExtendLog) {

}
