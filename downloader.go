package ChineseSubFinder

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_parser"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/srt"
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
	defer func() {
		// 抉择完毕，需要清理缓存目录
		err := common.ClearTmpFolder()
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
			d.log.Error("oneVideoFullPath", "Download Sub Error",err)
			continue
		}
		// 得到目标视频文件的根目录
		videoRootPath := filepath.Dir(oneVideoFullPath)
		// -------------------------------------------------
		// 调试缓存，把下载好的字幕写到对应的视频目录下，方便调试
		if d.reqParam.DebugMode == true {
			err = d.copySubFile2DesFolder(videoRootPath, organizeSubFiles)
			if err != nil {
				d.log.Error(err)
			}
		}
		// -------------------------------------------------
		// TODO 这里先处理 Top1 的字幕，后续再考虑怎么觉得 Top N 选择哪一个，很可能选择每个网站 Top 1就行了，具体的过滤逻辑在其内部实现
		// 一个网站可能就算取了 Top1 字幕，也可能是返回一个压缩包，然后解压完就是多个字幕，所以
		var subInfoDict = make(map[string][]sub_parser.SubFileInfo)
		// 拿到现有的字幕列表，开始抉择
		// 先判断当前字幕是什么语言（如果是简体，还需要考虑，判断这个字幕是简体还是繁体）
		subParserHub := NewSubParserHub(ass.NewParser(), srt.NewParser())
		for _, oneSubFileFullPath := range organizeSubFiles {
			subFileInfo, err := subParserHub.DetermineFileTypeFromFile(oneSubFileFullPath)
			if err != nil {
				d.log.Error(err)
				continue
			}
			if subFileInfo == nil {
				// 说明这个字幕无法解析
				d.log.Warning(oneSubFileFullPath, "DetermineFileTypeFromFile is nill")
				continue
			}

			_, ok := subInfoDict[subFileInfo.FromWhereSite]
			if ok == true {
				// 添加
				subInfoDict[subFileInfo.FromWhereSite] = append(subInfoDict[subFileInfo.FromWhereSite], *subFileInfo)
			} else {
				// 新建
				subInfoDict[subFileInfo.FromWhereSite] = make([]sub_parser.SubFileInfo, 0)
				subInfoDict[subFileInfo.FromWhereSite] = append(subInfoDict[subFileInfo.FromWhereSite], *subFileInfo)
			}
		}
		// 优先级别暂定 zimuku -> subhd -> xunlei -> shooter
		foundOne := false
		var finalSubFile sub_parser.SubFileInfo
		// -----------------------------------------------------
		// TODO 需要重构，这些写的冲忙，太恶心了
		value, ok := subInfoDict["zimuku"]
		if ok == true {
			for _, info := range value {
				if common.HasChineseLang(info.Lang) == true {
					finalSubFile = info
					foundOne = true
					break
				}
			}
		}
		if foundOne {
			// 找到了
			err := d.writeSubFile2VideoPath(oneVideoFullPath, finalSubFile)
			if err != nil {
				d.log.Error("writeSubFile2VideoPath",err)
				// 不行继续
				foundOne = false
			} else {
				continue
			}
		}
		// -----------------------------------------------------
		value, ok = subInfoDict["subhd"]
		if ok == true {
			for _, info := range value {
				if common.HasChineseLang(info.Lang) == true {
					finalSubFile = info
					foundOne = true
					break
				}
			}
		}
		if foundOne {
			// 找到了
			err := d.writeSubFile2VideoPath(oneVideoFullPath, finalSubFile)
			if err != nil {
				d.log.Error("writeSubFile2VideoPath",err)
				// 不行继续
				foundOne = false
			} else {
				continue
			}
		}
		// -----------------------------------------------------
		value, ok = subInfoDict["xunlei"]
		if ok == true {
			for _, info := range value {
				if common.HasChineseLang(info.Lang) == true {
					finalSubFile = info
					foundOne = true
					break
				} else {
					continue
				}
			}
		}
		if foundOne {
			// 找到了
			err := d.writeSubFile2VideoPath(oneVideoFullPath, finalSubFile)
			if err != nil {
				d.log.Error("writeSubFile2VideoPath",err)
				// 不行继续
				foundOne = false
			}
		}
		// -----------------------------------------------------
		value, ok = subInfoDict["shooter"]
		if ok == true {
			for _, info := range value {
				if common.HasChineseLang(info.Lang) == true {
					finalSubFile = info
					foundOne = true
					break
				} else {
					continue
				}
			}
		}
		if foundOne {
			// 找到了
			err := d.writeSubFile2VideoPath(oneVideoFullPath, finalSubFile)
			if err != nil {
				d.log.Error("writeSubFile2VideoPath",err)
				// 不行继续
				foundOne = false
			}
		}
		// -----------------------------------------------------
	}
	return nil
}

func (d Downloader) writeSubFile2VideoPath(videoFileFullPath string, finalSubFile sub_parser.SubFileInfo) error {
	videoRootPath := filepath.Dir(videoFileFullPath)
	// 需要符合 emby 的格式要求，在后缀名前面
	const emby_zh = ".zh"
	const emby_en = ".en"
	//TODO 日文 韩文 emby 字幕格式要求，瞎猜的，有需要再改（目标应该是中文字幕查找，所以···应该不需要）
	const emby_jp = ".jp"
	const emby_kr = ".kr"
	lan := ""
	if common.HasChineseLang(finalSubFile.Lang) == true {
		lan = emby_zh
	} else if finalSubFile.Lang == common.English {
		lan = emby_en
	}
	// 构建视频文件加 emby 的字幕预研要求名称
	videoFileNameWithOutExt := strings.ReplaceAll(filepath.Base(videoFileFullPath),
		filepath.Ext(videoFileFullPath), "")
	subNewName := videoFileNameWithOutExt + lan + finalSubFile.Ext
	desSubFullPath := path.Join(videoRootPath, subNewName)
	// 最后写入字幕
	err := utils.OutputFile(desSubFullPath, finalSubFile.Data)
	if err != nil {
		return err
	}
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
	desFolderFullPath := path.Join(desFolder, SubTmpFolderName)
	err := os.MkdirAll(desFolderFullPath, os.ModePerm)
	if err != nil {
		return err
	}
	// 复制下载在 tmp 文件夹中的字幕文件到视频文件夹下面
	for _, subFile := range subFiles {
		newFn := path.Join(desFolderFullPath, filepath.Base(subFile))
		_, err = common.CopyFile(newFn, subFile)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	VideoExtMp4 = ".mp4"
	VideoExtMkv = ".mkv"
	VideoExtRmvb = ".rmvb"
	VideoExtIso = ".iso"

	SubTmpFolderName = "subtmp"
)