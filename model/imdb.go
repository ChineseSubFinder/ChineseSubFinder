package model

import (
	"github.com/StalkR/imdb"
	"net/http"
)

// GetVideoInfoFromIMDB 从 IMDB ID 查询影片的信息
func GetVideoInfoFromIMDB(imdbID string) (*imdb.Title, error) {
	client := http.DefaultClient
	t, err := imdb.NewTitle(client, imdbID)
	if err != nil {
		return nil, err
	}
	return t, nil
}