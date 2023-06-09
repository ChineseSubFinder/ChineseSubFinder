package sub_timeline_fixer

import (
	"errors"
	"math"
	"os"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ffmpeg_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_timeline_fixer"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/vad"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/sirupsen/logrus"
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

func (s *SubTimelineFixerHelperEx) Process(videoFileFullPath, srcSubFPath string) error {

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
	bok, ffmpegInfo, err = s.ffmpegHelper.ExportFFMPEGInfo(videoFileFullPath, ffmpeg_helper.Subtitle)
	if err != nil {
		return err
	}
	if bok == false {
		return errors.New("SubTimelineFixerHelperEx.Process.ExportFFMPEGInfo = false Subtitle -- " + videoFileFullPath)
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
			return errors.New("SubTimelineFixerHelperEx.Process.ExportFFMPEGInfo Can`t Find SubTitle And Audio To Export -- " + videoFileFullPath)
		}

		// 如果内置字幕没有，那么就需要尝试获取音频信息
		bok, ffmpegInfo, err = s.ffmpegHelper.ExportFFMPEGInfo(videoFileFullPath, ffmpeg_helper.Audio)
		if err != nil {
			return err
		}
		if bok == false {
			return errors.New("SubTimelineFixerHelperEx.Process.ExportFFMPEGInfo = false Audio -- " + videoFileFullPath)
		}

		// 使用音频进行时间轴的校正
		if len(ffmpegInfo.AudioInfoList) <= 0 {
			s.log.Warnln("Can`t find audio info, skip time fix --", videoFileFullPath)
			return nil
		}
		bProcess, infoSrc, pipeResultMax, err = s.ProcessByAudioFile(ffmpegInfo.AudioInfoList[0].FullPath, srcSubFPath)
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
		bProcess, infoSrc, pipeResultMax, err = s.ProcessBySubFile(baseSubFPath, srcSubFPath)
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
	s.log.Infoln("TimeLine Fix -- Score:", pipeResultMax.Score, srcSubFPath)
	s.log.Infoln("Fix Offset:", pipeResultMax.GetOffsetTime(), srcSubFPath)
	s.log.Infoln("BackUp Org SubFile:", pipeResultMax.GetOffsetTime(), srcSubFPath+sub_timeline_fixer.BackUpExt)

	return nil
}

func (s *SubTimelineFixerHelperEx) ProcessBySubFileInfo(infoBase *subparser.FileInfo, infoSrc *subparser.FileInfo) (bool, *subparser.FileInfo, sub_timeline_fixer.PipeResult, error) {

	// ---------------------------------------------------------------------------------------
	pipeResult, err := s.timelineFixPipeLine.CalcOffsetTime(infoBase, infoSrc, nil, false)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}

	return true, infoSrc, pipeResult, nil
}

func (s *SubTimelineFixerHelperEx) ProcessBySubFile(baseSubFileFPath, srcSubFileFPath string) (bool, *subparser.FileInfo, sub_timeline_fixer.PipeResult, error) {

	bFind, infoBase, err := s.subParserHub.DetermineFileTypeFromFile(baseSubFileFPath)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}
	if bFind == false {
		s.log.Warnln("ProcessBySubFile.DetermineFileTypeFromFile sub not match --", baseSubFileFPath)
		return false, nil, sub_timeline_fixer.PipeResult{}, nil
	}

	bFind, infoSrc, err := s.subParserHub.DetermineFileTypeFromFile(srcSubFileFPath)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}
	if bFind == false {
		s.log.Warnln("ProcessBySubFile.DetermineFileTypeFromFile sub not match --", srcSubFileFPath)
		return false, nil, sub_timeline_fixer.PipeResult{}, nil
	}

	return s.ProcessBySubFileInfo(infoBase, infoSrc)
}

func (s *SubTimelineFixerHelperEx) ProcessByAudioVAD(audioVADInfos []vad.VADInfo, infoSrc *subparser.FileInfo) (bool, *subparser.FileInfo, sub_timeline_fixer.PipeResult, error) {

	// ---------------------------------------------------------------------------------------
	pipeResult, err := s.timelineFixPipeLine.CalcOffsetTime(nil, infoSrc, audioVADInfos, false)
	if err != nil {
		return false, nil, sub_timeline_fixer.PipeResult{}, err
	}

	return true, infoSrc, pipeResult, nil
}

func (s *SubTimelineFixerHelperEx) ProcessByAudioFile(baseAudioFileFPath, srcSubFileFPath string) (bool, *subparser.FileInfo, sub_timeline_fixer.PipeResult, error) {

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
		s.log.Warnln("ProcessByAudioFile.DetermineFileTypeFromFile sub not match --", srcSubFileFPath)
		return false, nil, sub_timeline_fixer.PipeResult{}, nil
	}

	return s.ProcessByAudioVAD(audioVADInfos, infoSrc)
}

func (s *SubTimelineFixerHelperEx) IsVideoCanExportSubtitleAndAudio(videoFileFullPath string) (bool, *ffmpeg_helper.FFMPEGInfo, []vad.VADInfo, *subparser.FileInfo, error) {

	// 先尝试获取内置字幕的信息
	bok, ffmpegInfo, err := s.ffmpegHelper.ExportFFMPEGInfo(videoFileFullPath, ffmpeg_helper.SubtitleAndAudio)
	if err != nil {
		return false, nil, nil, nil, err
	}
	if bok == false {
		return false, nil, nil, nil, nil
	}
	// ---------------------------------------------------------------------------------------
	// 音频
	if len(ffmpegInfo.AudioInfoList) <= 0 {
		return false, nil, nil, nil, nil
	}
	audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
		FileFullPath: ffmpegInfo.AudioInfoList[0].FullPath,
		SampleRate:   16000,
		BitDepth:     16,
	}, true)
	if err != nil {
		return false, nil, nil, nil, err
	}
	// ---------------------------------------------------------------------------------------
	// 字幕
	if len(ffmpegInfo.SubtitleInfoList) <= 0 {
		return false, nil, nil, nil, nil
	}
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
	bFind, infoBase, err := s.subParserHub.DetermineFileTypeFromFile(baseSubFPath)
	if err != nil {
		return false, nil, nil, nil, err
	}
	if bFind == false {
		return false, nil, nil, nil, nil
	}
	// ---------------------------------------------------------------------------------------

	return true, ffmpegInfo, audioVADInfos, infoBase, nil
}

func (s *SubTimelineFixerHelperEx) IsMatchBySubFile(ffmpegInfo *ffmpeg_helper.FFMPEGInfo, audioVADInfos []vad.VADInfo, infoBase *subparser.FileInfo, srcSubFileFPath string, config CompareConfig) (bool, *MatchResult, error) {

	bFind, srcBase, err := s.subParserHub.DetermineFileTypeFromFile(srcSubFileFPath)
	if err != nil {
		return false, nil, err
	}
	if bFind == false {
		return false, nil, nil
	}
	// ---------------------------------------------------------------------------------------
	// 音频
	s.log.Infoln("IsMatchBySubFile:", srcSubFileFPath)
	bProcess, _, pipeResultMaxAudio, err := s.ProcessByAudioVAD(audioVADInfos, srcBase)
	if err != nil {
		return false, nil, err
	}
	if bProcess == false {
		return false, nil, nil
	}
	// ---------------------------------------------------------------------------------------
	// 字幕
	bProcess, _, pipeResultMaxSub, err := s.ProcessBySubFileInfo(infoBase, srcBase)
	if err != nil {
		return false, nil, err
	}
	if bProcess == false {
		return false, nil, nil
	}

	targetSubEndTime := pkg.Time2SecondNumber(srcBase.GetEndTime())

	matchResult := &MatchResult{
		VideoDuration:          ffmpegInfo.Duration,
		TargetSubEndTime:       targetSubEndTime,
		AudioCompareScore:      pipeResultMaxAudio.Score,
		AudioCompareOffsetTime: pipeResultMaxAudio.GetOffsetTime(),
		SubCompareScore:        pipeResultMaxSub.Score,
		SubCompareOffsetTime:   pipeResultMaxSub.GetOffsetTime(),
	}
	// ---------------------------------------------------------------------------------------
	// 分数需要大于某个值
	if pipeResultMaxAudio.Score < config.MinScore || pipeResultMaxSub.Score < config.MinScore {
		return false, matchResult, nil
	}
	// 两种方式获取到的时间轴的偏移量，差值需要在一定范围内
	if math.Abs(pipeResultMaxAudio.GetOffsetTime()-pipeResultMaxSub.GetOffsetTime()) > config.OffsetRange {
		return false, matchResult, nil
	}
	// ---------------------------------------------------------------------------------------
	// 待判断的字幕的时间长度要小于等于视频的总长度
	if targetSubEndTime > ffmpegInfo.Duration {
		return false, matchResult, nil
	}
	// ---------------------------------------------------------------------------------------
	// 两个对比字幕的对白数量不能超过 10%
	minRage := float64(len(infoBase.Dialogues)) * config.DialoguesDifferencePercentage
	if math.Abs(float64(len(srcBase.Dialogues)-len(infoBase.Dialogues))) > minRage {
		return false, matchResult, nil
	}
	return true, matchResult, nil
}

func (s *SubTimelineFixerHelperEx) changeTimeLineAndSave(infoSrc *subparser.FileInfo, pipeResult sub_timeline_fixer.PipeResult, desSubSaveFPath string) error {
	/*
		修复的字幕先存放到缓存目录，然后需要把原有的字幕进行“备份”，改名，然后再替换过来
	*/
	subFileName := desSubSaveFPath + sub_timeline_fixer.TmpExt
	if pkg.IsFile(subFileName) == true {
		err := os.Remove(subFileName)
		if err != nil {
			return err
		}
	}
	_, err := s.timelineFixPipeLine.FixSubFileTimeline(infoSrc, pipeResult.ScaledFileInfo, pipeResult.GetOffsetTime(), subFileName)
	if err != nil {
		return err
	}

	if pkg.IsFile(desSubSaveFPath+sub_timeline_fixer.BackUpExt) == true {
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

type CompareConfig struct {
	MinScore                      float64 // 最低的分数
	OffsetRange                   float64 // 偏移量的范围
	DialoguesDifferencePercentage float64 // 两个字幕的对白字幕差异百分比
}

type MatchResult struct {
	VideoDuration          float64 // 视频的时长
	TargetSubEndTime       float64 // 目标字幕的结束时间
	AudioCompareScore      float64 // 音频的对比分数
	AudioCompareOffsetTime float64 // 音频的对比偏移量
	SubCompareScore        float64 // 字幕的对比分数
	SubCompareOffsetTime   float64 // 字幕的对比偏移量

}
