package ffmpeg_helper

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"os/exec"
	"path/filepath"

	"strings"
)

func GetFFMPEGInfo(videoFileFullPath string) error {

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
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}

	// 将获取到的反馈 json 字符串进行解析
	println(buf.String())
	println(filepath.Dir(videoFileFullPath))

	return nil
}

// parseJsonString2GetFFMPEGInfo 使用 ffprobe 获取视频的 stream 信息，从中解析出字幕和音频的索引
func parseJsonString2GetFFMPEGInfo(videoFileFullPath, inputFFProbeString string) (bool, *FFMPEGInfo) {

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
		if oneCodecType.String() == "subtitle" {
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
			}
			subInfo := NewSubtitileInfo(int(oneIndex.Num), oneCodecName.String(), oneCodecType.String(),
				oneTimeBase.String(), oneStartTime.String(),
				int(oneDurationTS.Num), oneDuration.String(), nowLanguageString)

			ffmpegInfo.SubtitleInfoList = append(ffmpegInfo.SubtitleInfoList, *subInfo)

		} else if oneCodecType.String() == "audio" {
			// 音频
			// 这里必要要能够解析到 language 字段
			if oneLanguage.Exists() == false {
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
func exportAudioAndSubtitles(videoFileFullPath string, ffmpegInfo *FFMPEGInfo) error {

	// 这里导出依赖的是 ffmpeg 这个程序，需要的是构建导出的语句
	return nil
}

func getAudioAndSubExportArgs(videoFileFullPath string, ffmpegInfo *FFMPEGInfo) ([]string, []string) {

	/*
		导出多个字幕
		ffmpeg.exe -i xx.mp4 -vn -an -map 0:7 subs-7.srt -map 0:6 subs-6.srt
		导出音频，从 1m 27s 开始，导出向后的 28 s，转换为 mp3 格式
		ffmpeg.exe -i xx.mp4 -vn -map 0:1 -ss 00:1:27 -f mp3 -t 28 audio.mp3
		导出音频，转换为 mp3 格式
		ffmpeg.exe -i xx.mp4 -vn -map 0:1 -f mp3 audio.mp3
	*/
	var subArgs = make([]string, 0)
	var audioArgs = make([]string, 0)

	subArgs = append(subArgs, "-i")
	audioArgs = append(audioArgs, "-i")

	subArgs = append(subArgs, videoFileFullPath)
	audioArgs = append(audioArgs, videoFileFullPath)

	subArgs = append(subArgs, "-vn")
	subArgs = append(subArgs, "-an")
	for _, subtitleInfo := range ffmpegInfo.SubtitleInfoList {
		addSubMapArg(subArgs, subtitleInfo.Index, filepath.Join())
	}
}

func addSubMapArg(subArgs []string, index int, subSaveFullPath string) {
	subArgs = append(subArgs, "-map")
	subArgs = append(subArgs, fmt.Sprintf("0:%d", index))
	subArgs = append(subArgs, subSaveFullPath)
}
