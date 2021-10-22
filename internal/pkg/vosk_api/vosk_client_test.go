package vosk_api

import "testing"

func TestGetResult(t *testing.T) {
	audioFPath := "C:\\Tmp\\audio.wav"
	err := GetResult(audioFPath)
	if err != nil {
		t.Fatal(err)
	}
}
