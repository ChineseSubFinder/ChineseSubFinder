package vad

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	webRTCVAD "github.com/baabaaox/go-webrtcvad"
	"io"
	"os"
	"time"
)

// GetVADInfoFromAudio 分析音频文件，得到 VAD 分析信息，看样子是不支持并发的，只能单线程使用
// 无需使用插值的函数
func GetVADInfoFromAudio(audioInfo AudioInfo) ([]VADInfo, error) {

	var (
		frameIndex  = 0
		frameSize   = audioInfo.SampleRate / 1000 * FrameDuration
		frameBuffer = make([]byte, audioInfo.SampleRate/1000*FrameDuration*audioInfo.BitDepth/8)
		frameActive = false
		vadInfos    = make([]VADInfo, 0)
	)

	audioFile, err := os.Open(audioInfo.FileFullPath)
	if err != nil {
		return nil, err
	}
	defer audioFile.Close()

	reader := bufio.NewReader(audioFile)

	vadInst := webRTCVAD.Create()
	defer webRTCVAD.Free(vadInst)

	err = webRTCVAD.Init(vadInst)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	err = webRTCVAD.SetMode(vadInst, Mode)
	if err != nil {
		return nil, err
	}

	if ok := webRTCVAD.ValidRateAndFrameLength(audioInfo.SampleRate, frameSize); !ok {
		return nil, errors.New(fmt.Sprintf("invalid rate or frame length, %v", audioInfo.FileFullPath))
	}
	var offset int

	report := func() {
		t := time.Duration(offset) * time.Second / time.Duration(audioInfo.SampleRate) / 2
		//log.Printf("Frame: %v, offset: %v, Active: %v, t = %v", frameIndex, offset, frameActive, t)
		vadInfos = append(vadInfos, *NewVADInfo(
			frameIndex,
			offset,
			frameActive,
			t,
		))
	}

	for {
		_, err = io.ReadFull(reader, frameBuffer)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		tmpFrameActive, err := webRTCVAD.Process(vadInst, audioInfo.SampleRate, frameBuffer, frameSize)
		if err != nil {
			return nil, err
		}
		if tmpFrameActive != frameActive || offset == 0 {
			frameActive = tmpFrameActive
		}
		report()
		offset += len(frameBuffer)
		frameIndex++
	}

	report()

	return vadInfos, nil
}

// GetVADInfoFromSubtitle 分析字幕文件(暂时考虑的是外置的字幕)，得到 VAD 分析信息，看样子是不支持并发的，只能单线程使用
func GetVADInfoFromSubtitle(subFileInfo *subparser.FileInfo, startTime, endIndex int) ([]VADInfo, error) {

	var vadInfos = make([]VADInfo, 0)
	timeFormat := subFileInfo.GetTimeFormat()
	println(timeFormat)
	for _, oneDialogueEx := range subFileInfo.DialoguesEx {

		// 考虑的是外置字幕，所以就应该是有中文的
		if oneDialogueEx.ChLine == "" {
			continue
		}

	}

	return vadInfos, nil
}
