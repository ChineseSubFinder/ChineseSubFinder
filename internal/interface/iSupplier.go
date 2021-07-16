package _interface

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
)

type ISupplier interface {

	GetSupplierName() string

	GetReqParam() types.ReqParam

	GetSubListFromFile4Movie(filePath string) ([]supplier.SubInfo, error)

	GetSubListFromFile4Series(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error)
	
	GetSubListFromFile4Anime(seriesInfo *series.SeriesInfo) ([]supplier.SubInfo, error)
}