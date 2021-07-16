package pkg

import (
	"github.com/StalkR/imdb"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
)

// GetVideoInfoFromIMDB 从 IMDB ID 查询影片的信息
func GetVideoInfoFromIMDB(imdbID string, _reqParam ...types.ReqParam) (*imdb.Title, error) {
	var reqParam types.ReqParam
	if len(_reqParam) > 0 {
		reqParam = _reqParam[0]
	}
	t, err := imdb.NewTitle(NewHttpClient(reqParam).GetClient(), imdbID)
	if err != nil {
		Notify.Add("imdb model - imdb.NewTitle :", err.Error())
		return nil, err
	}

	return t, nil
}