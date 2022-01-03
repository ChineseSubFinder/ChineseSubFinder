package vosk_api

import (
	"testing"
	"path/filepath"
)

func TestGetResult(t *testing.T) {
	audioFPath := filepath.FromSlash("../../../TestData/ffmpeg/org/sampleAudio.wav")
	err := GetResult(audioFPath)
	if err != nil {
		t.Fatal(err)
	}
}
