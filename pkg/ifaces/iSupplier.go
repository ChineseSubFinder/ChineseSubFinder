package ifaces

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/series"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"
	"github.com/sirupsen/logrus"
)

type ISupplier interface {
	CheckAlive() (bool, int64)

	IsAlive() bool

	GetSupplierName() string

	OverDailyDownloadLimit() bool

	GetLogger() *logrus.Logger

	GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error)

	GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error)

	GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error)
}
