package sub_timeline_fixer

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/ffmpeg_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"github.com/allanpk716/ChineseSubFinder/internal/types/sub_timeline_fiexer"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"os"
)

type SubTimelineFixerHelperEx struct {
	ffmpegHelper       *ffmpeg_helper.FFMPEGHelper
	subParserHub       *sub_parser_hub.SubParserHub
	timelineFixer      *sub_timeline_fixer.SubTimelineFixer
	needDownloadFFMPeg bool
}

func NewSubTimelineFixerHelperEx(fixerConfig sub_timeline_fiexer.SubTimelineFixerConfig) *SubTimelineFixerHelperEx {
	return &SubTimelineFixerHelperEx{
		ffmpegHelper:       ffmpeg_helper.NewFFMPEGHelper(),
		subParserHub:       sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser()),
		timelineFixer:      sub_timeline_fixer.NewSubTimelineFixer(fixerConfig),
		needDownloadFFMPeg: false,
	}
}

// Check 是否安装了 ffmpeg 和 ffprobe
func (s *SubTimelineFixerHelperEx) Check() bool {
	version, err := s.ffmpegHelper.Version()
	if err != nil {
		s.needDownloadFFMPeg = false
		log_helper.GetLogger().Errorln("Need Install ffmpeg and ffprobe !")
		return false
	}
	s.needDownloadFFMPeg = true
	log_helper.GetLogger().Infoln(version)
	return true
}

func (s SubTimelineFixerHelperEx) Process(videoFileFullPath, srcSubFPath string) error {

	if s.needDownloadFFMPeg == false {
		log_helper.GetLogger().Errorln("Need Install ffmpeg and ffprobe, Can't Do TimeLine Fix")
		return nil
	}

	var infoSrc *subparser.FileInfo
	bProcess := false
	offSetTime := 0.0
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
	// 内置的字幕，这里只列举一种格式出来，其实会有一个字幕的 srt 和 ass 两种格式都导出存在
	// len(ffmpegInfo.SubtitleInfoList)
	if len(ffmpegInfo.SubtitleInfoList) <= 0 {
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
			log_helper.GetLogger().Warnln("Can`t find audio info, skip time fix --", videoFileFullPath)
			return nil
		}
		bProcess, infoSrc, offSetTime, err = s.processByAudio(ffmpegInfo.AudioInfoList[0].FullPath, srcSubFPath)
		if err != nil {
			return err
		}
	} else {
		// 使用内置的字幕进行时间轴的校正，这里需要考虑一个问题，内置的字幕可能是有问题的（先考虑一种，就是字幕的长度不对，是一小段的）
		// 那么就可以比较多个内置字幕的大小选择大的去使用
		baseSubFPath := ""
		if len(ffmpegInfo.SubtitleInfoList) > 1 {
			// 如果有多个内置的字幕，还是要判断下的，选体积最大的那个吧
			fileSizes := treemap.NewWith(utils.Int64Comparator)
			for index, info := range ffmpegInfo.ExternalSubInfos {
				fi, err := os.Stat(info.FileFullPath)
				if err != nil {
					fileSizes.Put(0, index)
				} else {
					fileSizes.Put(fi.Size(), index)
				}
			}
			_, index := fileSizes.Max()
			baseSubFPath = ffmpegInfo.ExternalSubInfos[index.(int)].FileFullPath
		} else {
			// 如果只有一个字幕就没必要纠结了，用这个去对比吧
			baseSubFPath = ffmpegInfo.SubtitleInfoList[0].FullPath
		}
		bProcess, infoSrc, offSetTime, err = s.processBySub(baseSubFPath, srcSubFPath)
		if err != nil {
			return err
		}
	}

	// 开始调整字幕时间轴
	if bProcess == false {
		log_helper.GetLogger().Infoln("Skip TimeLine Fix --", srcSubFPath)
		return nil
	}
	err = s.changeTimeLineAndSave(infoSrc, offSetTime, srcSubFPath)
	if err != nil {
		return err
	}

	log_helper.GetLogger().Infoln("Fix Offset:", offSetTime, srcSubFPath)
	log_helper.GetLogger().Infoln("BackUp Org SubFile:", offSetTime, srcSubFPath+BackUpExt)

	return nil
}

func (s SubTimelineFixerHelperEx) processBySub(baseSubFileFPath, srcSubFileFPath string) (bool, *subparser.FileInfo, float64, error) {

	bFind, infoBase, err := s.subParserHub.DetermineFileTypeFromFile(baseSubFileFPath)
	if err != nil {
		return false, nil, 0, err
	}
	if bFind == false {
		log_helper.GetLogger().Warnln("processBySub.DetermineFileTypeFromFile sub not match --", baseSubFileFPath)
		return false, nil, 0, nil
	}
	bFind, infoSrc, err := s.subParserHub.DetermineFileTypeFromFile(srcSubFileFPath)
	if err != nil {
		return false, nil, 0, err
	}
	if bFind == false {
		log_helper.GetLogger().Warnln("processBySub.DetermineFileTypeFromFile sub not match --", srcSubFileFPath)
		return false, nil, 0, nil
	}
	// ---------------------------------------------------------------------------------------
	baseUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoBase, s.timelineFixer.FixerConfig.V2_FrontAndEndPerBase)
	if err != nil {
		return false, nil, 0, err
	}
	srcUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoSrc, s.timelineFixer.FixerConfig.V2_FrontAndEndPerSrc)
	if err != nil {
		return false, nil, 0, err
	}
	// ---------------------------------------------------------------------------------------
	bok, offsetTime, sd, err := s.timelineFixer.GetOffsetTimeV2(baseUnitNew, srcUnitNew, nil)
	if err != nil {
		return false, nil, 0, err
	}
	if bok == false {
		log_helper.GetLogger().Warnln("processSub.GetOffsetTimeV2 return false -- " + baseSubFileFPath)
		return false, nil, 0, nil
	}

	if s.jugOffsetAndSD("processBySub", offsetTime, sd) == false {
		return false, nil, 0, nil
	}

	return true, infoSrc, offsetTime, nil
}

func (s SubTimelineFixerHelperEx) processByAudio(baseAudioFileFPath, srcSubFileFPath string) (bool, *subparser.FileInfo, float64, error) {

	audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
		FileFullPath: baseAudioFileFPath,
		SampleRate:   16000,
		BitDepth:     16,
	}, true)
	if err != nil {
		return false, nil, 0, err
	}

	bFind, infoSrc, err := s.subParserHub.DetermineFileTypeFromFile(srcSubFileFPath)
	if err != nil {
		return false, nil, 0, err
	}
	if bFind == false {
		log_helper.GetLogger().Warnln("processByAudio.DetermineFileTypeFromFile sub not match --", srcSubFileFPath)
		return false, nil, 0, nil
	}
	// ---------------------------------------------------------------------------------------
	srcUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoSrc, s.timelineFixer.FixerConfig.V2_FrontAndEndPerSrc)
	if err != nil {
		return false, nil, 0, err
	}
	// ---------------------------------------------------------------------------------------
	bok, offsetTime, sd, err := s.timelineFixer.GetOffsetTimeV2(nil, srcUnitNew, audioVADInfos)
	if err != nil {
		return false, nil, 0, err
	}
	if bok == false {
		log_helper.GetLogger().Warnln("processSub.GetOffsetTimeV2 return false -- " + baseAudioFileFPath)
		return false, nil, 0, nil
	}

	if s.jugOffsetAndSD("processByAudio", offsetTime, sd) == false {
		return false, nil, 0, nil
	}

	return true, infoSrc, offsetTime, nil
}

func (s SubTimelineFixerHelperEx) jugOffsetAndSD(processName string, offsetTime, sd float64) bool {
	// SD 要达标
	if sd > s.timelineFixer.FixerConfig.V2_MaxStartTimeDiffSD {
		log_helper.GetLogger().Infoln(fmt.Sprintf("skip, %s sd: %v > %v", processName, sd, s.timelineFixer.FixerConfig.V2_MaxStartTimeDiffSD))
		return false
	}
	// 时间偏移的最小值才修正
	if offsetTime < s.timelineFixer.FixerConfig.V2_MinOffset && offsetTime > -s.timelineFixer.FixerConfig.V2_MinOffset {
		log_helper.GetLogger().Infoln(fmt.Sprintf("Skip, %s offset: %v > -%v && %v < %v",
			processName,
			offsetTime, s.timelineFixer.FixerConfig.V2_MinOffset,
			offsetTime, s.timelineFixer.FixerConfig.V2_MinOffset))
		return false
	}
	// sub_timeline_fixer.calcMeanAndSD 输出的可能的极小值
	if offsetTime <= sub_timeline_fixer.MinValue+0.1 || offsetTime >= sub_timeline_fixer.MinValue-0.1 {
		return false
	}

	return true
}

func (s SubTimelineFixerHelperEx) changeTimeLineAndSave(infoSrc *subparser.FileInfo, offsetTime float64, desSubSaveFPath string) error {
	/*
		修复的字幕先存放到缓存目录，然后需要把原有的字幕进行“备份”，改名，然后再替换过来
	*/
	subFileName := desSubSaveFPath + TmpExt
	if my_util.IsFile(subFileName) == true {
		err := os.Remove(subFileName)
		if err != nil {
			return err
		}
	}
	_, err := s.timelineFixer.FixSubTimeline(infoSrc, offsetTime, subFileName)
	if err != nil {
		return err
	}

	if my_util.IsFile(desSubSaveFPath+BackUpExt) == true {
		err = os.Remove(desSubSaveFPath + BackUpExt)
		if err != nil {
			return err
		}
	}

	err = os.Rename(desSubSaveFPath, desSubSaveFPath+BackUpExt)
	if err != nil {
		return err
	}

	err = os.Rename(subFileName, desSubSaveFPath)
	if err != nil {
		return err
	}

	return nil
}

const TmpExt = ".csf-tmp"
const BackUpExt = ".csf-bk"
