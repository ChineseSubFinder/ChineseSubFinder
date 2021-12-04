package vad

import (
	"bufio"
	"errors"
	"fmt"
	webRTCVAD "github.com/baabaaox/go-webrtcvad"
	"io"
	"os"
	"time"
)

// GetVADInfoFromAudio 分析音频文件，得到 VAD 分析信息，看样子是不支持并发的，只能单线程使用
// 无需使用插值的函数
func GetVADInfoFromAudio(audioInfo AudioInfo, insert bool) ([]VADInfo, error) {

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
			if insert == false {
				report()
			}
		}
		if insert == true {
			report()
		}
		offset += len(frameBuffer)
		frameIndex++
	}

	report()

	return vadInfos, nil
}

// GetFloatSlice 返回 1 -1 归一化的数组
func GetFloatSlice(inVADs []VADInfo) []float64 {
	outVADFloats := make([]float64, len(inVADs))
	for i, vad := range inVADs {
		if vad.Active == true {
			outVADFloats[i] = 1
		} else {
			outVADFloats[i] = -1
		}
	}

	return outVADFloats
}

// GetAudioIndex2Time 从 Audio 的 OffsetIndex 推算出它所在的时间，返回 float64 的秒
func GetAudioIndex2Time(index int) float64 {
	return float64(index*FrameDuration) / 1000.0
}

const (
	// Mode vad mode，VAD 的模式 0-3
	Mode = 3
	// FrameDuration frame duration，分析的时间窗口
	FrameDuration = 10
)
