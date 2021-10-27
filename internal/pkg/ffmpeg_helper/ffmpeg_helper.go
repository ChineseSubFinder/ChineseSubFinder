package ffmpeg_helper

import (
	"bytes"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/tidwall/gjson"
	"os"
	"os/exec"
	"path/filepath"

	"strings"
)

type FFMPEGHelper struct {
	subParserHub *sub_parser_hub.SubParserHub // 字幕内容的解析器
}

func NewFFMPEGHelper() *FFMPEGHelper {
	return &FFMPEGHelper{
		subParserHub: sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser()),
	}
}

// GetFFMPEGInfo 获取 视频的 FFMPEG 信息，包含音频和字幕
// 优先会导出 中、英、日、韩 类型的，字幕如果没有语言类型，则也导出，然后需要额外的字幕语言的判断去辅助标记（读取文件内容）
func (f *FFMPEGHelper) GetFFMPEGInfo(videoFileFullPath string) (bool, *FFMPEGInfo, error) {

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
	bok, ffMPEGInfo := f.parseJsonString2GetFFMPEGInfo(videoFileFullPath, buf.String())
	if bok == false {
		return false, nil, nil
	}
	// 在函数调用完毕后，判断是否需要清理
	defer func() {
		if bok == false && ffMPEGInfo != nil {
			err := os.RemoveAll(ffMPEGInfo.GetCacheFolderFPath())
			if err != nil {
				log_helper.GetLogger().Errorln("GetFFMPEGInfo - RemoveAll", err.Error())
				return
			}
		}
	}()

	// 判断这个视频是否已经导出过内置的字幕和音频文件了
	if ffMPEGInfo.IsExported() == false {
		// 说明缓存不存在，需要导出，这里需要注意，如果导出失败了，这个文件夹要清理掉
		if pkg.IsDir(ffMPEGInfo.GetCacheFolderFPath()) == true {
			// 如果存在则，先清空一个这个文件夹
			err = pkg.ClearFolder(ffMPEGInfo.GetCacheFolderFPath())
			if err != nil {
				bok = false
				return bok, nil, err
			}
		} else {
			// 如果不存在则，创建文件夹
			err = os.MkdirAll(ffMPEGInfo.GetCacheFolderFPath(), os.ModePerm)
			if err != nil {
				bok = false
				return bok, nil, err
			}
		}
		// 开始导出
		// 构建导出的命令参数
		subArgs, audioArgs := f.getAudioAndSubExportArgs(videoFileFullPath, ffMPEGInfo)
		// 执行导出
		execErrorString, err := f.exportAudioAndSubtitles(subArgs, audioArgs)
		if err != nil {
			log_helper.GetLogger().Errorln("exportAudioAndSubtitles", execErrorString)
			bok = false
			return bok, nil, err
		}
	}
	// 查找当前这个视频外置字幕列表
	err = ffMPEGInfo.GetExternalSubInfos(f.subParserHub)
	if err != nil {
		return false, nil, err
	}

	return bok, ffMPEGInfo, nil
}

// parseJsonString2GetFFMPEGInfo 使用 ffprobe 获取视频的 stream 信息，从中解析出字幕和音频的索引
func (f *FFMPEGHelper) parseJsonString2GetFFMPEGInfo(videoFileFullPath, inputFFProbeString string) (bool, *FFMPEGInfo) {

	streamsValue := gjson.Get(inputFFProbeString, "streams.#")
	if streamsValue.Exists() == false {
		return false, nil
	}

	ffmpegInfo := NewFFMPEGInfo(videoFileFullPath)

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
				continue
			}
			// 只导出 中、英、日、韩
			if language.IsSupportISOString(oneLanguage.String()) == false {
				continue
			}
			audioInfo := NewAudioInfo(int(oneIndex.Num), oneCodecName.String(), oneCodecType.String(),
				oneTimeBase.String(), oneStartTime.String(), oneLanguage.String())

			ffmpegInfo.AudioInfoList = append(ffmpegInfo.AudioInfoList, *audioInfo)

		} else {
			continue
		}
	}

	return true, ffmpegInfo
}

// exportAudioAndSubtitles 导出音频和字幕文件
func (f *FFMPEGHelper) exportAudioAndSubtitles(subArgs, audioArgs []string) (string, error) {

	// 这里导出依赖的是 ffmpeg 这个程序，需要的是构建导出的语句
	execErrorString, err := f.execFFMPEG(subArgs)
	if err != nil {
		return execErrorString, err
	}
	execErrorString, err = f.execFFMPEG(audioArgs)
	if err != nil {
		return execErrorString, err
	}

	return "", nil
}

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
	for _, subtitleInfo := range ffmpegInfo.SubtitleInfoList {
		f.addSubMapArg(&subArgs, subtitleInfo.Index,
			filepath.Join(ffmpegInfo.GetCacheFolderFPath(), subtitleInfo.GetName()+common.SubExtSRT))
		f.addSubMapArg(&subArgs, subtitleInfo.Index,
			filepath.Join(ffmpegInfo.GetCacheFolderFPath(), subtitleInfo.GetName()+common.SubExtASS))
	}
	// 音频导出的参数构建
	audioArgs = append(audioArgs, "-vn")
	for _, audioInfo := range ffmpegInfo.AudioInfoList {
		f.addAudioMapArg(&audioArgs, audioInfo.Index,
			filepath.Join(ffmpegInfo.GetCacheFolderFPath(), audioInfo.GetName()+extPCM))
	}

	return audioArgs, subArgs
}

func (f *FFMPEGHelper) addSubMapArg(subArgs *[]string, index int, subSaveFullPath string) {
	*subArgs = append(*subArgs, "-map")
	*subArgs = append(*subArgs, fmt.Sprintf("0:%d", index))
	*subArgs = append(*subArgs, subSaveFullPath)
}

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

const (
	codecTypeSub   = "subtitle"
	codecTypeAudio = "audio"
	extMP3         = ".mp3"
	extPCM         = ".pcm"
)
