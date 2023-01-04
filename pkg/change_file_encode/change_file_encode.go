package change_file_encode

import (
	"errors"
	"fmt"
	"os"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
)

func Process(srcSubFileFPath string, desCode int) error {

	fBytes, err := os.ReadFile(srcSubFileFPath)
	if err != nil {
		return err
	}
	// 默认 0 是 UTF-8，1 是 GBK
	if desCode == 0 {
		// 0 是 UTF-8
		coding2UTF8, err := language.ChangeFileCoding2UTF8(fBytes)
		if err != nil {
			return err
		}
		err = os.WriteFile(srcSubFileFPath, coding2UTF8, os.ModePerm)
		if err != nil {
			return err
		}
	} else if desCode == 1 {
		// 1 是 GBK
		coding2UTF8, err := language.ChangeFileCoding2GBK(fBytes)
		if err != nil {
			return err
		}
		err = os.WriteFile(srcSubFileFPath, coding2UTF8, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("change_file_encode.Process(), not support encode type == %v", desCode))
	}

	return nil
}
