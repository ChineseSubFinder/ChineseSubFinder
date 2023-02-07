package TestCode

import PTN "github.com/middelink/go-parse-torrent-name"

func getInfo() {
	//fileName := "1883 - S01E01 - 1883 WEBDL-1080p Proper.mkv"
	fileName := "1883 - S11E12 - 1883 WEBDL-1080p Proper.mkv"
	//fileName := "1883 - S01E02 - Behind Us, A Cliff WEBDL-1080p Proper.mkv"

	parse, err := PTN.Parse(fileName)
	if err != nil {
		return
	}

	println(parse.Season, parse.Episode)
}
