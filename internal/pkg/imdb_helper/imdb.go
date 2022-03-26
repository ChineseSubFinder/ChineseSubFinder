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

// GetVideoIMDBInfoFromLocal 从本地获取 IMDB 信息，如果找不到则去网络获取并写入本地缓存
func GetVideoIMDBInfoFromLocal(imdbID string, _proxySettings ...settings.ProxySettings) (*models.IMDBInfo, error) {

	var proxySettings settings.ProxySettings
	if len(_proxySettings) > 0 {
		proxySettings = _proxySettings[0]
	}

	// 首先从数据库中查找是否存在这个 IMDB 信息，如果不存在再使用 Web 查找，且写入数据库
	var imdbInfos []models.IMDBInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().Preload("VideoSubInfos").Limit(1).Where(&models.IMDBInfo{IMDBID: imdbID}).Find(&imdbInfos)

	if len(imdbInfos) <= 0 {
		// 没有找到，去网上获取
		t, err := GetVideoInfoFromIMDB(imdbID, proxySettings)
		if err != nil {
			return nil, err
		}
		// 存入数据库
		nowIMDBInfo := models.NewIMDBInfo(imdbID, t.Name, t.Year, t.Description, t.Languages, t.AKA)
		imdbInfos = make([]models.IMDBInfo, 0)
		imdbInfos = append(imdbInfos, *nowIMDBInfo)
		dao.GetDb().Create(nowIMDBInfo)

		return nowIMDBInfo, nil
	} else {
		// 找到
		return &imdbInfos[0], nil
	}
}

// IsChineseVideo 从 imdbID 去查询判断是否是中文视频
func IsChineseVideo(imdbID string, _proxySettings ...settings.ProxySettings) (bool, *models.IMDBInfo, error) {

	const chName0 = "chinese"
	const chName1 = "mandarin"

	var proxySettings settings.ProxySettings
	if len(_proxySettings) > 0 {
		proxySettings = _proxySettings[0]
	}

	localIMDBInfo, err := GetVideoIMDBInfoFromLocal(imdbID, proxySettings)
	if err != nil {
		return false, nil, err
	}
	if len(localIMDBInfo.Languages) <= 0 {
		return false, nil, nil
	}
	firstLangLowCase := strings.ToLower(localIMDBInfo.Languages[0])
	// 判断第一语言是否是中文
	switch firstLangLowCase {
	case chName0, chName1:
		return true, localIMDBInfo, nil
	default:
		return false, localIMDBInfo, nil
	}
}
