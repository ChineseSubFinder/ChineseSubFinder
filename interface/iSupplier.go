package _interface

import "github.com/allanpk716/ChineseSubFinder/common"

type ISupplier interface {

	GetSupplierName() string

	GetReqParam() common.ReqParam

	GetSubListFromFile4Movie(filePath string) ([]common.SupplierSubInfo, error)

	GetSubListFromFile4Series(seriesInfo *common.SeriesInfo) ([]common.SupplierSubInfo, error)
	
	GetSubListFromFile4Anime(seriesInfo *common.SeriesInfo) ([]common.SupplierSubInfo, error)
}