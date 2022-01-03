package vad

import (
	"testing"
	"path/filepath"
)

func TestGetVADInfo(t *testing.T) {

	var audioInfo = AudioInfo{
		FileFullPath: filepath.FromSlash("../../../TestData/ffmpeg/org/sampleAudio.wav"),
		// check below accordingly
		SampleRate:   16000,
		BitDepth:     16,
	}
	vadInfos, err := GetVADInfoFromAudio(audioInfo, false)
	if err != nil {
		t.Fatal(err)
	}

	println(vadInfos[0].Time.Milliseconds())
	println(vadInfos[1].Time.Milliseconds())
}
