package archive_helper

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"errors"
	"github.com/bodgit/sevenzip"
	"github.com/go-rod/rod/lib/utils"
	"github.com/mholt/archiver/v3"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// UnArchiveFile 7z 以外的都能搞定中文编码的问题，但是 7z 有梗，需要单独的库去解析，且编码是解决不了的，以后他们搞定了再测试
// 所以效果就是，7z 外的压缩包文件解压ok，字幕可以正常从名称解析出是简体还是繁体，但是7z就没办法了，一定乱码
func UnArchiveFile(fileFullPath, desRootPath string) error {
	switch filepath.Ext(strings.ToLower(fileFullPath)) {
	case ".zip":
		z := archiver.Zip{
			CompressionLevel:       flate.DefaultCompression,
			MkdirAll:               true,
			SelectiveCompression:   true,
			ContinueOnError:        false,
			OverwriteExisting:      true,
			ImplicitTopLevelFolder: false,
		}
		err := z.Walk(fileFullPath, func(f archiver.File) error {
			if f.IsDir() == true {
				return nil
			}
			zfh, ok := f.Header.(zip.FileHeader)
			if ok {
				err := processOneFile(f, zfh.NonUTF8, desRootPath)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	case ".tar":
		z := archiver.Tar{
			MkdirAll:               true,
			ContinueOnError:        false,
			OverwriteExisting:      true,
			ImplicitTopLevelFolder: false,
			StripComponents:        1,
		}
		err := z.Walk(fileFullPath, func(f archiver.File) error {
			if f.IsDir() == true {
				return nil
			}
			err := processOneFile(f, false, desRootPath)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	case ".rar":
		z := archiver.Rar{
			MkdirAll:               true,
			ContinueOnError:        false,
			OverwriteExisting:      true,
			ImplicitTopLevelFolder: false,
			StripComponents:        1,
		}
		err := z.Walk(fileFullPath, func(f archiver.File) error {
			if f.IsDir() == true {
				return nil
			}
			err := processOneFile(f, false, desRootPath)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}
	case ".7z":
		return unArr7z(fileFullPath, desRootPath)
	default:
		return errors.New("not support un archive file ext")
	}

	return nil
}

func processOneFile(f archiver.File, notUTF8 bool, desRootPath string) error {
	decodeName := f.Name()
	if notUTF8 == true {

		//ouBytes, err := ChangeFileCoding2UTF8([]byte(f.Name()))
		//if err != nil {
		//	return err
		//}
		i := bytes.NewReader([]byte(f.Name()))
		decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
		content, _ := ioutil.ReadAll(decoder)
		decodeName = string(content)
		//decodeName = string(ouBytes)
	}
	var chunk []byte
	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		//说明读取结束
		if n == 0 {
			break
		}
		//读取到最终的缓冲区中
		chunk = append(chunk, buf[:n]...)
	}
	err := utils.OutputFile(filepath.Join(desRootPath, decodeName), chunk)
	if err != nil {
		return err
	}
	return nil
}

func unArr7z(fileFullPath, desRootPath string) error {

	r, err := sevenzip.OpenReader(fileFullPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, file := range r.File {
		err = un7zOneFile(file, desRootPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func un7zOneFile(file *sevenzip.File, desRootPath string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}
	decodeName := file.Name
	decodeName = filepath.Base(decodeName)
	err = utils.OutputFile(filepath.Join(desRootPath, decodeName), data)
	if err != nil {
		return err
	}
	return nil
}

func IsWantedArchiveExtName(fileName string) bool {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".zip", ".tar", ".rar", "7z":
		return true
	default:
		return false
	}
}
