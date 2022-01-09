package vad

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
)

func TestGetVADInfo(t *testing.T) {

	var audioInfo = AudioInfo{

		FileFullPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"ffmpeg", "org"}, 4, false), "sampleAudio.wav"),
		// check below accordingly
		SampleRate: 16000,
		BitDepth:   16,
	}
	vadInfos, err := GetVADInfoFromAudio(audioInfo, false)
	if err != nil {
		t.Fatal(err)
	}

	println(vadInfos[0].Time.Milliseconds())
	println(vadInfos[1].Time.Milliseconds())
}
