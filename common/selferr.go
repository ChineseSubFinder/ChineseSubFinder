package common

import "errors"

var(
	NoMetadataFile         = errors.New("no metadata file, movie.xml or *.nfo")
	CanNotFindIMDBID       = errors.New("can not find IMDB Id")
	XunLeiCIdIsEmpty       = errors.New("cid is empty")
	VideoFileIsTooSmall    = errors.New("video file is too small")
	ShooterFileHashIsEmpty = errors.New("filehash is empty")

	ZiMuKuSearchKeyWordStep1NotFound = errors.New("zimuku search keyword step1 not found")
	ZiMuKuDownloadUrlStep2NotFound = errors.New("zimuku download url step2 not found")
)
