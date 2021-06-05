package common

import "errors"

var(
	NoMetadataFile = errors.New("no metadata file, movie.xml or *.nfo")
	CanNotFindIMDBID = errors.New("can not find IMDB Id")
	CIdIsEmpty = errors.New("cid is empty")
)
