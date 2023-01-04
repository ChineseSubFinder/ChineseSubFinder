package TestCode

import (
	"fmt"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	tmdb "github.com/cyruzin/golang-tmdb"
)

func imdb2tmdb() {
	tmdbClient, err := tmdb.Init("xxx")

	if err != nil {
		fmt.Println(err)
	}

	options := make(map[string]string)
	options["external_source"] = "imdb_id"
	//options["language"] = "en-US"
	options["language"] = "zh-CN"

	const keanuReevesID = "tt6264654"

	proxySettings := settings.NewProxySettings(true, "http",
		"19037",
		"192.168.50.252", "5269",
		"", "",
	)

	restyClient, err := pkg.NewHttpClient(proxySettings)

	tmdbClient.SetClientConfig(*restyClient.GetClient())

	result, err := tmdbClient.GetFindByID(keanuReevesID, options)
	if err != nil {
		fmt.Println(err)
		return
	}

	println(result.MovieResults[0].Title)
}
