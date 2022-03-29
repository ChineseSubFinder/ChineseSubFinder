package something_static

import (
	b64 "encoding/base64"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"os"
	"path/filepath"
)

func WriteFile(CloneProjectDesSaveDir, enString, nowTime, nowTimeFileNamePrix string) (bool, error) {

	saveFileFPath := filepath.Join(CloneProjectDesSaveDir, nowTimeFileNamePrix+common.StaticFileName00)
	saveFileFPathWait := filepath.Join(CloneProjectDesSaveDir, nowTimeFileNamePrix+common.StaticFileName00+waitExt)

	if my_util.IsFile(saveFileFPath) == true {
		// 目标文件存在，则需要判断准备写入覆盖的文件是否与当前存在的文件 SHA1 的值是一样的，一样就跳过后续的操作
		// 写入等待替换的文件
		err := writeFile(saveFileFPathWait, enString, nowTime)
		if err != nil {
			return false, err
		}
		orgFileSHA1, err := my_util.GetFileSHA1(saveFileFPath)
		if err != nil {
			return false, err
		}
		waitFileSHA1, err := my_util.GetFileSHA1(saveFileFPathWait)
		if err != nil {
			return false, err
		}
		// 如果一样的，那么外面就需要判断无需继续往下执行
		if orgFileSHA1 == waitFileSHA1 {
			// 删除 wait 文件
			err = os.Remove(saveFileFPathWait)
			if err != nil {
				return false, err
			}
			return false, nil
		}
		// 如果不一样，那么就需要删除原来的文件，然后把 wait 文件 rename 过去
		err = os.Remove(saveFileFPath)
		if err != nil {
			return false, err
		}
		err = os.Rename(saveFileFPathWait, saveFileFPath)
		if err != nil {
			return false, err
		}
	} else {
		// 如果不存在，那么就直接写入就行了
		err := writeFile(saveFileFPath, enString, nowTime)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func writeFile(saveFileFPath, enString, nowTime string) error {

	file, err := os.Create(saveFileFPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = file.WriteString(enString + b64.StdEncoding.EncodeToString([]byte(nowTime)))
	if err != nil {
		return err
	}

	return nil
}

func GetCodeFromWeb(nowTimeFileNamePrix string) (string, string, error) {
	const baseCodeFileUrl = "https://cdn.jsdelivr.net/gh/"
	const whichProject = "allanpk716/SomeThingsStatic/"

	desUrl := baseCodeFileUrl + whichProject + nowTimeFileNamePrix + common.StaticFileName00

	fileBytes, _, err := my_util.DownFile(desUrl)
	if err != nil {
		return "", "", err
	}

	timeB64String := fileBytes[len(fileBytes)-16:]
	decodedTime, err := b64.StdEncoding.DecodeString(string(timeB64String))
	if err != nil {
		return "", "", err
	}
	decodeTimeStr := string(decodedTime)

	codeB64String := fileBytes[:len(fileBytes)-16]
	decodedCode, err := b64.StdEncoding.DecodeString(string(codeB64String))
	if err != nil {
		return "", "", err
	}
	decodeCodeStr := string(decodedCode)

	return decodeTimeStr, decodeCodeStr, nil
}

const waitExt = ".wait"
