package imdb_helper

import (
	"github.com/StalkR/imdb"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
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
func IsChineseVideo(imdbID string, _proxySettings ...settings.ProxySettings) (bool, *models.IMDBInfo, error) {

	const chName0 = "chinese"
	const chName1 = "mandarin"

	var proxySettings settings.ProxySettings
	if len(_proxySettings) > 0 {
		proxySettings = _proxySettings[0]
	}

	// 首先从数据库中查找是否存在这个 IMDB 信息，如果不存在再使用 Web 查找，且写入数据库
	var imdbInfos []models.IMDBInfo
	dao.GetDb().Limit(1).Where(&models.IMDBInfo{IMDBID: imdbID}).Find(&imdbInfos)

	var firstLangLowCase string
	if len(imdbInfos) <= 0 {
		// 没有找到
		t, err := GetVideoInfoFromIMDB(imdbID, proxySettings)
		if err != nil {
			return false, nil, err
		}
		// 存入数据库
		nowIMDBInfo := models.NewIMDBInfo(imdbID, t.Name, t.Year, t.Description, t.Languages, t.AKA)
		imdbInfos = make([]models.IMDBInfo, 0)
		imdbInfos = append(imdbInfos, *nowIMDBInfo)
		dao.GetDb().Create(nowIMDBInfo)

		if len(t.Languages) <= 0 {
			return false, nil, nil
		}

		firstLangLowCase = strings.ToLower(t.Languages[0])

	} else {
		// 找到
		if len(imdbInfos[0].Languages) <= 0 {
			return false, nil, nil
		}

		firstLangLowCase = strings.ToLower(imdbInfos[0].Languages[0])
	}
	// 判断第一语言是否是中文
	switch firstLangLowCase {
	case chName0, chName1:
		return true, &imdbInfos[0], nil
	default:
		return false, &imdbInfos[0], nil
	}
}
