package ChineseSubFinder

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/zimuku"
	"github.com/go-rod/rod/lib/utils"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Downloader struct {
	reqParam common.ReqParam
	topic int					// 最多能够下载 Top 几的字幕，每一个网站
	wantedExtList []string		// 人工确认的需要监控的视频后缀名
	defExtList []string			// 内置支持的视频后缀名列表
}

func NewDownloader(_reqParam ... common.ReqParam) *Downloader {

	var downloader Downloader
	downloader.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		downloader.reqParam = _reqParam[0]
		if downloader.reqParam.Topic > 0 && downloader.reqParam.Topic != downloader.topic {
			downloader.topic = downloader.reqParam.Topic
		}
	}
	downloader.defExtList = make([]string, 0)
	downloader.defExtList = append(downloader.defExtList, VideoExtMp4)
	downloader.defExtList = append(downloader.defExtList, VideoExtMkv)
	downloader.defExtList = append(downloader.defExtList, VideoExtRmvb)
	downloader.defExtList = append(downloader.defExtList, VideoExtIso)

	if len(_reqParam) > 0 {
		// 如果用户设置了关注的视频后缀名列表，则用ta的
		if len(downloader.reqParam.UserExtList) > 0 {
			downloader.wantedExtList = downloader.reqParam.UserExtList
		} else {
			// 不然就是内置默认的
			downloader.wantedExtList = downloader.defExtList
		}
	} else {
		// 不然就是内置默认的
		downloader.wantedExtList = downloader.defExtList
	}
	return &downloader
}

func (d Downloader) GetNowSupportExtList() []string {
	return d.wantedExtList
}

func (d Downloader) GetDefSupportExtList() []string {
	return d.defExtList
}

func (d Downloader) DownloadSub(dir string) error {
	nowVideoList, err := d.searchFile(dir)
	if err != nil {
		return err
	}
	// 构建每个字幕站点下载者的实例
	var suppliers = make([]sub_supplier.ISupplier, 0)
	suppliers = append(suppliers, shooter.NewSupplier(d.reqParam))
	suppliers = append(suppliers, subhd.NewSupplier(d.reqParam))
	suppliers = append(suppliers, xunlei.NewSupplier(d.reqParam))
	suppliers = append(suppliers, zimuku.NewSupplier(d.reqParam))
	// TODO 后续再改为每个视频以上的流程都是一个 channel 来做，并且需要控制在一个并发量之下（很可能没必要，毕竟要在弱鸡机器上挂机用的）
	// 一个视频文件同时多个站点查询，阻塞完毕后，在进行下一个
	for i, oneVideoFullPath := range nowVideoList {
		ontVideoRootPath := filepath.Base(oneVideoFullPath)
		// 同时进行查询
		wg := sync.WaitGroup{}
		wg.Add(len(suppliers))
		println("DlSub Start", oneVideoFullPath)
		for _, supplier := range suppliers {
			println(i, supplier.GetSupplierName(), "Start...")
			subInfos, err := supplier.GetSubListFromFile(oneVideoFullPath)
			if err != nil {
				println(supplier.GetSupplierName(), "GetSubListFromFile", err.Error())
				wg.Done()
				continue
			}

			if d.reqParam.DebugMode == true {
				// 需要进行字幕文件的缓存
				// 把缓存的文件夹新建出来
				desFolderFullPath := path.Join(ontVideoRootPath, SubTmpFolderName)
				err = os.MkdirAll(desFolderFullPath, os.ModePerm)
				if err != nil{
					println(supplier.GetSupplierName(), "MkdirAll", err.Error())
					wg.Done()
					continue
				}
				for x, info := range subInfos {
					tmpSubFileName := info.Name
					if strings.Contains(tmpSubFileName, info.Ext) == false {
						tmpSubFileName = tmpSubFileName + info.Ext
					}
					desSubFileFullPath := path.Join(desFolderFullPath, strconv.Itoa(x) + "_" + tmpSubFileName)
					err = utils.OutputFile(desSubFileFullPath, info.Data)
					if err != nil {
						println(supplier.GetSupplierName(), "WriteSubFile", info.Name, err.Error())
						continue
					}
				}
			}

			println(supplier.GetSupplierName(), "End...")
			wg.Done()
		}
		println(i, "DlSub End", oneVideoFullPath)
		wg.Wait()
	}

	return nil
}

func (d Downloader)searchFile(dir string) ([]string, error) {

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
			oneList, _ := d.searchFile(fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if d.isWantedExtDef(curFile.Name()) == true {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

func (d Downloader) isWantedExtDef(fileName string) bool {
	fileName = strings.ToLower(filepath.Ext(fileName))
	for _, s := range d.wantedExtList {
		if s == fileName {
			return true
		}
	}
	return false
}

const (
	VideoExtMp4 = ".mp4"
	VideoExtMkv = ".mkv"
	VideoExtRmvb = ".rmvb"
	VideoExtIso = ".iso"

	SubTmpFolderName = "subTmp"
)