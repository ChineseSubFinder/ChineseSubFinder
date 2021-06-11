package sub_supplier

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/go-rod/rod/lib/utils"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type SubSupplierHub struct {
	Suppliers []ISupplier
	log *logrus.Logger
}

func NewSubSupplierHub(one ISupplier,_inSupplier ... ISupplier) *SubSupplierHub {
	s := SubSupplierHub{}
	s.log = common.GetLogger()
	s.Suppliers = make([]ISupplier, 0)
	s.Suppliers = append(s.Suppliers, one)
	if len(_inSupplier) > 0 {
		for _, supplier := range _inSupplier {
			s.Suppliers = append(s.Suppliers, supplier)
		}
	}
	return &s
}

// DownloadSub 某一个视频的字幕下载，下载完毕后，返回下载缓存每个字幕的位置
func (d SubSupplierHub) DownloadSub(videoFullPath string, index int) ([]string, error) {
	subInfos := d.downloadSub4OneVideo(videoFullPath, index)
	organizeSubFiles, err := d.organizeDlSubFiles(subInfos)
	if err != nil {
		return nil, err
	}
	return organizeSubFiles, nil
}

// downloadSub4OneVideo 为这个视频下载字幕，所有网站找到的字幕都会汇总输出
func (d SubSupplierHub) downloadSub4OneVideo(oneVideoFullPath string, i int) []SubInfo {
	var outSUbInfos = make([]SubInfo, 0)
	// 同时进行查询
	subInfosChannel := make(chan []SubInfo)
	d.log.Infoln("DlSub Start", oneVideoFullPath)
	for _, supplier := range d.Suppliers {
		supplier := supplier
		go func() {
			subInfos, err := d.downloadSub4OneSite(oneVideoFullPath, i, supplier)
			if err != nil {
				d.log.Error(err)
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
	d.log.Infoln(i, "DlSub End", oneVideoFullPath)
	return outSUbInfos
}

// downloadSub4OneSite 在一个站点下载这个视频的字幕
func (d SubSupplierHub) downloadSub4OneSite(oneVideoFullPath string, i int, supplier ISupplier) ([]SubInfo, error) {
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
	d.log.Infoln(i, supplier.GetSupplierName(), "End...")
	return subInfos, nil
}

// organizeDlSubFiles 需要从汇总来是网站字幕中，找到合适的
func (d SubSupplierHub) organizeDlSubFiles(subInfos []SubInfo) ([]string, error) {

	// 缓存列表，整理后的字幕列表
	var siteSubInfoDict = make([]string, 0)
	tmpFolderFullPath, err := common.GetTmpFolder()
	if err != nil {
		return nil, err
	}
	// 先清理缓存目录
	err = common.ClearTmpFolder()
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
			d.log.Error(subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
			continue
		}
		nowExt := strings.ToLower(subInfo.Ext)
		if nowExt != ".zip" && nowExt != ".tar" && nowExt != ".rar" && nowExt != ".7z" {
			// 是否是受支持的字幕类型
			if common.IsSubExtWanted(nowExt) == false {
				continue
			}
			// 加入缓存列表
			siteSubInfoDict = append(siteSubInfoDict, nowFileSaveFullPath)
		} else {
			// 那么就是需要解压的文件了
			// 解压，给一个单独的文件夹
			unzipTmpFolder := path.Join(tmpFolderFullPath, subInfo.FromWhere)
			err = archiver.Unarchive(nowFileSaveFullPath, unzipTmpFolder)
			// 解压完成后，遍历受支持的字幕列表，加入缓存列表
			if err != nil {
				d.log.Error(subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 搜索这个目录下的所有符合字幕格式的文件
			subFileFullPaths, err := d.searchMatchedSubFile(unzipTmpFolder)
			if err != nil {
				d.log.Error(subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 这里需要给这些下载到的文件进行改名，加是从那个网站来的前缀，后续好查找
			for _, fileFullPath := range subFileFullPaths {
				newSubName := d.addFrontName(subInfo, filepath.Base(fileFullPath))
				newSubNameFullPath := path.Join(tmpFolderFullPath, newSubName)
				// 改名
				err = os.Rename(fileFullPath, newSubNameFullPath)
				if err != nil {
					d.log.Error(subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
					continue
				}
				// 加入缓存列表
				siteSubInfoDict = append(siteSubInfoDict, newSubNameFullPath)
			}
		}
	}

	return siteSubInfoDict, nil
}

// searchMatchedSubFile 搜索符合后缀名的视频文件
func (d SubSupplierHub) searchMatchedSubFile(dir string) ([]string, error) {
	// 这里有个梗，会出现 __MACOSX 这类文件夹，那么里面会有一样的文件，需要用文件大小排除一下，至少大于 1 kb 吧
	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 内层的错误就无视了
			oneList, _ := d.searchMatchedSubFile(fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if curFile.Size() < 1000 {
				continue
			}
			if common.IsSubExtWanted(filepath.Ext(curFile.Name())) == true {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// 返回的名称包含，那个网站下载的，这个网站中排名第几，文件名
func (d SubSupplierHub) getFrontNameAndOrgName(info SubInfo) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN,10) + "_" + info.Name
}

// addFrontName 添加文件的前缀
func (d SubSupplierHub) addFrontName(info SubInfo, orgName string) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN,10) + "_" + orgName
}

type SubInfo struct {
	FromWhere string          `json:"from_where"` // 从哪个网站下载来的
	TopN      int64           `json:"top_n"`      // 是 Top 几？
	Name      string          `json:"name"`       // 字幕的名称，这个比较随意，优先是影片的名称，然后才是从网上下载字幕的对应名称
	Language  common.Language `json:"language"`   // 字幕的语言
	FileUrl   string          `json:"file-url"`   // 字幕文件下载的路径
	Score     int64           `json:"score"`      // TODO 字幕的评分，需要有一个独立的评价体系
	Offset    int64           `json:"offset"`     // 字幕的偏移
	Ext       string          `json:"ext"`        // 字幕文件的后缀名带点，有可能是直接能用的字幕文件，也可能是压缩包
	Data      []byte          `json:"data"`       // 字幕文件的二进制数据
}

func NewSubInfo(fromWhere string, topN int64, name string, language common.Language, fileUrl string, score int64, offset int64, ext string, data []byte) *SubInfo {
	return &SubInfo{FromWhere: fromWhere, TopN: topN,Name: name, Language: language, FileUrl: fileUrl, Score: score, Offset: offset, Ext: ext, Data: data}
}

