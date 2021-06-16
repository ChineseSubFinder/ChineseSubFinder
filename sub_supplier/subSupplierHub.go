package sub_supplier

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/interface"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/go-rod/rod/lib/utils"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type SubSupplierHub struct {
	Suppliers []_interface.ISupplier
	log *logrus.Logger
}

func NewSubSupplierHub(one _interface.ISupplier,_inSupplier ..._interface.ISupplier) *SubSupplierHub {
	s := SubSupplierHub{}
	s.log = model.GetLogger()
	s.Suppliers = make([]_interface.ISupplier, 0)
	s.Suppliers = append(s.Suppliers, one)
	if len(_inSupplier) > 0 {
		for _, supplier := range _inSupplier {
			s.Suppliers = append(s.Suppliers, supplier)
		}
	}
	return &s
}

// DownloadSub4Movie 某一个视频的字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub4Movie(videoFullPath string, index int, foundExistSubFileThanSkip bool) ([]string, error) {
	// 先清理缓存文件夹
	err := model.ClearTmpFolder()
	if err != nil {
		d.log.Error(err)
	}

	// TODO 这里可以考虑改为，要么就是视频内封有字幕（得研究下怎么判断） 。要么，就是在视频上线（影片上映时间，电影院时间 or DVD 时间，资源能够下载的时间？）后的多少天内都进行字幕的自动下载，替换原有的字幕
	// 是否需要跳过有字幕文件的视频
	if foundExistSubFileThanSkip == true {
		found, err := d.videoHasSub(videoFullPath)
		if err != nil {
			return nil, err
		}
		if found == true {
			d.log.Infoln("Skip", videoFullPath, "Sub Download, because sub file found")
			return nil, nil
		}
	}
	subInfos := d.downloadSub4OneVideo(videoFullPath, index)
	organizeSubFiles, err := d.organizeDlSubFiles(subInfos)
	if err != nil {
		return nil, err
	}
	return organizeSubFiles, nil
}

// downloadSub4OneVideo 为这个视频下载字幕，所有网站找到的字幕都会汇总输出
func (d SubSupplierHub) downloadSub4OneVideo(oneVideoFullPath string, i int) []common.SupplierSubInfo {
	var outSUbInfos = make([]common.SupplierSubInfo, 0)
	// 同时进行查询
	subInfosChannel := make(chan []common.SupplierSubInfo)
	d.log.Infoln("DlSub Start", oneVideoFullPath)
	for _, supplier := range d.Suppliers {
		supplier := supplier
		go func() {
			subInfos, err := d.downloadSub4OneSite(oneVideoFullPath, i, supplier)
			if err != nil {
				d.log.Errorln("downloadSub4OneSite", err)
			}
			subInfosChannel <- subInfos
		}()
	}
	for i := 0; i < len(d.Suppliers); i++ {
		v, ok := <-subInfosChannel
		if ok == true {
			outSUbInfos = append(outSUbInfos, v...)
		}
	}
	d.log.Infoln("DlSub End", oneVideoFullPath)
	return outSUbInfos
}

// downloadSub4OneSite 在一个站点下载这个视频的字幕
func (d SubSupplierHub) downloadSub4OneSite(oneVideoFullPath string, i int, supplier _interface.ISupplier) ([]common.SupplierSubInfo, error) {
	defer func() {
		d.log.Infoln(i, supplier.GetSupplierName(), "End...")
	}()
	d.log.Infoln(i, supplier.GetSupplierName(), "Start...")
	subInfos, err := supplier.GetSubListFromFile(oneVideoFullPath)
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

// organizeDlSubFiles 需要从汇总来是网站字幕中，找到合适的
func (d SubSupplierHub) organizeDlSubFiles(subInfos []common.SupplierSubInfo) ([]string, error) {

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
		nowFileSaveFullPath := path.Join(tmpFolderFullPath, d.getFrontNameAndOrgName(subInfo))
		err = utils.OutputFile(nowFileSaveFullPath, subInfo.Data)
		if err != nil {
			d.log.Errorln("getFrontNameAndOrgName - OutputFile",subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
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
				d.log.Errorln("archiver.Unarchive", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 搜索这个目录下的所有符合字幕格式的文件
			subFileFullPaths, err := model.SearchMatchedSubFile(unzipTmpFolder)
			if err != nil {
				d.log.Errorln("searchMatchedSubFile", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 这里需要给这些下载到的文件进行改名，加是从那个网站来的前缀，后续好查找
			for _, fileFullPath := range subFileFullPaths {
				newSubName := d.addFrontName(subInfo, filepath.Base(fileFullPath))
				newSubNameFullPath := path.Join(tmpFolderFullPath, newSubName)
				// 改名
				err = os.Rename(fileFullPath, newSubNameFullPath)
				if err != nil {
					d.log.Errorln("os.Rename", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
					continue
				}
				// 加入缓存列表
				siteSubInfoDict = append(siteSubInfoDict, newSubNameFullPath)
			}
		}
	}

	return siteSubInfoDict, nil
}

// 返回的名称包含，那个网站下载的，这个网站中排名第几，文件名
func (d SubSupplierHub) getFrontNameAndOrgName(info common.SupplierSubInfo) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN,10) + "_" + info.Name
}

// addFrontName 添加文件的前缀
func (d SubSupplierHub) addFrontName(info common.SupplierSubInfo, orgName string) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN,10) + "_" + orgName
}

// videoHasSub 这个视频文件的目录下面有字幕文件了没有
func (d SubSupplierHub) videoHasSub(videoFilePath string) (bool, error) {
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



