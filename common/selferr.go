package common

import "errors"

var(
	NoMetadataFile         = errors.New("no metadata file, movie.xml or *.nfo")
	CanNotFindIMDBID       = errors.New("can not find IMDB Id")
	CanNotFindEpAiredTime       = errors.New("can not find Ep Aired Time")
	XunLeiCIdIsEmpty       = errors.New("cid is empty")
	VideoFileIsTooSmall    = errors.New("video file is too small")
	ShooterFileHashIsEmpty = errors.New("filehash is empty")

	ZiMuKuSearchKeyWordStep0DetailPageUrlNotFound = errors.New("zimuku search keyword step0 not found, detail page url")
	ZiMuKuDownloadUrlStep2NotFound = errors.New("zimuku download url step2 not found")
	ZiMuKuDownloadUrlStep3NotFound = errors.New("zimuku download url step3 not found")
	ZiMuKuDownloadUrlStep3AllFailed = errors.New("zimuku download url step3 all failed")

	SubHDStep0SubCountNotFound = errors.New("subhd step0 sub count not found")
	SubHDStep0HrefIsNull = errors.New("subhd step0 href is Null")
	SubHDStep2SidIsNull = errors.New("subhd step2 sid is null")
	SubHDStep2DTokenIsNull = errors.New("subhd step2 dToken is null")
	SubHDStep2ResultIsNullOrNotTrue = errors.New("subhd step2 result is null or not true")
	SubHDStep2PostResultGetUrlNotFound= errors.New("subhd step2 post result get url not found")
	SubHDStep2ExCannotFindDownloadBtn= errors.New("subhd step2ex can not find download btn")
)
