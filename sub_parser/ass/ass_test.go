package ass

import (
	"testing"
)

func TestParser_DetermineFileType(t *testing.T) {

	filePath := "C:\\Tmp\\saw9.ass"
	parser := NewParser()
	_, _, err := parser.DetermineFileType(filePath)
	if err != nil {
		t.Fatal(err)
	}
}
