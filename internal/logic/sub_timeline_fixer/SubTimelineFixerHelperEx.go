package sub_timeline_fixer

import (
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/ffmpeg_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/sirupsen/logrus"
	"math"
	"os"
)

type SubTimelineFixerHelperEx struct {
	log                 *logrus.Logger
	ffmpegHelper        *ffmpeg_helper.FFMPEGHelper
	subParserHub        *sub_parser_hub.SubParserHub
	timelineFixPipeLine *sub_timeline_fixer.Pipeline
	fixerConfig         settings.TimelineFixerSettings
	needDownloadFFMPeg  bool
}

func NewSubTimelineFixerHelperEx(log *logrus.Logger, fixerConfig settings.TimelineFixerSettings) *SubTimelineFixerHelperEx {

	fixerConfig.Check()

	return &SubTimelineFixerHelperEx{
		log:                 log,
		ffmpegHelper:        ffmpeg_helper.NewFFMPEGHelper(log),
		subParserHub:        sub_parser_hub.NewSubParserHub(log, ass.NewParser(log), srt.NewParser(log)),
		timelineFixPipeLine: sub_timeline_fixer.NewPipeline(fixerConfig.MaxOffsetTime),
		fixerConfig:         fixerConfig,
		needDownloadFFMPeg:  false,
	}
}

// Check 是否安装了 ffmpeg 和 ffprobe
func (s *SubTimelineFixerHelperEx) Check() bool {
	version, err := s.ffmpegHelper.Version()
	if err != nil {
		s.needDownloadFFMPeg = false
		s.log.Errorln("Need Install ffmpeg and ffprobe !")
		return false
	}
	s.needDownloadFFMPeg = true
	s.log.Infoln(version)
	return true
}

func (s SubTimelineFixerHelperEx) Process(videoFileFullPath, srcSubFPath string) error {

	if s.needDownloadFFMPeg == false {
		s.log.Errorln("Need Install ffmpeg and ffprobe, Can't Do TimeLine Fix")
		return nil
	}

	var infoSrc *subparser.FileInfo
	var pipeResultMax sub_timeline_fixer.PipeResult
	bProcess := false
	bok := false
	var ffmpegInfo *ffmpeg_helper.FFMPEGInfo
	var err error
	// 先尝试获取内置字幕的信息
	bok, ffmpegInfo, err = s.ffmpegHelper.GetFFMPEGInfo(videoFileFullPath, ffmpeg_helper.Subtitle)
	if err != nil {
		return err
	}
	if bok == false {
		return errors.New("SubTimelineFixerHelperEx.Process.GetFFMPEGInfo = false Subtitle -- " + videoFileFullPath)
	}

	// 这个需要提前考虑，如果只有一个内置的字幕，且这个字幕的大小小于 2kb，那么认为这个字幕是有问题的，就直接切换到 audio 校正
	oneSubAndIsError := false
	if len(ffmpegInfo.SubtitleInfoList) == 1 {
		fi, err := os.Stat(ffmpegInfo.SubtitleInfoList[0].FullPath)
		if err != nil {
			oneSubAndIsError = true
		} else {
			if fi.Size() <= 2048 {
				oneSubAndIsError = true
			}
		}
	}
	// 内置的字幕，这里只列举一种格式出来，其实会有一个字幕的 srt 和 ass 两种格式都导出存在
	if ffmpegInfo.SubtitleInfoList == nil || len(ffmpegInfo.SubtitleInfoList) <= 0 || oneSubAndIsError == true {

		if ffmpegInfo.AudioInfoList == nil || len(ffmpegInfo.AudioInfoList) == 0 {
			return errors.New("SubTimelineFixerHelperEx.Process.GetFFMPEGInfo Can`t Find SubTitle And Audio To Export -- " + videoFileFullPath)
		}

		// 如果内置字幕没有，那么就需要尝试获取音频信息
		bok, ffmpegInfo, err = s.ffmpegHelper.GetFFMPEGInfo(videoFileFullPath, ffmpeg_helper.Audio)
		if err != nil {
			return err
		}
		if bok == false {
			return errors.New("SubTimelineFixerHelperEx.Process.GetFFMPEGInfo = false Audio -- " + videoFileFullPath)
		}

		// 使用音频进行时间轴的校正
		if len(ffmpegInfo.AudioInfoList) <= 0 {
			s.log.Warnln("Can`t find audio info, skip time fix --", videoFileFullPath)
			return nil
		}
		bProcess, infoSrc, pipeResultMax, err = s.processByAudio(ffmpegInfo.AudioInfoList[0].FullPath, srcSubFPath)
		if err != nil {
			return err
		}
	} else {
		// 使用内置的字幕进行时间轴的校正，这里需要考虑一个问题，内置的字幕可能是有问题的（先考虑一种，就是字幕的长度不对，是一小段的）
		// 那么就可以比较多个内置字幕的大小选择大的去使用
		// 如果有多个内置的字幕，还是要判断下的，选体积最大的那个吧
		fileSizes := treemap.NewWith(utils.Int64Comparator)
		for index, info := range ffmpegInfo.SubtitleInfoList {
			fi, err := os.Stat(info.FullPath)
			if err != nil {
				fileSizes.Put(0, index)
			} else {
				fileSizes.Put(fi.Size(), index)
			}
		}
		_, index := fileSizes.Max()
		baseSubFPath := ffmpegInfo.SubtitleInfoList[index.(int)].FullPath
		bProcess, infoSrc, pipeResultMax, err = s.processBySub(baseSubFPath, srcSubFPath)
		if err != nil {
			return err
		}
	}

	// 开始调整字幕时间轴
	if bProcess == false || math.Abs(pipeResultMax.GetOffsetTime()) < s.fixerConfig.MinOffset {
		s.log.Infoln("Skip TimeLine Fix -- OffsetTime:", pipeResultMax.GetOffsetTime(), srcSubFPath)
		return nil
	}
	err = s.changeTimeLineAndSave(infoSrc, pipeResultMax, srcSubFPath)
	if err != nil {
		return err
	}
	s.log.Infoln("Fix Offset:", pipeResultMax.GetOffsetTime(), srcSubFPath)
	s.log.Infoln("BackUp Org SubFile:", pipeResultMax.GetOffsetTime(), srcSubFPath+sub_timeline_fixer.BackUpExt)

	return nil
}

func (s SubTimelineFixerHelperEx) processBySub(baseSubFileFPath, srcSubFileFPath string) (bool, *subparser.FileInfo, sub_timeline_fixer.PipeResult, error) {

	bFind, infoBase, err := s.subParserHub.DetermineFileTypeFromFile(baseSubFileFPath)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}
	if bFind == false {
		s.log.Warnln("processBySub.DetermineFileTypeFromFile sub not match --", baseSubFileFPath)
		return false, nil, sub_timeline_fixer.PipeResult{}, nil
	}
	bFind, infoSrc, err := s.subParserHub.DetermineFileTypeFromFile(srcSubFileFPath)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}
	if bFind == false {
		s.log.Warnln("processBySub.DetermineFileTypeFromFile sub not match --", srcSubFileFPath)
		return false, nil, sub_timeline_fixer.PipeResult{}, nil
	}
	// ---------------------------------------------------------------------------------------
	pipeResult, err := s.timelineFixPipeLine.CalcOffsetTime(infoBase, infoSrc, nil, false)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}

	return true, infoSrc, pipeResult, nil
}

func (s SubTimelineFixerHelperEx) processByAudio(baseAudioFileFPath, srcSubFileFPath string) (bool, *subparser.FileInfo, sub_timeline_fixer.PipeResult, error) {

	audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
		FileFullPath: baseAudioFileFPath,
		SampleRate:   16000,
		BitDepth:     16,
	}, true)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}

	bFind, infoSrc, err := s.subParserHub.DetermineFileTypeFromFile(srcSubFileFPath)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}
	if bFind == false {
		s.log.Warnln("processByAudio.DetermineFileTypeFromFile sub not match --", srcSubFileFPath)
		return false, nil, sub_timeline_fixer.PipeResult{}, nil
	}
	// ---------------------------------------------------------------------------------------
	pipeResult, err := s.timelineFixPipeLine.CalcOffsetTime(nil, infoSrc, audioVADInfos, false)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}

	return true, infoSrc, pipeResult, nil
}

func (s SubTimelineFixerHelperEx) changeTimeLineAndSave(infoSrc *subparser.FileInfo, pipeResult sub_timeline_fixer.PipeResult, desSubSaveFPath string) error {
	/*
		修复的字幕先存放到缓存目录，然后需要把原有的字幕进行“备份”，改名，然后再替换过来
	*/
	subFileName := desSubSaveFPath + sub_timeline_fixer.TmpExt
	if my_util.IsFile(subFileName) == true {
		err := os.Remove(subFileName)
		if err != nil {
			return err
		}
	}
	_, err := s.timelineFixPipeLine.FixSubFileTimeline(infoSrc, pipeResult.ScaledFileInfo, pipeResult.GetOffsetTime(), subFileName)
	if err != nil {
		return err
	}

	if my_util.IsFile(desSubSaveFPath+sub_timeline_fixer.BackUpExt) == true {
		err = os.Remove(desSubSaveFPath + sub_timeline_fixer.BackUpExt)
		if err != nil {
			return err
		}
	}

	err = os.Rename(desSubSaveFPath, desSubSaveFPath+sub_timeline_fixer.BackUpExt)
	if err != nil {
		return err
	}

	err = os.Rename(subFileName, desSubSaveFPath)
	if err != nil {
		return err
	}

	return nil
}
