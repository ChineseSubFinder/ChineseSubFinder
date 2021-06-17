package model

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"strconv"
)

// GetFrontNameAndOrgName 返回的名称包含，那个网站下载的，这个网站中排名第几，文件名
func GetFrontNameAndOrgName(info common.SupplierSubInfo) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN,10) + "_" + info.Name
}

// AddFrontName 添加文件的前缀
func AddFrontName(info common.SupplierSubInfo, orgName string) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN,10) + "_" + orgName
}
