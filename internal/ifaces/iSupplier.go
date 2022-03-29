package ifaces

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
)

type ISupplier interface {
	CheckAlive() (bool, int64)

	IsAlive() bool

	GetSupplierName() string

	GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error)

	GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error)

	GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error)
}
