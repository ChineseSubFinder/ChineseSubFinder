package TestCode

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	common2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/task_control"

	"github.com/sirupsen/logrus"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ffmpeg_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
	"github.com/huandu/go-clone"

	"github.com/xuri/excelize/v2"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_timeline_fixer"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/vad"
)

func statistics_subs_score_one(baseAudioFileFPath, baseSubFileFPath, srcSubFileFPath string) {

	audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
		FileFullPath: baseAudioFileFPath,
		SampleRate:   16000,
		BitDepth:     16,
	}, true)
	if err != nil {
		return
	}

	subParserHub := sub_parser_hub.NewSubParserHub(
		log_helper.GetLogger4Tester(),
		ass.NewParser(log_helper.GetLogger4Tester()),
		srt.NewParser(log_helper.GetLogger4Tester()),
	)
	bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(baseSubFileFPath)
	if err != nil {
		return
	}
	if bFind == false {
		return
	}

	bFind, srcBase, err := subParserHub.DetermineFileTypeFromFile(srcSubFileFPath)
	if err != nil {
		return
	}
	if bFind == false {
		return
	}

	s := sub_timeline_fixer.NewSubTimelineFixerHelperEx(log_helper.GetLogger4Tester(), *settings.NewTimelineFixerSettings())
	// path X:\电影\21座桥 (2019)\21座桥 (2019) 720p AAC.chinese(简,subhd).ass
	// 音频处理
	cloneSrcBase := clone.Clone(srcBase).(*subparser.FileInfo)
	bok, _, pipeResultAudio, err := s.ProcessByAudioVAD(audioVADInfos, cloneSrcBase)
	if err != nil {
		return
	}
	if bok == false {
		return
	}
	// 字幕处理
	cloneSrcBase = clone.Clone(srcBase).(*subparser.FileInfo)
	bok, _, pipeResultSub, err := s.ProcessBySubFileInfo(infoBase, cloneSrcBase)
	if err != nil {
		return
	}
	if bok == false {
		return
	}

	println(fmt.Sprintf("Audio Score: %f  Offset:%f\n", pipeResultAudio.Score, pipeResultAudio.GetOffsetTime()))
	println(fmt.Sprintf("Sub Score: %f  Offset:%f\n", pipeResultSub.Score, pipeResultSub.GetOffsetTime()))
}

func statistics_subs_score(baseAudioFileFPath, baseSubFileFPath, subSearchRootPath string) {

	f := excelize.NewFile()
	// Create a new sheet.
	sheetName := filepath.Base(subSearchRootPath)
	newSheet := f.NewSheet(sheetName)
	err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", 1), "SubFPath")
	if err != nil {
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("B%d", 1), "AudioScore")
	if err != nil {
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("C%d", 1), "AudioOffset")
	if err != nil {
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("D%d", 1), "SubScore")
	if err != nil {
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("E%d", 1), "SubOffset")
	if err != nil {
		return
	}

	audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
		FileFullPath: baseAudioFileFPath,
		SampleRate:   16000,
		BitDepth:     16,
	}, true)
	if err != nil {
		return
	}

	subParserHub := sub_parser_hub.NewSubParserHub(
		log_helper.GetLogger4Tester(),
		ass.NewParser(log_helper.GetLogger4Tester()),
		srt.NewParser(log_helper.GetLogger4Tester()),
	)
	bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(baseSubFileFPath)
	if err != nil {
		return
	}
	if bFind == false {
		return
	}

	subCounter := 1
	err = filepath.Walk(subSearchRootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() == true {
				return nil
			}
			if sub_parser_hub.IsSubExtWanted(info.Name()) == false {
				return nil
			}

			bFind, srcBase, err := subParserHub.DetermineFileTypeFromFile(path)
			if err != nil {
				return nil
			}
			if bFind == false {
				return nil
			}

			s := sub_timeline_fixer.NewSubTimelineFixerHelperEx(log_helper.GetLogger4Tester(), *settings.NewTimelineFixerSettings())
			// path X:\电影\21座桥 (2019)\21座桥 (2019) 720p AAC.chinese(简,subhd).ass
			// 音频处理
			cloneSrcBase := clone.Clone(srcBase).(*subparser.FileInfo)
			bok, _, pipeResultAudio, err := s.ProcessByAudioVAD(audioVADInfos, cloneSrcBase)
			if err != nil {
				return nil
			}
			if bok == false {
				return nil
			}
			// 字幕处理
			cloneSrcBase = clone.Clone(srcBase).(*subparser.FileInfo)
			bok, _, pipeResultSub, err := s.ProcessBySubFileInfo(infoBase, cloneSrcBase)
			if err != nil {
				return nil
			}
			if bok == false {
				return nil
			}

			subCounter++
			err = f.SetCellValue(sheetName, fmt.Sprintf("A%d", subCounter+1), info.Name())
			if err != nil {
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("B%d", subCounter+1), pipeResultAudio.Score)
			if err != nil {
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("C%d", subCounter+1), pipeResultAudio.GetOffsetTime())
			if err != nil {
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("D%d", subCounter+1), pipeResultSub.Score)
			if err != nil {
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("E%d", subCounter+1), pipeResultSub.GetOffsetTime())
			if err != nil {
				return nil
			}
			fmt.Println(subCounter, path, info.Size())

			return nil
		})
	if err != nil {
		fmt.Println("Walk", err)
		return
	}

	f.SetActiveSheet(newSheet)
	err = f.SaveAs(fmt.Sprintf("%s.xlsx", filepath.Dir(baseSubFileFPath)))
	if err != nil {
		fmt.Println("SaveAs", err)
		return
	}
}

func statistics_subs_score_is_match(
	logger *logrus.Logger,
	s *sub_timeline_fixer.SubTimelineFixerHelperEx,
	ffmpegInfo *ffmpeg_helper.FFMPEGInfo,
	audioVADInfos []vad.VADInfo, infoBase *subparser.FileInfo,
	subSearchRootPath, excelFileName string) {

	var err error
	f := excelize.NewFile()
	// Create a new sheet.
	sheetName := filepath.Base(subSearchRootPath)
	newSheet := f.NewSheet(sheetName)
	err = f.SetCellValue(sheetName, fmt.Sprintf("A%d", 1), "SubFPath")
	if err != nil {
		logger.Errorln("SetCellValue A Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("B%d", 1), "AudioScore")
	if err != nil {
		logger.Errorln("SetCellValue B Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("C%d", 1), "AudioOffset")
	if err != nil {
		logger.Errorln("SetCellValue C Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("D%d", 1), "SubScore")
	if err != nil {
		logger.Errorln("SetCellValue D Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("E%d", 1), "SubOffset")
	if err != nil {
		logger.Errorln("SetCellValue E Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("F%d", 1), "IsMatch")
	if err != nil {
		logger.Errorln("SetCellValue F Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("G%d", 1), "VideoDuration")
	if err != nil {
		logger.Errorln("SetCellValue G Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("H%d", 1), "TargetSubEndTime")
	if err != nil {
		logger.Errorln("SetCellValue H Header", err)
		return
	}
	// --------------------------------------------------
	// 并发控制
	var taskControl *task_control.TaskControl
	taskControl, err = task_control.NewTaskControl(6, logger)
	if err != nil {
		logger.Errorln("NewTaskControl", err)
		return
	}
	taskControl.SetCtxProcessFunc("ScanSubPlayedPool", dealOne, common2.ScanPlayedSubTimeOut)
	// --------------------------------------------------
	subCounter = 1
	err = filepath.Walk(subSearchRootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() == true {
				return nil
			}
			if sub_parser_hub.IsSubExtWanted(info.Name()) == false {
				return nil
			}

			// 并发控制
			err = taskControl.Invoke(&task_control.TaskData{
				Index: 0,
				Count: 0,
				DataEx: ExcelInputData{
					Logger:            logger,
					F:                 f,
					S:                 s,
					FfmpegInfo:        ffmpegInfo,
					AudioVADInfos:     audioVADInfos,
					InfoBase:          infoBase,
					NowTargetSubFPath: path,
					SubFileName:       info.Name(),
					SheetName:         sheetName,
				},
			})
			if err != nil {
				logger.Errorln("Invoke", err)
			}

			return nil
		})
	if err != nil {
		logger.Errorln("Walk", err)
		return
	}

	taskControl.Hold()

	f.SetActiveSheet(newSheet)
	err = f.SaveAs(fmt.Sprintf("%s.xlsx", excelFileName))
	if err != nil {
		logger.Errorln("SaveAs", err)
		return
	}

	logger.Infoln("Done")
}

func dealOne(ctx context.Context, inData interface{}) error {

	taskData := inData.(*task_control.TaskData)
	excelInputData := taskData.DataEx.(ExcelInputData)

	bok, matchResult, err := excelInputData.S.IsMatchBySubFile(
		excelInputData.FfmpegInfo,
		excelInputData.AudioVADInfos,
		excelInputData.InfoBase,
		excelInputData.NowTargetSubFPath,
		sub_timeline_fixer.CompareConfig{
			MinScore:                      40000,
			OffsetRange:                   2,
			DialoguesDifferencePercentage: 0.25,
		})
	if err != nil {
		return nil
	}

	if bok == false && matchResult == nil {
		return nil
	}
	counterLock.Lock()
	defer counterLock.Unlock()

	subCounter++

	err = excelInputData.F.SetCellValue(excelInputData.SheetName, fmt.Sprintf("A%d", subCounter+1), excelInputData.SubFileName)
	if err != nil {
		excelInputData.Logger.Errorln("SetCellValue A", excelInputData.SubFileName, subCounter+1, err)
		return nil
	}
	err = excelInputData.F.SetCellValue(excelInputData.SheetName, fmt.Sprintf("B%d", subCounter+1), matchResult.AudioCompareScore)
	if err != nil {
		excelInputData.Logger.Errorln("SetCellValue B", excelInputData.SubFileName, subCounter+1, err)
		return nil
	}
	err = excelInputData.F.SetCellValue(excelInputData.SheetName, fmt.Sprintf("C%d", subCounter+1), matchResult.AudioCompareOffsetTime)
	if err != nil {
		excelInputData.Logger.Errorln("SetCellValue C", excelInputData.SubFileName, subCounter+1, err)
		return nil
	}
	err = excelInputData.F.SetCellValue(excelInputData.SheetName, fmt.Sprintf("D%d", subCounter+1), matchResult.SubCompareScore)
	if err != nil {
		excelInputData.Logger.Errorln("SetCellValue D", excelInputData.SubFileName, subCounter+1, err)
		return nil
	}
	err = excelInputData.F.SetCellValue(excelInputData.SheetName, fmt.Sprintf("E%d", subCounter+1), matchResult.SubCompareOffsetTime)
	if err != nil {
		excelInputData.Logger.Errorln("SetCellValue E", excelInputData.SubFileName, subCounter+1, err)
		return nil
	}
	iTrue := 0
	if bok == true {
		iTrue = 1
	}
	err = excelInputData.F.SetCellValue(excelInputData.SheetName, fmt.Sprintf("F%d", subCounter+1), iTrue)
	if err != nil {
		excelInputData.Logger.Errorln("SetCellValue F", excelInputData.SubFileName, subCounter+1, err)
		return nil
	}
	err = excelInputData.F.SetCellValue(excelInputData.SheetName, fmt.Sprintf("G%d", subCounter+1), matchResult.VideoDuration)
	if err != nil {
		excelInputData.Logger.Errorln("SetCellValue G", excelInputData.SubFileName, subCounter+1, err)
		return nil
	}
	err = excelInputData.F.SetCellValue(excelInputData.SheetName, fmt.Sprintf("H%d", subCounter+1), matchResult.TargetSubEndTime)
	if err != nil {
		excelInputData.Logger.Errorln("SetCellValue H", excelInputData.SubFileName, subCounter+1, err)
		return nil
	}

	excelInputData.Logger.Infoln(subCounter, excelInputData.NowTargetSubFPath)

	return nil
}

var counterLock sync.Mutex
var subCounter int

type ExcelInputData struct {
	Logger            *logrus.Logger
	F                 *excelize.File
	S                 *sub_timeline_fixer.SubTimelineFixerHelperEx
	FfmpegInfo        *ffmpeg_helper.FFMPEGInfo
	AudioVADInfos     []vad.VADInfo
	InfoBase          *subparser.FileInfo
	NowTargetSubFPath string
	SubFileName       string
	SheetName         string
}

type ExcelMathResult struct {
	Index       int
	Name        string
	MatchResult *sub_timeline_fixer.MatchResult
}
