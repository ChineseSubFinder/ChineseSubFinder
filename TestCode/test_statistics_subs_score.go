package TestCode

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/allanpk716/ChineseSubFinder/pkg/ffmpeg_helper"
	"github.com/prometheus/common/log"

	"github.com/allanpk716/ChineseSubFinder/pkg/types/subparser"
	"github.com/huandu/go-clone"

	"github.com/xuri/excelize/v2"

	"github.com/allanpk716/ChineseSubFinder/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/pkg/vad"
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
		log.Errorln("SetCellValue A Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("B%d", 1), "AudioScore")
	if err != nil {
		log.Errorln("SetCellValue B Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("C%d", 1), "AudioOffset")
	if err != nil {
		log.Errorln("SetCellValue C Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("D%d", 1), "SubScore")
	if err != nil {
		log.Errorln("SetCellValue D Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("E%d", 1), "SubOffset")
	if err != nil {
		log.Errorln("SetCellValue E Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("F%d", 1), "IsMatch")
	if err != nil {
		log.Errorln("SetCellValue F Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("G%d", 1), "VideoDuration")
	if err != nil {
		log.Errorln("SetCellValue G Header", err)
		return
	}
	err = f.SetCellValue(sheetName, fmt.Sprintf("H%d", 1), "TargetSubEndTime")
	if err != nil {
		log.Errorln("SetCellValue H Header", err)
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

			bok, matchResult, err := s.IsMatchBySubFile(ffmpegInfo, audioVADInfos, infoBase, path,
				sub_timeline_fixer.CompareConfig{
					MinScore:                      40000,
					OffsetRange:                   2,
					DialoguesDifferencePercentage: 0.1,
				})
			if err != nil {
				return nil
			}

			if bok == false {
				return nil
			}

			subCounter++
			err = f.SetCellValue(sheetName, fmt.Sprintf("A%d", subCounter+1), info.Name())
			if err != nil {
				log.Errorln("SetCellValue A", info.Name(), subCounter+1, err)
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("B%d", subCounter+1), matchResult.AudioCompareScore)
			if err != nil {
				log.Errorln("SetCellValue B", info.Name(), subCounter+1, err)
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("C%d", subCounter+1), matchResult.AudioCompareOffsetTime)
			if err != nil {
				log.Errorln("SetCellValue C", info.Name(), subCounter+1, err)
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("D%d", subCounter+1), matchResult.SubCompareScore)
			if err != nil {
				log.Errorln("SetCellValue D", info.Name(), subCounter+1, err)
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("E%d", subCounter+1), matchResult.SubCompareOffsetTime)
			if err != nil {
				log.Errorln("SetCellValue E", info.Name(), subCounter+1, err)
				return nil
			}
			iTrue := 0
			if bok == true {
				iTrue = 1
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("F%d", subCounter+1), iTrue)
			if err != nil {
				log.Errorln("SetCellValue F", info.Name(), subCounter+1, err)
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("G%d", subCounter+1), matchResult.VideoDuration)
			if err != nil {
				log.Errorln("SetCellValue G", info.Name(), subCounter+1, err)
				return nil
			}
			err = f.SetCellValue(sheetName, fmt.Sprintf("H%d", subCounter+1), matchResult.TargetSubEndTime)
			if err != nil {
				log.Errorln("SetCellValue H", info.Name(), subCounter+1, err)
				return nil
			}

			log.Infoln(subCounter, path, info.Size())

			return nil
		})
	if err != nil {
		log.Errorln("Walk", err)
		return
	}

	f.SetActiveSheet(newSheet)
	err = f.SaveAs(fmt.Sprintf("%s.xlsx", excelFileName))
	if err != nil {
		log.Errorln("SaveAs", err)
		return
	}

	log.Infoln("Done")
}
