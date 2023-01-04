package save_sub_helper

import (
	"os"
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/change_file_encode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/chs_cht_changer"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ifaces"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_timeline_fixer"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
	"github.com/sirupsen/logrus"
)

type SaveSubHelper struct {
	log                      *logrus.Logger
	SubFormatter             ifaces.ISubFormatter                         // 字幕格式化命名的实现
	subTimelineFixerHelperEx *sub_timeline_fixer.SubTimelineFixerHelperEx // 字幕时间轴校正
}

func NewSaveSubHelper(log *logrus.Logger, subFormatter ifaces.ISubFormatter, subTimelineFixerHelperEx *sub_timeline_fixer.SubTimelineFixerHelperEx) *SaveSubHelper {
	return &SaveSubHelper{log: log, SubFormatter: subFormatter, subTimelineFixerHelperEx: subTimelineFixerHelperEx}
}

// WriteSubFile2VideoPath 在前面需要进行语言的筛选、排序，这里仅仅是存储， extraSubPreName 这里传递是字幕的网站，有就认为是多字幕的存储。空就是单字幕，单字幕就可以setDefault
func (s *SaveSubHelper) WriteSubFile2VideoPath(videoFileFullPath string, finalSubFile subparser.FileInfo, extraSubPreName string, setDefault bool, skipExistFile bool) error {
	defer s.log.Infoln("----------------------------------")
	videoRootPath := filepath.Dir(videoFileFullPath)
	subNewName, subNewNameWithDefault, _ := s.SubFormatter.GenerateMixSubName(videoFileFullPath, finalSubFile.Ext, finalSubFile.Lang, extraSubPreName)

	desSubFullPath := filepath.Join(videoRootPath, subNewName)
	if setDefault == true {
		// 先判断没有 default 的字幕是否存在了，在的话，先删除，然后再写入
		if pkg.IsFile(desSubFullPath) == true {
			_ = os.Remove(desSubFullPath)
		}
		desSubFullPath = filepath.Join(videoRootPath, subNewNameWithDefault)
	}

	if skipExistFile == true {
		// 需要判断文件是否存在在，有则跳过
		if pkg.IsFile(desSubFullPath) == true {
			s.log.Infoln("OrgSubName:", finalSubFile.Name)
			s.log.Infoln("Sub Skip DownAt:", desSubFullPath)
			return nil
		}
	}
	// 最后写入字幕
	err := pkg.WriteFile(desSubFullPath, finalSubFile.Data)
	if err != nil {
		return err
	}
	s.log.Infoln("----------------------------------")
	s.log.Infoln("OrgSubName:", finalSubFile.Name)
	s.log.Infoln("SubDownAt:", desSubFullPath)

	// 然后还需要判断是否需要校正字幕的时间轴
	if settings.Get().AdvancedSettings.FixTimeLine == true {
		err = s.subTimelineFixerHelperEx.Process(videoFileFullPath, desSubFullPath)
		if err != nil {
			return err
		}
	}
	// 判断是否需要转换字幕的编码
	if settings.Get().ExperimentalFunction.AutoChangeSubEncode.Enable == true {
		s.log.Infoln("----------------------------------")
		s.log.Infoln("change_file_encode to", settings.Get().ExperimentalFunction.AutoChangeSubEncode.GetDesEncodeType())
		err = change_file_encode.Process(desSubFullPath, settings.Get().ExperimentalFunction.AutoChangeSubEncode.DesEncodeType)
		if err != nil {
			return err
		}
	}

	// 判断是否需要进行简繁互转
	// 一定得是 UTF-8 才能够执行简繁转换
	// 测试了先转 UTF-8 进行简繁转换然后再转 GBK，有些时候会出错，所以还是不支持这样先
	if settings.Get().ExperimentalFunction.AutoChangeSubEncode.Enable == true &&
		settings.Get().ExperimentalFunction.AutoChangeSubEncode.DesEncodeType == 0 &&
		settings.Get().ExperimentalFunction.ChsChtChanger.Enable == true {
		s.log.Infoln("----------------------------------")
		s.log.Infoln("chs_cht_changer to", settings.Get().ExperimentalFunction.ChsChtChanger.GetDesChineseLanguageTypeString())
		err = chs_cht_changer.Process(desSubFullPath, settings.Get().ExperimentalFunction.ChsChtChanger.DesChineseLanguageType)
		if err != nil {
			return err
		}
	}

	return nil
}
