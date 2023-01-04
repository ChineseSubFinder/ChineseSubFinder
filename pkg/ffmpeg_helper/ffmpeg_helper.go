package ffmpeg_helper

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"strings"
)

type FFMPEGHelper struct {
	log          *logrus.Logger
	SubParserHub *sub_parser_hub.SubParserHub // 字幕内容的解析器
}

func NewFFMPEGHelper(log *logrus.Logger) *FFMPEGHelper {
	return &FFMPEGHelper{
		log:          log,
		SubParserHub: sub_parser_hub.NewSubParserHub(log, ass.NewParser(log), srt.NewParser(log)),
	}
}

// Version 获取版本信息，如果不存在 FFMPEG 和 ffprobe 则报错
func (f *FFMPEGHelper) Version() (string, error) {

	outMsg0, err := f.getVersion("ffmpeg")
	if err != nil {
		return "", err
	}
	outMsg1, err := f.getVersion("ffprobe")
	if err != nil {
		return "", err
	}

	return outMsg0 + "\r\n" + outMsg1, nil
}

// ExportFFMPEGInfo 获取 视频的 FFMPEG 信息，包含音频和字幕
// 优先会导出 中、英、日、韩 类型的，字幕如果没有语言类型，则也导出，然后需要额外的字幕语言的判断去辅助标记（读取文件内容）
// 音频只会导出一个，优先导出 中、英、日、韩 类型的
func (f *FFMPEGHelper) ExportFFMPEGInfo(videoFileFullPath string, exportType ExportType) (bool, *FFMPEGInfo, error) {

	const args = "-v error -show_format -show_streams -print_format json"
	cmdArgs := strings.Fields(args)
	cmdArgs = append(cmdArgs, videoFileFullPath)
	cmd := exec.Command("ffprobe", cmdArgs...)
	buf := bytes.NewBufferString("")
	//指定输出位置
	cmd.Stderr = buf
	cmd.Stdout = buf
	err := cmd.Start()
	if err != nil {
		return false, nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return false, nil, err
	}
	// 解析得到的字符串反馈
	bok, ffMPEGInfo, _ := f.parseJsonString2GetFFProbeInfo(videoFileFullPath, buf.String())
	if bok == false {
		return false, nil, nil
	}
	nowCacheFolderPath, err := ffMPEGInfo.GetCacheFolderFPath()
	if err != nil {
		return false, nil, err
	}
	// 在函数调用完毕后，判断是否需要清理
	defer func() {
		if bok == false && ffMPEGInfo != nil {
			err := os.RemoveAll(nowCacheFolderPath)
			if err != nil {
				f.log.Errorln("ExportFFMPEGInfo - RemoveAll", err.Error())
				return
			}
		}
	}()

	// 查找当前这个视频外置字幕列表
	err = ffMPEGInfo.GetExternalSubInfos(f.SubParserHub)
	if err != nil {
		return false, nil, err
	}

	ffMPEGInfo.Duration = f.GetVideoDuration(videoFileFullPath)

	// 判断这个视频是否已经导出过内置的字幕和音频文件了
	if ffMPEGInfo.IsExported(exportType) == false {
		// 说明缓存不存在，需要导出，这里需要注意，如果导出失败了，这个文件夹要清理掉
		if pkg.IsDir(nowCacheFolderPath) == true {
			// 如果存在则，先清空一个这个文件夹
			err = pkg.ClearFolder(nowCacheFolderPath)
			if err != nil {
				bok = false
				return bok, nil, err
			}
		}
		// 开始导出
		// 构建导出的命令参数
		exportAudioArgs, exportSubArgs := f.getAudioAndSubExportArgs(videoFileFullPath, ffMPEGInfo)

		// 上面导出的信息，可能是 nil 参数，那么就直接把导出的 List 信息给置为 nil，让后续有依据可以跳出，不继续执行
		if exportType == Subtitle {
			if exportSubArgs == nil {
				ffMPEGInfo.SubtitleInfoList = nil
				return true, ffMPEGInfo, nil
			}
		} else if exportType == Audio {
			if exportAudioArgs == nil {
				ffMPEGInfo.AudioInfoList = nil
				return true, ffMPEGInfo, nil
			}
		} else if exportType == SubtitleAndAudio {
			if exportAudioArgs == nil || exportSubArgs == nil {
				if exportAudioArgs == nil {
					ffMPEGInfo.AudioInfoList = nil
				}
				if exportSubArgs == nil {
					ffMPEGInfo.SubtitleInfoList = nil
				}
				return true, ffMPEGInfo, nil
			}
		} else {
			f.log.Errorln("ExportFFMPEGInfo.getAudioAndSubExportArgs Not Support ExportType")
			return false, nil, nil
		}
		// 上面的操作为了就是确保后续的导出不会出问题
		// 执行导出，音频和内置的字幕
		execErrorString, err := f.exportAudioAndSubtitles(exportAudioArgs, exportSubArgs, exportType)
		if err != nil {
			f.log.Errorln("exportAudioAndSubtitles", execErrorString)
			bok = false
			return bok, nil, err
		}
		// 导出后，需要把现在导出的文件的路径给复制给 ffMPEGInfo 中
		// 音频是否导出了
		ffMPEGInfo.isAudioExported(nowCacheFolderPath)
		// 字幕都要导出了
		ffMPEGInfo.isSubExported(nowCacheFolderPath)
		// 创建 exportedMakeFileName 这个文件
		// 成功，那么就需要生成这个 exportedMakeFileName 文件
		err = ffMPEGInfo.CreateExportedMask()
		if err != nil {
			return false, nil, err
		}
	}

	return bok, ffMPEGInfo, nil
}

// ExportAudioDurationInfo 获取音频的长度信息
func (f *FFMPEGHelper) ExportAudioDurationInfo(audioFileFullPath string) (bool, float64, error) {

	const args = "-v error -show_format -show_streams -print_format json -f s16le -ac 1 -ar 16000"
	cmdArgs := strings.Fields(args)
	cmdArgs = append(cmdArgs, audioFileFullPath)
	cmd := exec.Command("ffprobe", cmdArgs...)
	buf := bytes.NewBufferString("")
	//指定输出位置
	cmd.Stderr = buf
	cmd.Stdout = buf
	err := cmd.Start()
	if err != nil {
		return false, 0, err
	}
	err = cmd.Wait()
	if err != nil {
		return false, 0, err
	}

	bok, duration := f.parseJsonString2GetAudioInfo(buf.String())
	if bok == false {
		return false, 0, errors.New("ffprobe get " + audioFileFullPath + " duration error")
	}

	return true, duration, nil
}

// ExportAudioAndSubArgsByTimeRange 根据输入的时间轴导出音频分段信息 "0:1:27" "28.2"
func (f *FFMPEGHelper) ExportAudioAndSubArgsByTimeRange(audioFullPath, subFullPath string, startTimeString, timeLength string) (string, string, string, error) {

	outStartTimeString := strings.ReplaceAll(startTimeString, ":", "-")
	outStartTimeString = strings.ReplaceAll(outStartTimeString, ".", "#")

	outTimeLength := strings.ReplaceAll(timeLength, ".", "#")

	frontName := strings.ReplaceAll(filepath.Base(audioFullPath), filepath.Ext(audioFullPath), "")

	outAudioName := frontName + "_" + outStartTimeString + "_" + outTimeLength + filepath.Ext(audioFullPath)
	outSubName := frontName + "_" + outStartTimeString + "_" + outTimeLength + common.SubExtSRT

	var outAudioFullPath = filepath.Join(filepath.Dir(audioFullPath), outAudioName)
	var outSubFullPath = filepath.Join(filepath.Dir(audioFullPath), outSubName)

	// 导出音频
	if pkg.IsFile(outAudioFullPath) == true {
		err := os.Remove(outAudioFullPath)
		if err != nil {
			return "", "", "", err
		}
	}
	args := f.getAudioExportArgsByTimeRange(audioFullPath, startTimeString, timeLength, outAudioFullPath)
	execFFMPEG, err := f.execFFMPEG(args)
	if err != nil {
		return "", "", execFFMPEG, err
	}
	// 导出字幕
	if pkg.IsFile(outSubFullPath) == true {
		err := os.Remove(outSubFullPath)
		if err != nil {
			return "", "", "", err
		}
	}
	args = f.getSubExportArgsByTimeRange(subFullPath, startTimeString, timeLength, outSubFullPath)
	execFFMPEG, err = f.execFFMPEG(args)
	if err != nil {
		return "", "", execFFMPEG, err
	}

	return outAudioFullPath, outSubFullPath, "", nil
}

// ExportSubArgsByTimeRange 根据输入的时间轴导出字幕分段信息 "0:1:27" "28.2"
func (f *FFMPEGHelper) ExportSubArgsByTimeRange(subFullPath, outName string, startTimeString, timeLength string) (string, string, error) {

	outStartTimeString := strings.ReplaceAll(startTimeString, ":", "-")
	outStartTimeString = strings.ReplaceAll(outStartTimeString, ".", "#")

	outTimeLength := strings.ReplaceAll(timeLength, ".", "#")

	frontName := strings.ReplaceAll(filepath.Base(subFullPath), filepath.Ext(subFullPath), "")

	outSubName := frontName + "_" + outStartTimeString + "_" + outTimeLength + "_" + outName + common.SubExtSRT

	var outSubFullPath = filepath.Join(filepath.Dir(subFullPath), outSubName)

	// 导出字幕
	if pkg.IsFile(outSubFullPath) == true {
		err := os.Remove(outSubFullPath)
		if err != nil {
			return "", "", err
		}
	}
	args := f.getSubExportArgsByTimeRange(subFullPath, startTimeString, timeLength, outSubFullPath)
	execFFMPEG, err := f.execFFMPEG(args)
	if err != nil {
		return "", execFFMPEG, err
	}

	return outSubFullPath, "", nil
}

// ExportVideoHLSAndSubByTimeRange 导出指定的时间轴的视频HLS和字幕，然后从 outDirPath 中获取 outputlist.m3u8 和字幕的文件
func (f *FFMPEGHelper) ExportVideoHLSAndSubByTimeRange(videoFullPath string, subFullPaths []string, startTimeString, timeLength, segmentTime, outDirPath string) (string, []string, error) {

	// 导出视频
	if pkg.IsFile(videoFullPath) == false {

		bok, _, steamDirPath := decode.IsFakeBDMVWorked(videoFullPath)
		if bok == true {
			// 需要从 steamDirPath 搜索最大的一个文件出来
			videoFullPath = pkg.GetMaxSizeFile(steamDirPath)
		} else {
			return "", nil, errors.New("video file not found")
		}
	}

	for _, subFullPath := range subFullPaths {
		if pkg.IsFile(subFullPath) == false {
			return "", nil, errors.New("sub file not exist:" + subFullPath)
		}
	}

	fileName := filepath.Base(videoFullPath)
	frontName := strings.ReplaceAll(fileName, filepath.Ext(fileName), "")

	outDirSubPath := filepath.Join(outDirPath, frontName, startTimeString+"-"+timeLength)
	if pkg.IsDir(outDirSubPath) == true {
		err := os.RemoveAll(outDirSubPath)
		if err != nil {
			return "", nil, err
		}
	}
	err := os.MkdirAll(outDirSubPath, os.ModePerm)
	if err != nil {
		return "", nil, err
	}

	// 先剪切
	//videoExt := filepath.Ext(fileName)
	//cutOffVideoFPath := filepath.Join(outDirPath, frontName+"_cut"+videoExt)
	//args := f.getVideoExportArgsByTimeRange(videoFullPath, startTimeString, timeLength, cutOffVideoFPath)
	//execFFMPEG, err := f.execFFMPEG(args)
	//if err != nil {
	//	return "", nil, errors.New(execFFMPEG + err.Error())
	//}
	//// 转换 HLS
	//args = f.getVideo2HLSArgs(cutOffVideoFPath, segmentTime, outDirPath)
	//execFFMPEG, err = f.execFFMPEG(args)
	//if err != nil {
	//	return errors.New(execFFMPEG + err.Error())
	//}

	// 直接导出
	args := f.getVideoHLSExportArgsByTimeRange(videoFullPath, startTimeString, timeLength, segmentTime, outDirSubPath)
	execFFMPEG, err := f.execFFMPEG(args)
	if err != nil {
		return "", nil, errors.New(execFFMPEG + err.Error())
	}

	// 导出字幕
	outSubFPaths := make([]string, 0)
	for i, subFullPath := range subFullPaths {

		tmpSubFPath := subFullPath
		nowSubExt := filepath.Ext(tmpSubFPath)

		if strings.ToLower(nowSubExt) != common.SubExtSRT {
			// 这里需要优先判断字幕是否是 SRT，如果是 ASS 的，那么需要转换一次才行
			middleSubFPath := filepath.Join(outDirSubPath, fmt.Sprintf(frontName+"_middle_%d"+common.SubExtSRT, i))
			args = f.getSubASS2SRTArgs(tmpSubFPath, middleSubFPath)
			execFFMPEG, err = f.execFFMPEG(args)
			if err != nil {
				return "", nil, errors.New(execFFMPEG + err.Error())
			}
			tmpSubFPath = middleSubFPath
		}
		outSubFileFPath := filepath.Join(outDirSubPath, fmt.Sprintf(frontName+"_%d"+common.SubExtSRT, i))
		args = f.getSubExportArgsByTimeRange(tmpSubFPath, startTimeString, timeLength, outSubFileFPath)
		execFFMPEG, err = f.execFFMPEG(args)
		if err != nil {
			return "", nil, errors.New(execFFMPEG + err.Error())
		}
		// 字幕的相对位置
		subRelPath, err := filepath.Rel(outDirPath, outSubFileFPath)
		if err != nil {
			return "", nil, err
		}
		outSubFPaths = append(outSubFPaths, subRelPath)
	}

	// outputlist.m3u8 的相对位置
	outputListRelPath, err := filepath.Rel(outDirPath, filepath.Join(outDirSubPath, "outputlist.m3u8"))
	if err != nil {
		return "", nil, err
	}

	return outputListRelPath, outSubFPaths, nil
}

// parseJsonString2GetFFProbeInfo 使用 ffprobe 获取视频的 stream 信息，从中解析出字幕和音频的索引
func (f *FFMPEGHelper) parseJsonString2GetFFProbeInfo(videoFileFullPath, inputFFProbeString string) (bool, *FFMPEGInfo, *FFMPEGInfo) {

	streamsValue := gjson.Get(inputFFProbeString, "streams.#")
	if streamsValue.Exists() == false {
		return false, nil, nil
	}

	ffmpegInfoFlitter := NewFFMPEGInfo(f.log, videoFileFullPath)
	ffmpegInfoFull := NewFFMPEGInfo(f.log, videoFileFullPath)

	// 进行字幕和音频的缓存，优先当然是导出 中、英、日、韩 相关的字幕和音频
	// 但是如果都没得这些的时候，那么也需要导出至少一个字幕或者音频，用于字幕的校正
	cacheAudios := make([]AudioInfo, 0)
	cacheSubtitleInfos := make([]SubtitleInfo, 0)

	for i := 0; i < int(streamsValue.Num); i++ {

		oneIndex := gjson.Get(inputFFProbeString, fmt.Sprintf("streams.%d.index", i))
		oneCodecName := gjson.Get(inputFFProbeString, fmt.Sprintf("streams.%d.codec_name", i))
		oneCodecType := gjson.Get(inputFFProbeString, fmt.Sprintf("streams.%d.codec_type", i))
		oneTimeBase := gjson.Get(inputFFProbeString, fmt.Sprintf("streams.%d.time_base", i))
		oneStartTime := gjson.Get(inputFFProbeString, fmt.Sprintf("streams.%d.start_time", i))

		oneLanguage := gjson.Get(inputFFProbeString, fmt.Sprintf("streams.%d.tags.language", i))
		// 任意一个字段不存在则跳过
		if oneIndex.Exists() == false {
			continue
		}
		if oneCodecName.Exists() == false {
			continue
		}
		if oneCodecType.Exists() == false {
			continue
		}
		if oneTimeBase.Exists() == false {
			continue
		}
		if oneStartTime.Exists() == false {
			continue
		}
		// 这里需要区分是字幕还是音频
		if oneCodecType.String() == codecTypeSub {
			// 字幕
			// 只解析 subrip 类型的，不支持 hdmv_pgs_subtitle 的字幕导出
			if f.isSupportSubCodecName(oneCodecName.String()) == false {
				continue
			}
			// 这里非必须解析到 language 字段，把所有的都导出来，然后通过额外字幕语言判断即可
			oneDurationTS := gjson.Get(inputFFProbeString, fmt.Sprintf("streams.%d.duration_ts", i))
			oneDuration := gjson.Get(inputFFProbeString, fmt.Sprintf("streams.%d.duration", i))
			// 必须存在的
			if oneDurationTS.Exists() == false {
				continue
			}
			if oneDuration.Exists() == false {
				continue
			}
			// 非必须存在的
			nowLanguageString := ""
			if oneLanguage.Exists() == true {
				nowLanguageString = oneLanguage.String()
				// 只导出 中、英、日、韩
				if language.IsSupportISOString(nowLanguageString) == false {

					subInfo := NewSubtitleInfo(int(oneIndex.Num), oneCodecName.String(), oneCodecType.String(),
						oneTimeBase.String(), oneStartTime.String(),
						int(oneDurationTS.Num), oneDuration.String(), nowLanguageString)
					// 不符合的也存在下来，万一，符合要求的一个都没得的时候，就需要从里面挑几个出来了
					cacheSubtitleInfos = append(cacheSubtitleInfos, *subInfo)
					continue
				}
			}

			subInfo := NewSubtitleInfo(int(oneIndex.Num), oneCodecName.String(), oneCodecType.String(),
				oneTimeBase.String(), oneStartTime.String(),
				int(oneDurationTS.Num), oneDuration.String(), nowLanguageString)

			ffmpegInfoFlitter.SubtitleInfoList = append(ffmpegInfoFlitter.SubtitleInfoList, *subInfo)

		} else if oneCodecType.String() == codecTypeAudio {
			// 音频
			// 这里必要要能够解析到 language 字段
			if oneLanguage.Exists() == false {
				// 不符合的也存在下来，万一，符合要求的一个都没得的时候，就需要从里面挑几个出来了
				audioInfo := NewAudioInfo(int(oneIndex.Num), oneCodecName.String(), oneCodecType.String(),
					oneTimeBase.String(), oneStartTime.String(), oneLanguage.String())

				cacheAudios = append(cacheAudios, *audioInfo)
				continue
			}
			// 只导出 中、英、日、韩
			if language.IsSupportISOString(oneLanguage.String()) == false {
				// 不符合的也存在下来，万一，符合要求的一个都没得的时候，就需要从里面挑几个出来了
				audioInfo := NewAudioInfo(int(oneIndex.Num), oneCodecName.String(), oneCodecType.String(),
					oneTimeBase.String(), oneStartTime.String(), oneLanguage.String())

				cacheAudios = append(cacheAudios, *audioInfo)
				continue
			}
			audioInfo := NewAudioInfo(int(oneIndex.Num), oneCodecName.String(), oneCodecType.String(),
				oneTimeBase.String(), oneStartTime.String(), oneLanguage.String())

			ffmpegInfoFlitter.AudioInfoList = append(ffmpegInfoFlitter.AudioInfoList, *audioInfo)

		} else {
			continue
		}
	}
	// 把过滤的和缓存的都拼接到一起
	for _, audioInfo := range ffmpegInfoFlitter.AudioInfoList {
		ffmpegInfoFull.AudioInfoList = append(ffmpegInfoFull.AudioInfoList, audioInfo)
	}
	for _, audioInfo := range cacheAudios {
		ffmpegInfoFull.AudioInfoList = append(ffmpegInfoFull.AudioInfoList, audioInfo)
	}
	for _, subInfo := range ffmpegInfoFlitter.SubtitleInfoList {
		ffmpegInfoFull.SubtitleInfoList = append(ffmpegInfoFull.SubtitleInfoList, subInfo)
	}
	for _, subInfo := range cacheSubtitleInfos {
		ffmpegInfoFull.SubtitleInfoList = append(ffmpegInfoFull.SubtitleInfoList, subInfo)
	}

	// 如何没有找到合适的字幕，那么就要把缓存的字幕选一个填充进去
	if len(ffmpegInfoFlitter.SubtitleInfoList) == 0 {
		if len(cacheSubtitleInfos) != 0 {
			ffmpegInfoFlitter.SubtitleInfoList = append(ffmpegInfoFlitter.SubtitleInfoList, cacheSubtitleInfos[0])
		}
	}
	// 如何没有找到合适的音频，那么就要把缓存的音频选一个填充进去
	if len(ffmpegInfoFlitter.AudioInfoList) == 0 {
		if len(cacheAudios) != 0 {
			ffmpegInfoFlitter.AudioInfoList = append(ffmpegInfoFlitter.AudioInfoList, cacheAudios[0])
		}
	} else {
		// 音频只需要导出一个就行了，取第一个
		newAudioList := make([]AudioInfo, 0)
		newAudioList = append(newAudioList, ffmpegInfoFlitter.AudioInfoList[0])
		ffmpegInfoFlitter.AudioInfoList = newAudioList
	}

	return true, ffmpegInfoFlitter, ffmpegInfoFull
}

// parseJsonString2GetAudioInfo 获取 pcm 音频的长度
func (f *FFMPEGHelper) parseJsonString2GetAudioInfo(inputFFProbeString string) (bool, float64) {

	durationValue := gjson.Get(inputFFProbeString, "format.duration")
	if durationValue.Exists() == false {
		return false, 0
	}
	return true, durationValue.Float()
}

// exportAudioAndSubtitles 导出音频和字幕文件
func (f *FFMPEGHelper) exportAudioAndSubtitles(audioArgs, subArgs []string, exportType ExportType) (string, error) {

	// 输入的两个数组，有可能是 nil

	// 这里导出依赖的是 ffmpeg 这个程序，需要的是构建导出的语句
	if exportType == SubtitleAndAudio {
		execErrorString, err := f.execFFMPEG(audioArgs)
		if err != nil {
			return execErrorString, err
		}
		execErrorString, err = f.execFFMPEG(subArgs)
		if err != nil {
			return execErrorString, err
		}
	} else if exportType == Audio {
		execErrorString, err := f.execFFMPEG(audioArgs)
		if err != nil {
			return execErrorString, err
		}
	} else if exportType == Subtitle {
		execErrorString, err := f.execFFMPEG(subArgs)
		if err != nil {
			return execErrorString, err
		}
	} else {
		return "", errors.New("FFMPEGHelper ExportType not support")
	}

	return "", nil
}

// execFFMPEG 执行 ffmpeg 命令
func (f *FFMPEGHelper) execFFMPEG(cmds []string) (string, error) {

	if cmds == nil || len(cmds) == 0 {
		return "", nil
	}
	cmd := exec.Command("ffmpeg", cmds...)
	buf := bytes.NewBufferString("")
	//指定输出位置
	cmd.Stderr = buf
	cmd.Stdout = buf
	err := cmd.Start()
	if err != nil {
		return buf.String(), err
	}
	err = cmd.Wait()
	if err != nil {
		return buf.String(), err
	}

	return "", nil
}

// getAudioAndSubExportArgs 构建从原始视频导出字幕、音频的 ffmpeg 的参数 audioArgs, subArgs
func (f *FFMPEGHelper) getAudioAndSubExportArgs(videoFileFullPath string, ffmpegInfo *FFMPEGInfo) ([]string, []string) {

	/*
		导出多个字幕
		ffmpeg.exe -i xx.mp4 -vn -an -map 0:7 subs-7.srt -map 0:6 subs-6.srt
		导出音频，从 1m 27s 开始，导出向后的 28 s，转换为 mp3 格式
		ffmpeg.exe -i xx.mp4 -vn -map 0:1 -ss 00:1:27 -f mp3 -t 28 audio.mp3
		导出音频，转换为 mp3 格式
		ffmpeg.exe -i xx.mp4 -vn -map 0:1 -f mp3 audio.mp3
		导出音频，转换为 16000k 16bit 单通道 采样率的 test.pcm
		ffmpeg.exe -i xx.mp4 -vn -map 0:1 -ss 00:1:27 -t 28 -acodec pcm_s16le -f s16le -ac 1 -ar 16000 test.pcm
		截取字幕的时间片段
		ffmpeg.exe -i "subs-3.srt" -ss 00:1:27 -t 28 subs-3-cut-from-org.srt
	*/
	var subArgs = make([]string, 0)
	var audioArgs = make([]string, 0)
	// 基础的输入视频参数
	subArgs = append(subArgs, "-i")
	audioArgs = append(audioArgs, "-i")
	subArgs = append(subArgs, videoFileFullPath)
	audioArgs = append(audioArgs, videoFileFullPath)
	// 字幕导出的参数构建
	subArgs = append(subArgs, "-vn") // 不输出视频流
	subArgs = append(subArgs, "-an") // 不输出音频流

	nowCacheFolderPath, err := ffmpegInfo.GetCacheFolderFPath()
	if err != nil {
		f.log.Errorln("getAudioAndSubExportArgs", videoFileFullPath, err.Error())
		return nil, nil
	}

	if len(ffmpegInfo.SubtitleInfoList) == 0 {
		// 如果没有，就返回空
		subArgs = nil
	} else {
		for _, subtitleInfo := range ffmpegInfo.SubtitleInfoList {

			f.addSubMapArg(&subArgs, subtitleInfo.Index,
				filepath.Join(nowCacheFolderPath, subtitleInfo.GetName()+common.SubExtSRT))
			f.addSubMapArg(&subArgs, subtitleInfo.Index,
				filepath.Join(nowCacheFolderPath, subtitleInfo.GetName()+common.SubExtASS))
		}
	}

	// 音频导出的参数构建
	audioArgs = append(audioArgs, "-vn")
	if len(ffmpegInfo.AudioInfoList) == 0 {
		// 如果没有，就返回空
		audioArgs = nil
	} else {
		for _, audioInfo := range ffmpegInfo.AudioInfoList {
			f.addAudioMapArg(&audioArgs, audioInfo.Index,
				filepath.Join(nowCacheFolderPath, audioInfo.GetName()+extPCM))
		}
	}

	return audioArgs, subArgs
}

// getVideoExportArgsByTimeRange 导出某个时间范围内的视频 startTimeString 00:1:27 timeLeng 向后多少秒
func (f *FFMPEGHelper) getVideoExportArgsByTimeRange(videoFullPath string, startTimeString, timeLeng, outVideiFullPath string) []string {

	/*
		这个是用 to 到那个时间点
		ffmpeg.exe -i '.\Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.mp4' -ss 00:00:00 -to 00:05:00  -c:v copy -c:a copy wawa.mp4
		这个是向后多少秒
		ffmpeg.exe -i '.\Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.mp4' -ss 00:00:00 -t 300  -c:v copy -c:a copy wawa.mp4
	*/
	videoArgs := make([]string, 0)
	videoArgs = append(videoArgs, "-y")
	videoArgs = append(videoArgs, "-ss")
	videoArgs = append(videoArgs, startTimeString)
	videoArgs = append(videoArgs, "-t")
	videoArgs = append(videoArgs, timeLeng)

	// 解决开头黑屏问题
	videoArgs = append(videoArgs, "-accurate_seek")

	videoArgs = append(videoArgs, "-i")
	videoArgs = append(videoArgs, videoFullPath)

	//videoArgs = append(videoArgs, "-s")
	//videoArgs = append(videoArgs, "640x480")
	//videoArgs = append(videoArgs, "-vframes")
	//videoArgs = append(videoArgs, "90")
	//videoArgs = append(videoArgs, "-r")
	//videoArgs = append(videoArgs, "29.97")
	//videoArgs = append(videoArgs, "-c:v")
	//videoArgs = append(videoArgs, "h264")
	//videoArgs = append(videoArgs, "-b:v")
	//videoArgs = append(videoArgs, "500k")
	//videoArgs = append(videoArgs, "-b:a")
	//videoArgs = append(videoArgs, "48k")
	//videoArgs = append(videoArgs, "-ac")
	//videoArgs = append(videoArgs, "2")

	videoArgs = append(videoArgs, "-c:v")
	videoArgs = append(videoArgs, "copy")
	videoArgs = append(videoArgs, "-c:a")
	videoArgs = append(videoArgs, "copy")

	videoArgs = append(videoArgs, outVideiFullPath)

	return videoArgs
}

// getVideoHLSExportArgsByTimeRange 导出某个时间范围内的视频的 HLS 信息 startTimeString 00:1:27 timeLeng 向后多少秒
func (f *FFMPEGHelper) getVideoHLSExportArgsByTimeRange(videoFullPath string, startTimeString, timeLeng, sgmentTime, outVideiDirPath string) []string {

	/*
		ffmpeg.exe -i '111.mp4' -ss 00:00:00 -to 00:05:00  -c:v copy -c:a copy -f segment -segment_time 10 -segment_list outputlist.m3u8 -segment_format mpegts output%03d.ts
	*/

	videoArgs := make([]string, 0)

	videoArgs = append(videoArgs, "-ss")
	videoArgs = append(videoArgs, startTimeString)
	videoArgs = append(videoArgs, "-t")
	videoArgs = append(videoArgs, timeLeng)
	// 解决开头黑屏问题
	videoArgs = append(videoArgs, "-accurate_seek")

	videoArgs = append(videoArgs, "-i")
	videoArgs = append(videoArgs, videoFullPath)

	// 限制线程数
	videoArgs = append(videoArgs, "-threads")
	videoArgs = append(videoArgs, "2")
	// 约束强制贞切割?
	videoArgs = append(videoArgs, "-force_key_frames")
	videoArgs = append(videoArgs, "\"expr:gte(t,n_forced*"+sgmentTime+")\"")
	// 原编码格式
	videoArgs = append(videoArgs, "-c:v")
	videoArgs = append(videoArgs, "copy")
	videoArgs = append(videoArgs, "-c:a")
	videoArgs = append(videoArgs, "copy")
	// 转码为 h264
	//videoArgs = append(videoArgs, "-vcodec")
	//videoArgs = append(videoArgs, "h264")
	// -s 640x480 -vframes 90 -r 29.97 -c:v h264 -b:v 500k -b:a 48k -ac 2
	//videoArgs = append(videoArgs, "-s")
	//videoArgs = append(videoArgs, "640x480")
	//videoArgs = append(videoArgs, "-vframes")
	//videoArgs = append(videoArgs, "90")
	//videoArgs = append(videoArgs, "-r")
	//videoArgs = append(videoArgs, "29.97")
	//videoArgs = append(videoArgs, "-c:v")
	//videoArgs = append(videoArgs, "h264")
	//videoArgs = append(videoArgs, "-b:v")
	//videoArgs = append(videoArgs, "500k")
	//videoArgs = append(videoArgs, "-b:a")
	//videoArgs = append(videoArgs, "48k")
	//videoArgs = append(videoArgs, "-ac")
	//videoArgs = append(videoArgs, "2")

	videoArgs = append(videoArgs, "-f")
	videoArgs = append(videoArgs, "segment")
	videoArgs = append(videoArgs, "-segment_time")
	videoArgs = append(videoArgs, sgmentTime)
	videoArgs = append(videoArgs, "-segment_list")
	videoArgs = append(videoArgs, filepath.Join(outVideiDirPath, "outputlist.m3u8"))
	videoArgs = append(videoArgs, "-segment_format")
	videoArgs = append(videoArgs, "mpegts")
	videoArgs = append(videoArgs, filepath.Join(outVideiDirPath, "output%03d.ts"))

	return videoArgs
}

func (f *FFMPEGHelper) getVideo2HLSArgs(videoFullPath, segmentTime, outVideoDirPath string) []string {

	/*
		ffmpeg.exe -i '111.mp4' -c copy -map 0 -f segment -segment_list playlist.m3u8 -segment_time 10 -segment_format mpegts output%03d.ts
	*/

	videoArgs := make([]string, 0)
	videoArgs = append(videoArgs, "-i")
	videoArgs = append(videoArgs, videoFullPath)

	videoArgs = append(videoArgs, "-force_key_frames")
	videoArgs = append(videoArgs, "\"expr:gte(t,n_forced*"+segmentTime+")\"")

	videoArgs = append(videoArgs, "-c:v")
	videoArgs = append(videoArgs, "copy")
	videoArgs = append(videoArgs, "-c:a")
	videoArgs = append(videoArgs, "copy")

	videoArgs = append(videoArgs, "-f")
	videoArgs = append(videoArgs, "segment")
	videoArgs = append(videoArgs, "-segment_list")
	videoArgs = append(videoArgs, filepath.Join(outVideoDirPath, "playlist.m3u8"))
	videoArgs = append(videoArgs, "-segment_time")
	videoArgs = append(videoArgs, segmentTime)
	videoArgs = append(videoArgs, "-segment_format")
	videoArgs = append(videoArgs, "mpegts")
	videoArgs = append(videoArgs, filepath.Join(outVideoDirPath, "output%03d.ts"))

	return videoArgs
}

// getAudioAndSubExportArgsByTimeRange 导出某个时间范围内的音频和字幕文件文件 startTimeString 00:1:27 timeLeng 向后多少秒
func (f *FFMPEGHelper) getAudioExportArgsByTimeRange(audioFullPath string, startTimeString, timeLeng, outAudioFullPath string) []string {

	/*
		ffmpeg.exe -ar 16000 -ac 1 -f s16le -i aa.pcm -ss 00:1:27 -t 28 -acodec pcm_s16le -f s16le -ac 1 -ar 16000 bb.pcm

		ffmpeg.exe -i aa.srt -ss 00:1:27 -t 28 bb.srt
	*/

	var audioArgs = make([]string, 0)
	// 指定读取的音频文件编码格式
	audioArgs = append(audioArgs, "-ar")
	audioArgs = append(audioArgs, "16000")
	audioArgs = append(audioArgs, "-ac")
	audioArgs = append(audioArgs, "1")
	audioArgs = append(audioArgs, "-f")
	audioArgs = append(audioArgs, "s16le")

	audioArgs = append(audioArgs, "-i")
	audioArgs = append(audioArgs, audioFullPath)
	audioArgs = append(audioArgs, "-ss")
	audioArgs = append(audioArgs, startTimeString)
	audioArgs = append(audioArgs, "-t")
	audioArgs = append(audioArgs, timeLeng)

	// 指定导出的音频文件编码格式
	audioArgs = append(audioArgs, "-acodec")
	audioArgs = append(audioArgs, "pcm_s16le")
	audioArgs = append(audioArgs, "-f")
	audioArgs = append(audioArgs, "s16le")
	audioArgs = append(audioArgs, "-ac")
	audioArgs = append(audioArgs, "1")
	audioArgs = append(audioArgs, "-ar")
	audioArgs = append(audioArgs, "16000")

	audioArgs = append(audioArgs, outAudioFullPath)

	return audioArgs
}

func (f *FFMPEGHelper) getSubExportArgsByTimeRange(subFullPath string, startTimeString, timeLength, outSubFullPath string) []string {

	/*
		ffmpeg.exe -i aa.srt -ss 00:1:27 -t 28 bb.srt
	*/
	var subArgs = make([]string, 0)

	subArgs = append(subArgs, "-i")
	subArgs = append(subArgs, subFullPath)

	subArgs = append(subArgs, "-ss")
	subArgs = append(subArgs, startTimeString)
	subArgs = append(subArgs, "-t")
	subArgs = append(subArgs, timeLength)

	subArgs = append(subArgs, outSubFullPath)

	return subArgs
}

// getSubASS2SRTArgs 从 ASS 字幕 转到 SRT 字幕
func (f *FFMPEGHelper) getSubASS2SRTArgs(subFullPath, outSubFullPath string) []string {

	var subArgs = make([]string, 0)
	// 指定读取的音频文件编码格式
	subArgs = append(subArgs, "-i")
	subArgs = append(subArgs, subFullPath)
	subArgs = append(subArgs, outSubFullPath)

	return subArgs
}

// addSubMapArg 构建字幕的导出参数
func (f *FFMPEGHelper) addSubMapArg(subArgs *[]string, index int, subSaveFullPath string) {
	*subArgs = append(*subArgs, "-map")
	*subArgs = append(*subArgs, fmt.Sprintf("0:%d", index))
	*subArgs = append(*subArgs, subSaveFullPath)
}

// addAudioMapArg 构建音频的导出参数
func (f *FFMPEGHelper) addAudioMapArg(subArgs *[]string, index int, audioSaveFullPath string) {
	// -acodec pcm_s16le -f s16le -ac 1 -ar 16000
	*subArgs = append(*subArgs, "-map")
	*subArgs = append(*subArgs, fmt.Sprintf("0:%d", index))
	*subArgs = append(*subArgs, "-acodec")
	*subArgs = append(*subArgs, "pcm_s16le")
	*subArgs = append(*subArgs, "-f")
	*subArgs = append(*subArgs, "s16le")
	*subArgs = append(*subArgs, "-ac")
	*subArgs = append(*subArgs, "1")
	*subArgs = append(*subArgs, "-ar")
	*subArgs = append(*subArgs, "16000")
	*subArgs = append(*subArgs, audioSaveFullPath)
}

func (f *FFMPEGHelper) getVersion(exeName string) (string, error) {
	const args = "-version"
	cmdArgs := strings.Fields(args)
	cmd := exec.Command(exeName, cmdArgs...)
	buf := bytes.NewBufferString("")
	//指定输出位置
	cmd.Stderr = buf
	cmd.Stdout = buf
	err := cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// isSupportSubCodecName 是否是 FFMPEG 支持的 CodecName
func (f *FFMPEGHelper) isSupportSubCodecName(name string) bool {
	switch name {
	case Subtitle_StreamCodec_subrip,
		Subtitle_StreamCodec_ass,
		Subtitle_StreamCodec_ssa,
		Subtitle_StreamCodec_srt:
		return true
	default:
		return false
	}
}

func (f *FFMPEGHelper) GetVideoDuration(videoFileFullPath string) float64 {

	const args = "-v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 -i"
	cmdArgs := strings.Fields(args)
	cmdArgs = append(cmdArgs, videoFileFullPath)
	cmd := exec.Command("ffprobe", cmdArgs...)
	buf := bytes.NewBufferString("")
	//指定输出位置
	cmd.Stderr = buf
	cmd.Stdout = buf
	err := cmd.Start()
	if err != nil {
		return 0
	}
	err = cmd.Wait()
	if err != nil {
		return 0
	}

	// 字符串转 float64
	durationStr := strings.TrimSpace(buf.String())
	duration, err := strconv.ParseFloat(durationStr, 32)
	if err != nil {
		return 0
	}
	return duration
}

const (
	codecTypeSub   = "subtitle"
	codecTypeAudio = "audio"
	extMP3         = ".mp3"
	extPCM         = ".pcm"
)

type ExportType int

const (
	Subtitle         ExportType = iota // 导出字幕
	Audio                              // 导出音频
	SubtitleAndAudio                   // 导出字幕和音频
)

/*
	FFMPEG 支持的字幕 Codec Name：
	 ..S... arib_caption         ARIB STD-B24 caption
	 DES... ass                  ASS (Advanced SSA) subtitle (decoders: ssa ass ) (encoders: ssa ass )
	 DES... dvb_subtitle         DVB subtitles (decoders: dvbsub ) (encoders: dvbsub )
	 ..S... dvb_teletext         DVB teletext
	 DES... dvd_subtitle         DVD subtitles (decoders: dvdsub ) (encoders: dvdsub )
	 D.S... eia_608              EIA-608 closed captions (decoders: cc_dec )
	 D.S... hdmv_pgs_subtitle    HDMV Presentation Graphic Stream subtitles (decoders: pgssub )
	 ..S... hdmv_text_subtitle   HDMV Text subtitle
	 D.S... jacosub              JACOsub subtitle
	 D.S... microdvd             MicroDVD subtitle
	 DES... mov_text             MOV text
	 D.S... mpl2                 MPL2 subtitle
	 D.S... pjs                  PJS (Phoenix Japanimation Society) subtitle
	 D.S... realtext             RealText subtitle
	 D.S... sami                 SAMI subtitle
	 ..S... srt                  SubRip subtitle with embedded timing
	 ..S... ssa                  SSA (SubStation Alpha) subtitle
	 D.S... stl                  Spruce subtitle format
	 DES... subrip               SubRip subtitle (decoders: srt subrip ) (encoders: srt subrip )
	 D.S... subviewer            SubViewer subtitle
	 D.S... subviewer1           SubViewer v1 subtitle
	 DES... text                 raw UTF-8 text
	 ..S... ttml                 Timed Text Markup Language
	 D.S... vplayer              VPlayer subtitle
	 DES... webvtt               WebVTT subtitle
	 DES... xsub                 XSUB
*/
const (
	Subtitle_StreamCodec_subrip = "subrip"
	Subtitle_StreamCodec_ass    = "ass"
	Subtitle_StreamCodec_ssa    = "ssa"
	Subtitle_StreamCodec_srt    = "srt"
)
