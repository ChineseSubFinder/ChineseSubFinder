package _interface

import "github.com/allanpk716/ChineseSubFinder/common"

type ISupplier interface {

	GetSupplierName() string

	GetReqParam() common.ReqParam

	GetSubListFromFile4Movie(filePath string) ([]common.SupplierSubInfo, error)

	GetSubListFromFile4Series(seriesPath string) ([]common.SupplierSubInfo, error)
	
	GetSubListFromFile4Anime(AnimePath string) ([]common.SupplierSubInfo, error)

	GetSubListFromFile(filePath string) ([]common.SupplierSubInfo, error)

	GetSubListFromKeyword(keyword string) ([]common.SupplierSubInfo, error)
}