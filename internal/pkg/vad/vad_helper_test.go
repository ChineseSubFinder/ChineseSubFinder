package vad

import (
	"testing"
)

func TestGetVADInfo(t *testing.T) {

	var audioInfo = AudioInfo{
		FileFullPath: "C:\\Tmp\\Rick and Morty - S05E10\\英_1.pcm",
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
