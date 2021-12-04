package ffmpeg_helper

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/tidwall/gjson"
	"os"
	"os/exec"
	"path/filepath"

	"strings"
)

type FFMPEGHelper struct {
	SubParserHub *sub_parser_hub.SubParserHub // 字幕内容的解析器
}

func NewFFMPEGHelper() *FFMPEGHelper {
	return &FFMPEGHelper{
		SubParserHub: sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser()),
	}
}

// Version 获取版本信息，如果不存在 FFMPEG 和 ffprobe 则报错
func (f FFMPEGHelper) Version() (string, error) {

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

// GetFFMPEGInfo 获取 视频的 FFMPEG 信息，包含音频和字幕
// 优先会导出 中、英、日、韩 类型的，字幕如果没有语言类型，则也导出，然后需要额外的字幕语言的判断去辅助标记（读取文件内容）
func (f *FFMPEGHelper) GetFFMPEGInfo(videoFileFullPath string, exportType ExportType) (bool, *FFMPEGInfo, error) {

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
	bok, ffMPEGInfo := f.parseJsonString2GetFFProbeInfo(videoFileFullPath, buf.String())
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
				log_helper.GetLogger().Errorln("GetFFMPEGInfo - RemoveAll", err.Error())
				return
			}
		}
	}()

	// 查找当前这个视频外置字幕列表
	err = ffMPEGInfo.GetExternalSubInfos(f.SubParserHub)
	if err != nil {
		return false, nil, err
	}

	// 判断这个视频是否已经导出过内置的字幕和音频文件了
	if ffMPEGInfo.IsExported(exportType) == false {
		// 说明缓存不存在，需要导出，这里需要注意，如果导出失败了，这个文件夹要清理掉
		if my_util.IsDir(nowCacheFolderPath) == true {
			// 如果存在则，先清空一个这个文件夹
			err = my_folder.ClearFolder(nowCacheFolderPath)
			if err != nil {
				bok = false
				return bok, nil, err
			}
		}
		// 开始导出
		// 构建导出的命令参数
		exportAudioArgs, exportSubArgs := f.getAudioAndSubExportArgs(videoFileFullPath, ffMPEGInfo)
		// 执行导出，音频和内置的字幕
		execErrorString, err := f.exportAudioAndSubtitles(exportAudioArgs, exportSubArgs, exportType)
		if err != nil {
			log_helper.GetLogger().Errorln("exportAudioAndSubtitles", execErrorString)
			bok = false
			return bok, nil, err
		}
	}

	return bok, ffMPEGInfo, nil
}

// GetAudioDurationInfo 获取音频的长度信息
func (f *FFMPEGHelper) GetAudioDurationInfo(audioFileFullPath string) (bool, float64, error) {

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
	if my_util.IsFile(outAudioFullPath) == true {
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
	if my_util.IsFile(outSubFullPath) == true {
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
	if my_util.IsFile(outSubFullPath) == true {
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

// parseJsonString2GetFFProbeInfo 使用 ffprobe 获取视频的 stream 信息，从中解析出字幕和音频的索引
func (f *FFMPEGHelper) parseJsonString2GetFFProbeInfo(videoFileFullPath, inputFFProbeString string) (bool, *FFMPEGInfo) {

	streamsValue := gjson.Get(inputFFProbeString, "streams.#")
	if streamsValue.Exists() == false {
		return false, nil
	}

	ffmpegInfo := NewFFMPEGInfo(videoFileFullPath)

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

			ffmpegInfo.SubtitleInfoList = append(ffmpegInfo.SubtitleInfoList, *subInfo)

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

			ffmpegInfo.AudioInfoList = append(ffmpegInfo.AudioInfoList, *audioInfo)

		} else {
			continue
		}
	}
	// 如何没有找到合适的字幕，那么就要把缓存的字幕选一个填充进去
	if len(ffmpegInfo.SubtitleInfoList) == 0 {
		if len(cacheSubtitleInfos) != 0 {
			ffmpegInfo.SubtitleInfoList = append(ffmpegInfo.SubtitleInfoList, cacheSubtitleInfos[0])
		}
	}
	// 如何没有找到合适的音频，那么就要把缓存的音频选一个填充进去
	if len(ffmpegInfo.AudioInfoList) == 0 {
		if len(cacheAudios) != 0 {
			ffmpegInfo.AudioInfoList = append(ffmpegInfo.AudioInfoList, cacheAudios[0])
		}
	}

	return true, ffmpegInfo
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
		log_helper.GetLogger().Errorln("getAudioAndSubExportArgs", videoFileFullPath, err.Error())
		return nil, nil
	}

	for _, subtitleInfo := range ffmpegInfo.SubtitleInfoList {
		f.addSubMapArg(&subArgs, subtitleInfo.Index,
			filepath.Join(nowCacheFolderPath, subtitleInfo.GetName()+common.SubExtSRT))
		f.addSubMapArg(&subArgs, subtitleInfo.Index,
			filepath.Join(nowCacheFolderPath, subtitleInfo.GetName()+common.SubExtASS))
	}
	// 音频导出的参数构建
	audioArgs = append(audioArgs, "-vn")
	for _, audioInfo := range ffmpegInfo.AudioInfoList {
		f.addAudioMapArg(&audioArgs, audioInfo.Index,
			filepath.Join(nowCacheFolderPath, audioInfo.GetName()+extPCM))
	}

	return audioArgs, subArgs
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

func (f FFMPEGHelper) getVersion(exeName string) (string, error) {
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
