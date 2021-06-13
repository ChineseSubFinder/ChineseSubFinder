package ChineseSubFinder

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/mark_system"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/zimuku"
	"github.com/go-rod/rod/lib/utils"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Downloader struct {
	reqParam      common.ReqParam
	log           *logrus.Logger
	topic         int                        // 最多能够下载 Top 几的字幕，每一个网站
	wantedExtList []string                   // 人工确认的需要监控的视频后缀名
	defExtList    []string                   // 内置支持的视频后缀名列表
	mk            *mark_system.MarkingSystem // MarkingSystem
}

func NewDownloader(_reqParam ...common.ReqParam) *Downloader {

	var downloader Downloader
	downloader.log = model.GetLogger()
	downloader.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		downloader.reqParam = _reqParam[0]
		if downloader.reqParam.Topic > 0 && downloader.reqParam.Topic != downloader.topic {
			downloader.topic = downloader.reqParam.Topic
		}
	}
	downloader.defExtList = make([]string, 0)
	downloader.defExtList = append(downloader.defExtList, common.VideoExtMp4)
	downloader.defExtList = append(downloader.defExtList, common.VideoExtMkv)
	downloader.defExtList = append(downloader.defExtList, common.VideoExtRmvb)
	downloader.defExtList = append(downloader.defExtList, common.VideoExtIso)

	var sitesSequence = make([]string, 0)
	// TODO 这里写固定了抉择字幕的顺序
	sitesSequence = append(sitesSequence, common.SubSiteZiMuKu)
	sitesSequence = append(sitesSequence, common.SubSiteSubHd)
	sitesSequence = append(sitesSequence, common.SubSiteXunLei)
	sitesSequence = append(sitesSequence, common.SubSiteShooter)
	downloader.mk = mark_system.NewMarkingSystem(sitesSequence)

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
	defer func() {
		// 抉择完毕，需要清理缓存目录
		err := model.ClearTmpFolder()
		if err != nil {
			d.log.Error(err)
		}
	}()
	nowVideoList, err := d.searchMatchedVideoFile(dir)
	if err != nil {
		return err
	}
	// 构建每个字幕站点下载者的实例
	subSupplierHub := sub_supplier.NewSubSupplierHub(shooter.NewSupplier(d.reqParam),
									subhd.NewSupplier(d.reqParam),
									xunlei.NewSupplier(d.reqParam),
									zimuku.NewSupplier(d.reqParam),
	)
	// TODO 后续再改为每个视频以上的流程都是一个 channel 来做，并且需要控制在一个并发量之下（很可能没必要，毕竟要在弱鸡机器上挂机用的）
	// 一个视频文件同时多个站点查询，阻塞完毕后，在进行下一个
	for i, oneVideoFullPath := range nowVideoList {
		// 字幕都下载缓存好了，需要抉择存哪一个，优先选择中文双语的，然后到中文
		organizeSubFiles, err := subSupplierHub.DownloadSub(oneVideoFullPath, i)
		if err != nil {
			d.log.Errorln("subSupplierHub.DownloadSub", oneVideoFullPath ,err)
			continue
		}
		// 得到目标视频文件的根目录
		videoRootPath := filepath.Dir(oneVideoFullPath)
		// -------------------------------------------------
		// 调试缓存，把下载好的字幕写到对应的视频目录下，方便调试
		if d.reqParam.DebugMode == true {
			err = d.copySubFile2DesFolder(videoRootPath, organizeSubFiles)
			if err != nil {
				d.log.Errorln("copySubFile2DesFolder", err)
			}
		}
		// -------------------------------------------------
		if d.reqParam.SaveMultiSub == false {
			// 选择最优的一个字幕
			var finalSubFile *common.SubParserFileInfo
			finalSubFile = d.mk.SelectOneSubFile(organizeSubFiles)
			if finalSubFile == nil {
				d.log.Warnln("Found", len(organizeSubFiles), " subtitles but not one fit:", oneVideoFullPath)
				continue
			}
			// 找到了，写入文件
			err = d.writeSubFile2VideoPath(oneVideoFullPath, *finalSubFile, "")
			if err != nil {
				d.log.Errorln("SaveMultiSub:", d.reqParam.SaveMultiSub ,"writeSubFile2VideoPath:", err)
				continue
			}
		} else {
			// 每个网站 Top1 的字幕
			siteNames, finalSubFiles := d.mk.SelectEachSiteTop1SubFile(organizeSubFiles)
			if len(siteNames) < 0 {
				d.log.Warnln("SelectEachSiteTop1SubFile found none sub file")
				continue
			}

			for i, file := range finalSubFiles {
				err = d.writeSubFile2VideoPath(oneVideoFullPath, file, siteNames[i])
				if err != nil {
					d.log.Errorln("SaveMultiSub:", d.reqParam.SaveMultiSub ,"writeSubFile2VideoPath:", err)
					continue
				}
			}
		}
		// -----------------------------------------------------
	}
	return nil
}

// 在前面需要进行语言的筛选、排序，这里仅仅是存储
func (d Downloader) writeSubFile2VideoPath(videoFileFullPath string, finalSubFile common.SubParserFileInfo, extraSubPreName string) error {
	videoRootPath := filepath.Dir(videoFileFullPath)
	embyLanExtName := model.Lang2EmbyName(finalSubFile.Lang)
	// 构建视频文件加 emby 的字幕预研要求名称
	videoFileNameWithOutExt := strings.ReplaceAll(filepath.Base(videoFileFullPath),
		filepath.Ext(videoFileFullPath), "")
	if extraSubPreName != "" {
		extraSubPreName = "[" + extraSubPreName +"]"
	}
	subNewName := videoFileNameWithOutExt + embyLanExtName + extraSubPreName + finalSubFile.Ext
	desSubFullPath := path.Join(videoRootPath, subNewName)
	// 最后写入字幕
	err := utils.OutputFile(desSubFullPath, finalSubFile.Data)
	if err != nil {
		return err
	}
	d.log.Infoln("OrgSubName:", finalSubFile.Name)
	d.log.Infoln("SubDownAt:", desSubFullPath)
	return nil
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

func (d Downloader) copySubFile2DesFolder(desFolder string, subFiles []string) error {

	// 需要进行字幕文件的缓存
	// 把缓存的文件夹新建出来
	desFolderFullPath := path.Join(desFolder, common.SubTmpFolderName)
	err := os.MkdirAll(desFolderFullPath, os.ModePerm)
	if err != nil {
		return err
	}
	// 复制下载在 tmp 文件夹中的字幕文件到视频文件夹下面
	for _, subFile := range subFiles {
		newFn := path.Join(desFolderFullPath, filepath.Base(subFile))
		_, err = model.CopyFile(newFn, subFile)
		if err != nil {
			return err
		}
	}

	return nil
}

