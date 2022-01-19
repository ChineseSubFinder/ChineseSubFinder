package imdb_helper

import (
	"github.com/StalkR/imdb"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"strings"
)

// GetVideoInfoFromIMDB 从 IMDB ID 查询影片的信息
func GetVideoInfoFromIMDB(imdbID string, _proxySettings ...settings.ProxySettings) (*imdb.Title, error) {
	var proxySettings settings.ProxySettings
	if len(_proxySettings) > 0 {
		proxySettings = _proxySettings[0]
	}
	t, err := imdb.NewTitle(my_util.NewHttpClient(proxySettings).GetClient(), imdbID)
	if err != nil {
		notify_center.Notify.Add("imdb model - imdb.NewTitle :", err.Error())
		return nil, err
	}

	return t, nil
}

// IsChineseVideo 从 imdbID 去查询判断是否是中文视频
func IsChineseVideo(imdbID string, _proxySettings ...settings.ProxySettings) (bool, *imdb.Title, error) {

	const chName0 = "chinese"
	const chName1 = "mandarin"

	var proxySettings settings.ProxySettings
	if len(_proxySettings) > 0 {
		proxySettings = _proxySettings[0]
	}

	t, err := GetVideoInfoFromIMDB(imdbID, proxySettings)
	if err != nil {
		return false, nil, err
	}

	if len(t.Languages) <= 0 {
		return false, nil, nil
	}
	firstLangLowCase := strings.ToLower(t.Languages[0])
	// 判断第一语言是否是中文
	switch firstLangLowCase {
	case chName0, chName1:
		return true, t, nil
	default:
		return false, t, nil
	}
}
