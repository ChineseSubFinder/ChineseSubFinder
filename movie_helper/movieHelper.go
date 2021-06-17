package movie_helper

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	_interface "github.com/allanpk716/ChineseSubFinder/interface"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/go-rod/rod/lib/utils"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// OneMovieDlSubInAllSite 一部电影在所有的网站下载相应的字幕
func OneMovieDlSubInAllSite(Suppliers []_interface.ISupplier, oneVideoFullPath string, i int) []common.SupplierSubInfo {
	var outSUbInfos = make([]common.SupplierSubInfo, 0)
	// 同时进行查询
	subInfosChannel := make(chan []common.SupplierSubInfo)
	model.GetLogger().Infoln("DlSub Start", oneVideoFullPath)
	for _, supplier := range Suppliers {
		supplier := supplier
		go func() {
			subInfos, err := OneMovieDlSubInOneSite(oneVideoFullPath, i, supplier)
			if err != nil {
				model.GetLogger().Errorln("oneMovieDlSubInOneSite", err)
			}
			subInfosChannel <- subInfos
		}()
	}
	for i := 0; i < len(Suppliers); i++ {
		v, ok := <-subInfosChannel
		if ok == true {
			outSUbInfos = append(outSUbInfos, v...)
		}
	}
	model.GetLogger().Infoln("DlSub End", oneVideoFullPath)
	return outSUbInfos
}

// OneMovieDlSubInOneSite 一部电影在一个站点下载字幕
func OneMovieDlSubInOneSite(oneVideoFullPath string, i int, supplier _interface.ISupplier) ([]common.SupplierSubInfo, error) {
	defer func() {
		model.GetLogger().Infoln(i, supplier.GetSupplierName(), "End...")
	}()
	model.GetLogger().Infoln(i, supplier.GetSupplierName(), "Start...")
	subInfos, err := supplier.GetSubListFromFile4Movie(oneVideoFullPath)
	if err != nil {
		return nil, err
	}
	// 把后缀名给改好
	for x, info := range subInfos {
		tmpSubFileName := info.Name
		if strings.Contains(tmpSubFileName, info.Ext) == false {
			subInfos[x].Name = tmpSubFileName + info.Ext
		}
	}

	return subInfos, nil
}

// OrganizeMovieDlSubFiles 需要从汇总来是网站字幕中，找到合适的
func OrganizeMovieDlSubFiles(subInfos []common.SupplierSubInfo) ([]string, error) {

	// 缓存列表，整理后的字幕列表
	var siteSubInfoDict = make([]string, 0)
	tmpFolderFullPath, err := model.GetTmpFolder()
	if err != nil {
		return nil, err
	}
	// 先清理缓存目录
	err = model.ClearTmpFolder()
	if err != nil {
		return nil, err
	}
	// 第三方的解压库，首先不支持 io.Reader 的操作，也就是得缓存到本地硬盘再读取解压
	// 且使用 walk 会无法解压 rar，得指定具体的实例，太麻烦了，直接用通用的接口得了，就是得都缓存下来再判断
	// 基于以上两点，写了一堆啰嗦的逻辑···
	for _, subInfo := range subInfos {
		// 先存下来，保存是时候需要前缀，前缀就是从那个网站下载来的
		nowFileSaveFullPath := path.Join(tmpFolderFullPath, model.GetFrontNameAndOrgName(subInfo))
		err = utils.OutputFile(nowFileSaveFullPath, subInfo.Data)
		if err != nil {
			model.GetLogger().Errorln("getFrontNameAndOrgName - OutputFile",subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
			continue
		}
		nowExt := strings.ToLower(subInfo.Ext)
		if nowExt != ".zip" && nowExt != ".tar" && nowExt != ".rar" && nowExt != ".7z" {
			// 是否是受支持的字幕类型
			if model.IsSubExtWanted(nowExt) == false {
				continue
			}
			// 加入缓存列表
			siteSubInfoDict = append(siteSubInfoDict, nowFileSaveFullPath)
		} else {
			// 那么就是需要解压的文件了
			// 解压，给一个单独的文件夹
			unzipTmpFolder := path.Join(tmpFolderFullPath, subInfo.FromWhere)
			err = os.MkdirAll(unzipTmpFolder, os.ModePerm)
			if err != nil {
				return nil, err
			}
			err = model.UnArchiveFile(nowFileSaveFullPath, unzipTmpFolder)
			// 解压完成后，遍历受支持的字幕列表，加入缓存列表
			if err != nil {
				model.GetLogger().Errorln("archiver.UnArchive", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 搜索这个目录下的所有符合字幕格式的文件
			subFileFullPaths, err := model.SearchMatchedSubFile(unzipTmpFolder)
			if err != nil {
				model.GetLogger().Errorln("searchMatchedSubFile", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 这里需要给这些下载到的文件进行改名，加是从那个网站来的前缀，后续好查找
			for _, fileFullPath := range subFileFullPaths {
				newSubName := model.AddFrontName(subInfo, filepath.Base(fileFullPath))
				newSubNameFullPath := path.Join(tmpFolderFullPath, newSubName)
				// 改名
				err = os.Rename(fileFullPath, newSubNameFullPath)
				if err != nil {
					model.GetLogger().Errorln("os.Rename", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
					continue
				}
				// 加入缓存列表
				siteSubInfoDict = append(siteSubInfoDict, newSubNameFullPath)
			}
		}
	}

	return siteSubInfoDict, nil
}

// MovieHasSub 这个视频文件的目录下面有字幕文件了没有
func MovieHasSub(videoFilePath string) (bool, error) {
	dir := filepath.Dir(videoFilePath)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, curFile := range files {
		if curFile.IsDir() {
			continue
		} else {
			// 文件
			if model.IsSubExtWanted(curFile.Name()) == true {
				return true, nil
			}
		}
	}

	return false, nil
}

func SkipChineseMovie(videoFullPath string, _reqParam ...common.ReqParam) (bool, error) {
	var reqParam common.ReqParam
	if len(_reqParam) > 0 {
		reqParam = _reqParam[0]
	}
	imdbInfo, err := model.GetImdbInfo(filepath.Dir(videoFullPath))
	if err != nil {
		return false, err
	}
	t, err := model.GetVideoInfoFromIMDB(imdbInfo.ImdbId, reqParam)
	if err != nil {
		return false, err
	}
	if len(t.Languages) > 0 && strings.ToLower(t.Languages[0]) == "chinese" {
		model.GetLogger().Infoln("Skip", videoFullPath, "Sub Download, because movie is Chinese")
		return true, nil
	}
	return false, nil
}

func MoiveNeedDlSub(videoFullPath string) (bool, error) {
	// 视频下面有不有字幕
	found, err := MovieHasSub(videoFullPath)
	if err != nil {
		return false, err
	}
	// 资源下载的时间后的多少天内都进行字幕的自动下载，替换原有的字幕
	currentTime := time.Now()
	dayRange, _ := time.ParseDuration(common.DownloadSubDuring30Days)
	_, modifyTime, err := model.GetVideoInfoFromFileFullPath(videoFullPath)
	if err != nil {
		return false, err
	}
	// 30 天内，或者没有字幕都要进行下载
	if modifyTime.Add(dayRange).After(currentTime) == true || found == false {
		// 需要下载的
		return true, nil
	} else {
		if modifyTime.Add(dayRange).After(currentTime) == false {
			model.GetLogger().Infoln("Skip", videoFullPath, "Sub Download, because movie has sub and downloaded more than 30 days")
			return false, nil
		}
		if found == true {
			model.GetLogger().Infoln("Skip", videoFullPath, "Sub Download, because sub file found")
			return false, nil
		}

		return false, nil
	}
}
