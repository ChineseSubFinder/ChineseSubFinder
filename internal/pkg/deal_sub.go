package pkg

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"path/filepath"
	"strconv"
)

// GetFrontNameAndOrgName 返回的名称包含，那个网站下载的，这个网站中排名第几，文件名
func GetFrontNameAndOrgName(info *supplier.SubInfo) string {

	infoName := ""
	path, err := GetVideoInfoFromFileName(info.Name)
	if err != nil {
		GetLogger().Warnln("", err)
		infoName = info.Name
	} else {
		infoName = path.Title + "_S" + strconv.Itoa(path.Season) + "E" + strconv.Itoa(path.Episode) + filepath.Ext(info.Name)
	}
	info.Name = infoName

	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN,10) + "_" + infoName
}

// AddFrontName 添加文件的前缀
func AddFrontName(info supplier.SubInfo, orgName string) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN,10) + "_" + orgName
}
