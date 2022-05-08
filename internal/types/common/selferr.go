package common

import (
	"errors"
	"fmt"
)

var (
	NoMetadataFile         = errors.New("no metadata file, movie.xml or *.nfo")
	CanNotFindIMDBID       = errors.New("can not find IMDB Id")
	CanNotFindEpAiredTime  = errors.New("can not find Ep Aired Time")
	XunLeiCIdIsEmpty       = errors.New("cid is empty")
	VideoFileIsTooSmall    = errors.New("video file is too small")
	ShooterFileHashIsEmpty = errors.New("filehash is empty")

	ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound = errors.New("zimuku search keyword step0 not found, detail page url")
	ZiMuKuDownloadUrlStep2NotFound                = errors.New("zimuku download url step2 not found")
	ZiMuKuDownloadUrlStep3NotFound                = errors.New("zimuku download url step3 not found")
	ZiMuKuDownloadUrlDownFileFailed               = errors.New("zimuku download url DownFile failed")

	SubHDStep0SubCountElementNotFound = errors.New("subhd step0 sub count element not found")
	SubHDStep0ImgParentLessThan1      = errors.New("subhd step0 Img Parent less than 1")
	SubHDStep0HrefIsNull              = errors.New("subhd step0 href is Null")
	SubHDStep2ExCannotFindDownloadBtn = errors.New("subhd step2ex can not find download btn")
	SubHDDownloadUrlDownFileFailed    = errors.New("subhd download url DownFile failed")
)

var (
	AllSiteDownloadSubNotFound = errors.New("all site download sub not found")
	SkipCreateInDB             = fmt.Errorf("skip create in db")
)
