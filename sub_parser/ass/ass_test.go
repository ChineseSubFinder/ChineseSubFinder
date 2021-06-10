package ass

import (
	"testing"
)

func TestParser_DetermineFileType(t *testing.T) {

	filePath := "C:\\Tmp\\saw9.ass"
	parser := NewParser()
	sfi, err := parser.DetermineFileTypeFromFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	println(sfi.Name, sfi.Lang.String(), sfi.Ext)
}
