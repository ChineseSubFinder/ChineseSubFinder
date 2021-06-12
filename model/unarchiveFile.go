package model

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/flate"
	"errors"
	"fmt"
	"github.com/gen2brain/go-unarr"
	"github.com/go-rod/rod/lib/utils"
	"github.com/mholt/archiver/v3"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func UnArchiveFile(fileFullPath, desRootPath string) error {
	switch filepath.Ext(strings.ToLower(fileFullPath)) {
	case ".zip":
		z := archiver.Zip{
			CompressionLevel:       flate.DefaultCompression,
			MkdirAll:               true,
			SelectiveCompression:   true,
			ContinueOnError:        false,
			OverwriteExisting:      false,
			ImplicitTopLevelFolder: false,
		}
		err := z.Walk(fileFullPath, func(f archiver.File) error {
			if f.IsDir() == true {
				return nil
			}
			zfh, ok := f.Header.(zip.FileHeader)
			if ok {
				isUTF8 := utf8.Valid([]byte(zfh.Name))
				if isUTF8 != zfh.NonUTF8 {
					println("the same")
				} else {
					println("not the same")
				}

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
			OverwriteExisting:      false,
			ImplicitTopLevelFolder: false,
		}
		err := z.Walk(fileFullPath, func(f archiver.File) error {
			if f.IsDir() == true {
				return nil
			}
			zfh, ok := f.Header.(tar.Header)
			if ok {
				err := processOneFile(f, utf8.Valid([]byte(zfh.Name)), desRootPath)
				if err != nil {
					return err
				}
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
			OverwriteExisting:      false,
			ImplicitTopLevelFolder: false,
		}
		err := z.Walk(fileFullPath, func(f archiver.File) error {
			if f.IsDir() == true {
				return nil
			}
			zfh, ok := f.Header.(tar.Header)
			if ok {
				err := processOneFile(f, utf8.Valid([]byte(zfh.Name)), desRootPath)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	default:
		return errors.New("not support un archive file ext")
	}

	return nil
}

func processOneFile(f archiver.File, notUTF8 bool, desRootPath string) error {
	detector := chardet.NewTextDetector()
	decodeName := f.Name()
	result, err := detector.DetectBest([]byte(decodeName))
	if err != nil {
		return err
	}
	fmt.Printf("Detected charset is %s, language is %s",
		result.Charset,
		result.Language)

	if notUTF8 == true {
		i := bytes.NewReader([]byte(f.Name()))
		decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
		content, _ := ioutil.ReadAll(decoder)
		decodeName = string(content)
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
	err = utils.OutputFile(path.Join(desRootPath, decodeName), chunk)
	if err != nil {
		return err
	}
	return nil
}

func UnArr(fileFullPath, desRootPath string) error {
	a, err := unarr.NewArchive(fileFullPath)
	if err != nil {
		return err
	}
	defer a.Close()
	for {
		err := a.Entry()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		data, err := a.ReadAll()
		if err != nil {
			return err
		}
		decodeName := a.Name()
		decodeName = filepath.Base(decodeName)
		err = utils.OutputFile(path.Join(desRootPath, decodeName), data)
		if err != nil {
			return err
		}
	}

	return nil
}