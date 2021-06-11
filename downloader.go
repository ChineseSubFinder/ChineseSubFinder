package ChineseSubFinder

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/zimuku"
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

type Downloader struct {
	reqParam common.ReqParam
	log *logrus.Logger
	topic int					// 最多能够下载 Top 几的字幕，每一个网站
	wantedExtList []string		// 人工确认的需要监控的视频后缀名
	defExtList []string			// 内置支持的视频后缀名列表
}

func NewDownloader(_reqParam ... common.ReqParam) *Downloader {

	var downloader Downloader
	downloader.log = common.GetLogger()
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
	nowVideoList, err := d.searchMatchedVideoFile(dir)
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
		nowSubInfos := d.downloadSub4OneVideo(oneVideoFullPath, suppliers, i)
		// 字幕都下载缓存好了，需要抉择存哪一个，优先选择中文双语的，然后到中文
		err = d.chooseAndSaveSubFile(oneVideoFullPath, nowSubInfos)
		if err != nil {
			d.log.Error(oneVideoFullPath, "Download Sub Error",err)
			continue
		}
	}
	return nil
}

// chooseAndSaveSubFile 需要从汇总来是网站字幕中，找到合适的
func (d Downloader) chooseAndSaveSubFile(oneVideoFullPath string, subInfos []sub_supplier.SubInfo) error {

	// 得到目标视频文件的根目录
	videoRootPath := filepath.Dir(oneVideoFullPath)
	tmpFolderFullPath, err := common.GetTmpFolder()
	if err != nil {
		return err
	}

	var siteSubInfoDict = make([]string, 0)
	// 第三方的解压函数，首先不支持 io.Reader 的操作，也就是得缓存到本地硬盘再读取解压
	// 且使用 walk 会无法解压 rar，得指定具体的实例，太麻烦了，直接用通用的接口得了，就是得都缓存下来再判断
	for _, subInfo := range subInfos {
		// TODO 这里先处理 Top1 的字幕，后续再考虑怎么觉得 Top N 选择哪一个，很可能选择每个网站 Top 1就行了，具体的过滤逻辑在其内部实现
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
			for _, fileFullPath := range subFileFullPaths {
				// 加入缓存列表
				siteSubInfoDict = append(siteSubInfoDict, fileFullPath)
			}
		}
	}
	// 拿到现有的字幕列表，开始抉择
	// 还需要考虑，判断这个字幕是简体还是繁体
	
	println(videoRootPath)

	// 抉择完毕，需要清理缓存目录
	err = common.ClearTmpFolder()
	if err != nil {
		return err
	}
	return nil
}

// downloadSub4OneVideo 为这个视频下载字幕，所有网站找到的字幕都会汇总输出
func (d Downloader) downloadSub4OneVideo(oneVideoFullPath string, suppliers []sub_supplier.ISupplier, i int) []sub_supplier.SubInfo {
	ontVideoRootPath := filepath.Dir(oneVideoFullPath)
	var outSUbInfos = make([]sub_supplier.SubInfo, 0)
	// 同时进行查询
	subInfosChannel := make(chan []sub_supplier.SubInfo)
	d.log.Infoln("DlSub Start", oneVideoFullPath)
	for _, supplier := range suppliers {
		supplier := supplier
		go func() {
			subInfos, err := d.downloadSub4OneSite(oneVideoFullPath, i, supplier, ontVideoRootPath)
			if err != nil {
				d.log.Error(err)
			}
			subInfosChannel <- subInfos
		}()
	}
	for i := 0; i < len(suppliers); i++ {
		v, ok := <-subInfosChannel
		if ok == true {
			outSUbInfos = append(outSUbInfos, v...)
		}
	}
	d.log.Infoln(i, "DlSub End", oneVideoFullPath)
	return outSUbInfos
}

// downloadSub4OneSite 在一个站点下载这个视频的字幕
func (d Downloader) downloadSub4OneSite(oneVideoFullPath string, i int, supplier sub_supplier.ISupplier, ontVideoRootPath string) ([]sub_supplier.SubInfo, error) {
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
	if d.reqParam.DebugMode == true {
		// 需要进行字幕文件的缓存
		// 把缓存的文件夹新建出来
		desFolderFullPath := path.Join(ontVideoRootPath, SubTmpFolderName)
		err = os.MkdirAll(desFolderFullPath, os.ModePerm)
		if err != nil {
			d.log.Error(err)
			return subInfos, nil
		}
		for x, info := range subInfos {
			desSubFileFullPath := path.Join(desFolderFullPath, supplier.GetSupplierName() + "_" + strconv.Itoa(x)+"_"+info.Name)
			err = utils.OutputFile(desSubFileFullPath, info.Data)
			if err != nil {
				d.log.Error(err)
				break
			}
		}
	}
	d.log.Infoln(i, supplier.GetSupplierName(), "End...")
	return subInfos, nil
}

// searchMatchedVideoFile 搜索符合后缀名的视频文件
func (d Downloader) searchMatchedVideoFile(dir string) ([]string, error) {

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
			oneList, _ := d.searchMatchedVideoFile(fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if d.isWantedVideoExtDef(curFile.Name()) == true {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// searchMatchedSubFile 搜索符合后缀名的视频文件
func (d Downloader) searchMatchedSubFile(dir string) ([]string, error) {

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
			if common.IsSubExtWanted(filepath.Ext(curFile.Name())) == true {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// isWantedVideoExtDef 后缀名是否符合规则
func (d Downloader) isWantedVideoExtDef(fileName string) bool {
	fileName = strings.ToLower(filepath.Ext(fileName))
	for _, s := range d.wantedExtList {
		if s == fileName {
			return true
		}
	}
	return false
}

// 返回的名称包含，那个网站下载的，这个网站中排名第几，文件名
func (d Downloader) getFrontNameAndOrgName(info sub_supplier.SubInfo) string {
	return "[" + info.FromWhere + "]" + strconv.FormatInt(info.TopN,10) +info.Name
}

const (
	VideoExtMp4 = ".mp4"
	VideoExtMkv = ".mkv"
	VideoExtRmvb = ".rmvb"
	VideoExtIso = ".iso"

	SubTmpFolderName = "subTmp"
)