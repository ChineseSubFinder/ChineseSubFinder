package _interface

import "github.com/allanpk716/ChineseSubFinder/common"

type ISupplier interface {

	GetSupplierName() string

	GetSubListFromFile4Movie(filePath string) ([]common.SupplierSubInfo, error)

	GetSubListFromFile4Series(filePath string) ([]common.SupplierSubInfo, error)
	
	GetSubListFromFile4Anime(filePath string) ([]common.SupplierSubInfo, error)

	GetSubListFromFile(filePath string) ([]common.SupplierSubInfo, error)

	GetSubListFromKeyword(keyword string) ([]common.SupplierSubInfo, error)
}