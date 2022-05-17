package sub_formatter

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	movieHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/movie_helper"
	seriesHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter/normal"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	interCommon "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
)

type SubFormatChanger struct {
	log            *logrus.Logger
	movieRootDirs  []string
	seriesRootDirs []string
	formatter      map[string]ifaces.ISubFormatter
}

func NewSubFormatChanger(log *logrus.Logger, movieRootDirs []string, seriesRootDirs []string) *SubFormatChanger {

	formatter := SubFormatChanger{movieRootDirs: movieRootDirs, seriesRootDirs: seriesRootDirs}
	formatter.log = log
	// TODO 如果字幕格式新增了实现，这里也需要添加对应的实例
	// 初始化支持的 formatter
	// normal
	formatter.formatter = make(map[string]ifaces.ISubFormatter)
	normalM := normal.NewFormatter(log)
	formatter.formatter[normalM.GetFormatterName()] = normalM
	// emby
	embyM := emby.NewFormatter()
	formatter.formatter[embyM.GetFormatterName()] = embyM
	return &formatter
}

// AutoDetectThenChangeTo 自动检测字幕的命名格式，然后转换到目标的 formatter 上
func (s *SubFormatChanger) AutoDetectThenChangeTo(desFormatter common.FormatterName) (RenameResults, error) {

	outStruct := RenameResults{}
	outStruct.RenamedFiles = make(map[string]int)
	outStruct.ErrFiles = make(map[string]int)

	for i, dir := range s.movieRootDirs {
		s.log.Infoln("AutoDetectThenChangeTo Movie Index", i, dir, "Start")

		err := s.autoDetectMovieThenChangeTo(&outStruct, desFormatter, dir)
		if err != nil {
			s.log.Infoln("AutoDetectThenChangeTo Movie Index", i, dir, "End")
			return RenameResults{}, err
		}

		s.log.Infoln("AutoDetectThenChangeTo Movie Index", i, dir, "Start")
	}

	for i, dir := range s.seriesRootDirs {
		s.log.Infoln("AutoDetectThenChangeTo Series Index", i, dir, "Start")

		err := s.autoDetectMovieThenChangeTo(&outStruct, desFormatter, dir)
		if err != nil {
			s.log.Infoln("AutoDetectThenChangeTo Series Index", i, dir, "End")
			return RenameResults{}, err
		}

		s.log.Infoln("AutoDetectThenChangeTo Series Index", i, dir, "Start")
	}

	return outStruct, nil
}

func (s *SubFormatChanger) autoDetectMovieThenChangeTo(outStruct *RenameResults, desFormatter common.FormatterName, movieRootDir string) error {

	var err error
	if my_util.IsDir(movieRootDir) == false {
		return errors.New("movieRootDir path not exist: " + movieRootDir)
	}
	// 先找出有那些电影文件夹和连续剧文件夹
	var movieFullPathList = make([]string, 0)
	movieFullPathList, err = my_util.SearchMatchedVideoFile(s.log, movieRootDir)
	// fmt.Println("No. of Movies: ", len(movieFullPathList), "  dir:  ", s.movieRootDir)
	if err != nil {
		return err
	}
	// 搜索所有的字幕，找到相关的字幕进行修改
	for _, one := range movieFullPathList {

		// 需要判断这个视频根目录是否有 .ignore 文件，有也跳过
		if my_util.IsFile(filepath.Join(filepath.Dir(one), interCommon.Ignore)) == true {
			s.log.Infoln("Found", interCommon.Ignore, "Skip", one)
			// 跳过下载字幕
			continue
		}

		found := false
		var fitMovieNameSubList = make([]string, 0)
		found, _, fitMovieNameSubList, err = movieHelper.MovieHasChineseSub(s.log, one)
		if err != nil || found == false {
			continue
		}
		// 判断是否是符合要求
		for _, fitSubName := range fitMovieNameSubList {
			s.autoDetectAndChange(outStruct, fitSubName, desFormatter)
		}
	}

	return nil
}

func (s *SubFormatChanger) autoDetectMSeriesThenChangeTo(outStruct *RenameResults, desFormatter common.FormatterName, seriesRootDir string) error {

	var err error
	if my_util.IsDir(seriesRootDir) == false {
		return errors.New("seriesRootDir path not exist: " + seriesRootDir)
	}
	// 先找出有那些电影文件夹和连续剧文件夹
	seriesDirList, err := seriesHelper.GetSeriesList(s.log, seriesRootDir)
	if err != nil {
		return err
	}
	// 连续剧
	var seriesSubFiles = make([]string, 0)
	for _, oneSeriesDir := range seriesDirList {

		// 需要判断这个视频根目录是否有 .ignore 文件，有也跳过
		if my_util.IsFile(filepath.Join(oneSeriesDir, interCommon.Ignore)) == true {
			s.log.Infoln("Found", interCommon.Ignore, "Skip", oneSeriesDir)
			// 跳过下载字幕
			continue
		}

		seriesSubFiles, err = sub_helper.SearchMatchedSubFileByDir(s.log, oneSeriesDir)
		if err != nil {
			return err
		}
		// 判断是否是符合要求
		for _, fitSubName := range seriesSubFiles {
			s.autoDetectAndChange(outStruct, fitSubName, desFormatter)
		}
	}

	return nil
}

// autoDetectAndChange 自动检测命名格式，然后修改至目标的命名格式
func (s *SubFormatChanger) autoDetectAndChange(outStruct *RenameResults, fitSubName string, desFormatter common.FormatterName) {

	for _, formatter := range s.formatter {

		// true,   ,  ./../../TestData/sub_format_changer/test/movie_org_emby/AAA/AAA.chinese(简英,subhd).ass, 未知语言,   ,
		bok, fileNameWithOutExt, subExt, subLang, extraSubPreName := formatter.IsMatchThisFormat(fitSubName)
		if bok == false {
			continue
		}
		// 如果检测到的格式和目标要转换到的格式是一个，那么就跳过
		if common.FormatterName(formatter.GetFormatterFormatterName()) == desFormatter {
			return
		}
		// 这里得到的 subExt 可能是 .ass or .default.ass or .forced.ass
		// 需要进行剔除，因为后续的 GenerateMixSubName 会自动生成对应的附加后缀名
		// 转换格式后，需要保留之前的 default 或者 forced
		findDefault := false
		findForce := false
		if strings.Contains(subExt, subparser.Sub_Ext_Mark_Default) == true {
			subExt = strings.Replace(subExt, subparser.Sub_Ext_Mark_Default, "", -1)
			findDefault = true
		}
		if strings.Contains(subExt, subparser.Sub_Ext_Mark_Forced) == true {
			subExt = strings.Replace(subExt, subparser.Sub_Ext_Mark_Forced, "", -1)
			findForce = true
		}
		// 通过传入的目标格式化 formatter 的名称去调用
		newSubFileName := ""
		newName, newDefaultName, newForcedName := s.formatter[fmt.Sprintf("%s", desFormatter)].
			GenerateMixSubNameBase(fileNameWithOutExt, subExt, subLang, extraSubPreName)

		// fmt.Println(fmt.Sprintf("%s", desFormatter))
		// fmt.Println("newName       : " + newName)
		// fmt.Println("newDefaultName: " + newDefaultName)
		// fmt.Println("newForcedName : " + newForcedName)
		if findDefault == false && findForce == false {
			// 使用没得额外 Default 或者 Forced 的名称即可
			newSubFileName = newName
		} else if findDefault == true {
			newSubFileName = newDefaultName
		} else if findForce == true {
			newSubFileName = newForcedName
		}
		if newSubFileName == "" {
			continue
		}
		// 确认改格式
		err := os.Rename(fitSubName, newSubFileName)
		if err != nil {
			tmpName := my_util.FixWindowPathBackSlash(fitSubName)
			outStruct.ErrFiles[tmpName] += 1
			continue
		} else {
			tmpName := my_util.FixWindowPathBackSlash(newSubFileName)
			outStruct.RenamedFiles[tmpName] += 1
		}
	}
}

type RenameResults struct {
	RenamedFiles map[string]int
	ErrFiles     map[string]int
}

// GetSubFormatter 选择字幕命名格式化的实例
func GetSubFormatter(log *logrus.Logger, subNameFormatter int) ifaces.ISubFormatter {
	var subFormatter ifaces.ISubFormatter
	switch subNameFormatter {
	case int(common.Emby):
		{
			subFormatter = emby.NewFormatter()
			break
		}
	case int(common.Normal):
		{
			subFormatter = normal.NewFormatter(log)
			break
		}
	default:
		{
			subFormatter = emby.NewFormatter()
			break
		}
	}

	return subFormatter
}

// SubFormatChangerProcess 执行 SubFormatChanger 逻辑，并且更新数据库缓存
func SubFormatChangerProcess(log *logrus.Logger, movieRootDirs []string, seriesRootDirs []string, nowDesFormatter common.FormatterName) (RenameResults, error) {
	var subFormatRec models.SubFormatRec
	re := dao.GetDb().First(&subFormatRec)
	if re == nil {
		return RenameResults{}, errors.New(fmt.Sprintf("SubFormatChangerProcess dao.GetDb().First return nil"))
	}
	if re.Error != nil {
		if re.Error != gorm.ErrRecordNotFound {
			return RenameResults{}, errors.New(fmt.Sprintf("SubFormatChangerProcess dao.GetDb().First, %v", re.Error))
		}
	}
	subFormatChanger := NewSubFormatChanger(log, movieRootDirs, seriesRootDirs)
	// 理论上有且仅有一条记录
	if subFormatRec.Done == false {
		// 没有找到，认为是第一次执行
		renameResults, err := subFormatChanger.AutoDetectThenChangeTo(nowDesFormatter)
		if err != nil {
			return renameResults, err
		}

		// 需要记录到数据库中
		oneSubFormatter := models.SubFormatRec{FormatName: int(nowDesFormatter), Done: true}
		re = dao.GetDb().Create(&oneSubFormatter)
		if re == nil {
			return RenameResults{}, errors.New(fmt.Sprintf("SubFormatChangerProcess dao.GetDb().Create return nil"))
		}
		if re.Error != nil {
			return RenameResults{}, errors.New(fmt.Sprintf("SubFormatChangerProcess dao.GetDb().Create, %v", re.Error))
		}
		return renameResults, nil
	} else {
		// 找到了，需要判断上一次执行的目标 formatter 是啥，如果这次的目标 formatter 不一样则执行
		// 如果是一样的则跳过
		if common.FormatterName(subFormatRec.FormatName) == nowDesFormatter {
			log.Infoln("DesSubFormatter == LateTimeSubFormatter then skip process")
			return RenameResults{}, nil
		}
		// 执行更改
		renameResults, err := subFormatChanger.AutoDetectThenChangeTo(nowDesFormatter)
		if err != nil {
			return renameResults, err
		}
		// 更新数据库
		subFormatRec.FormatName = int(nowDesFormatter)
		subFormatRec.Done = true
		re = dao.GetDb().Save(subFormatRec)
		if re == nil {
			return RenameResults{}, errors.New(fmt.Sprintf("SubFormatChangerProcess dao.GetDb().Save return nil"))
		}
		if re.Error != nil {
			return RenameResults{}, errors.New(fmt.Sprintf("SubFormatChangerProcess dao.GetDb().Save, %v", re.Error))
		}
		return renameResults, nil
	}
}
