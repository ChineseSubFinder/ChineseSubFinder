package strcut_json

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// ToFile 注意传入的不是指针
func ToFile(srcJsonFileFPath string, input interface{}) error {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.FromSlash(srcJsonFileFPath))
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.Write(jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

// ToStruct 传入的必须是指针
func ToStruct(desJsonFileFPath string, output interface{}) error {

	file, err := os.Open(filepath.FromSlash(desJsonFileFPath))
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, output)
	if err != nil {
		return err
	}

	return nil
}
