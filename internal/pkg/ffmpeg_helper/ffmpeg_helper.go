package ffmpeg_helper

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"os/exec"

	"strings"
)

func GetSubTileIndexList(videoFileFullPath string) error {

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

	return nil
	//buf := bytes.NewBuffer(nil)
	//inFileName := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 5\\Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.mkv"
	//stream := ffmpeg.Input(inFileName).Output("-hide_banner").Filter("stream", "")
	//
	//println(stream.String())
}

func parseJsonString2GetAudioAndSubs(input string) error {

	streamsValue := gjson.Get(input, "streams.#")
	if streamsValue.Exists() == false {
		return nil
	}
	for i := 0; i < int(streamsValue.Num); i++ {

		oneIndex := gjson.Get(input, fmt.Sprintf("streams.%d.index", i))
		oneCodec_name := gjson.Get(input, fmt.Sprintf("streams.%d.codec_name", i))
		oneCodec_type := gjson.Get(input, fmt.Sprintf("streams.%d.codec_type", i))
		oneLanguage := gjson.Get(input, fmt.Sprintf("streams.%d.tags.language", i))

		// 任意一个字段不存在则跳过
		if oneIndex.Exists() == false || oneCodec_name.Exists() == false ||
			oneCodec_type.Exists() == false {
			continue
		}

		// 这里需要区分是字幕还是音频
		if oneCodec_type.String() != "subtitle" {
			// 字幕
			if oneLanguage.Exists() == false {
				continue
			}

		} else if oneCodec_type.String() != "subtitle" {
			// 音频

		} else {
			continue
		}

		// 不是字幕也跳过
		if oneCodec_name.String() != "subrip" || oneCodec_type.String() != "subtitle" {
			continue
		}
	}

	return nil
}
