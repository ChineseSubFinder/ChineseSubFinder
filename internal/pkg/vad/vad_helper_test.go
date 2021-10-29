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
	vadInfo, err := GetVADInfoFromAudio(audioInfo)
	if err != nil {
		t.Fatal(err)
	}

	println(len(vadInfo))
}
