package shooter

import (
	"crypto/md5"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Supplier struct {
	reqParam common.ReqParam
	log      *logrus.Logger
	topic    int
}

func NewSupplier(_reqParam ...common.ReqParam) *Supplier {

	sup := Supplier{}
	sup.log = model.GetLogger()
	sup.topic = common.DownloadSubsPerSite
	if len(_reqParam) > 0 {
		sup.reqParam = _reqParam[0]
		if sup.reqParam.Topic > 0 && sup.reqParam.Topic != sup.topic {
			sup.topic = sup.reqParam.Topic
		}
	}
	return &sup
}

func (s Supplier) GetSupplierName() string {
	return common.SubSiteShooter
}

func (s Supplier) GetSubListFromFile4Movie(filePath string) ([]common.SupplierSubInfo, error){
	return s.GetSubListFromFile(filePath)
}

func (s Supplier) GetSubListFromFile4Series(filePath string) ([]common.SupplierSubInfo, error) {
	return s.GetSubListFromFile(filePath)
}

func (s Supplier) GetSubListFromFile4Anime(filePath string) ([]common.SupplierSubInfo, error){
	return s.GetSubListFromFile(filePath)
}

func (s Supplier) GetSubListFromFile(filePath string) ([]common.SupplierSubInfo, error) {

	// 可以提供的字幕查询 eng或者chn
	const qLan = "Chn"
	var outSubInfoList []common.SupplierSubInfo
	var jsonList []SublistShooter

	hash, err := s.ComputeFileHash(filePath)
	if err != nil {
		return nil, err
	}
	if hash == "" {
		return nil, common.ShooterFileHashIsEmpty
	}

	fileName := filepath.Base(filePath)

	httpClient := model.NewHttpClient(s.reqParam)

	_, err = httpClient.R().
		SetFormData(map[string]string{
			"filehash": hash,
			"pathinfo": fileName,
			"format": "json",
			"lang": qLan,
		}).
		SetResult(&jsonList).
		Post(common.SubShooterRootUrl)
	if err != nil {
		return nil, err
	}
	for i, shooter := range jsonList {
		for _, file := range shooter.Files {
			subExt := file.Ext
			if strings.Contains(file.Ext, ".") == false {
				subExt = "." + subExt
			}

			data, _, err := model.DownFile(file.Link)
			if err != nil {
				s.log.Error(err)
				continue
			}
			outSubInfoList = append(outSubInfoList, *common.NewSupplierSubInfo(s.GetSupplierName(), int64(i), fileName, common.ChineseSimple, file.Link, 0, shooter.Delay, subExt, data))
			// 如果够了那么多个字幕就返回
			if len(outSubInfoList) >= s.topic {
				return outSubInfoList, nil
			}
			// 一层里面，下载一个文件就行了
			break
		}
	}
	return outSubInfoList, nil
}

func (s Supplier) GetSubListFromKeyword(keyword string) ([]common.SupplierSubInfo, error) {
	panic("not implemented")
}

func (s Supplier) ComputeFileHash(filePath string) (string, error) {
	hash := ""
	fp, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fp.Close()
	stat, err := fp.Stat()
	if err != nil {
		return "", err
	}
	size := float64(stat.Size())
	if size < 0xF000 {
		return "", common.VideoFileIsTooSmall
	}
	samplePositions := [4]int64{
		4 * 1024,
		int64(math.Floor(size / 3 * 2)),
		int64(math.Floor(size / 3)),
		int64(size - 8*1024)}
	var samples [4][]byte
	for i, position := range samplePositions {
		samples[i] = make([]byte, 4*1024)
		_, err = fp.ReadAt(samples[i], position)
		if err != nil {
			return "", err
		}
	}
	for _, sample := range samples {
		if len(hash) > 0 {
			hash += ";"
		}
		hash += fmt.Sprintf("%x", md5.Sum(sample))
	}

	return hash, nil
}


type FilesShooter struct {
	Ext  string `json:"ext"`
	Link string `json:"link"`
}
type SublistShooter struct {
	Desc  string         `json:"desc"`
	Delay int64          `json:"delay"`
	Files []FilesShooter `json:"files"`
}