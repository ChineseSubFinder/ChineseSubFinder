package TestCode

import "testing"

func TestDownloadTest(t *testing.T) {

	err := DownloadTest()
	if err != nil {
		t.Fatal(err)
	}
}
