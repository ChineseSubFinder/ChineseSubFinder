package imdb_helper

import (
	"github.com/StalkR/imdb"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"strconv"
	"strings"
)

// GetVideoInfoFromIMDBWeb 从 IMDB 网站 ID 查询影片的信息
func GetVideoInfoFromIMDBWeb(imdbInfo types.VideoIMDBInfo, _proxySettings ...*settings.ProxySettings) (*imdb.Title, error) {

	client, err := my_util.NewHttpClient(_proxySettings...)
	if err != nil {
		return nil, err
	}

	t, err := imdb.NewTitle(client.GetClient(), imdbInfo.ImdbId)
	if err != nil {
		notify_center.Notify.Add("imdb model - imdb.NewTitle :", err.Error())
		return nil, err
	}
	if t.Year == 0 {
		// IMDB 信息获取的库(1.0.7)，目前有bug，比如，tt6856242 年份为 0
		if imdbInfo.Year != "" {
			year, err := strconv.Atoi(imdbInfo.Year)
			if err != nil {
				return nil, err
			}
			t.Year = year
		}
	}

	return t, nil
}

// GetVideoIMDBInfoFromLocal 从本地获取 IMDB 信息，如果找不到则去网络获取并写入本地缓存
func GetVideoIMDBInfoFromLocal(imdbInfo types.VideoIMDBInfo, _proxySettings ...*settings.ProxySettings) (*models.IMDBInfo, error) {

	log_helper.GetLogger().Debugln("GetVideoIMDBInfoFromLocal", 0)

	// 首先从数据库中查找是否存在这个 IMDB 信息，如果不存在再使用 Web 查找，且写入数据库
	var imdbInfos []models.IMDBInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().
		Preload("VideoSubInfos").
		Limit(1).Where(&models.IMDBInfo{IMDBID: imdbInfo.ImdbId}).Find(&imdbInfos)

	log_helper.GetLogger().Debugln("GetVideoIMDBInfoFromLocal", 1)

	if len(imdbInfos) <= 0 {
		// 没有找到，去网上获取
		t, err := GetVideoInfoFromIMDBWeb(imdbInfo, _proxySettings...)
		if err != nil {
			return nil, err
		}
		log_helper.GetLogger().Debugln("GetVideoIMDBInfoFromLocal", 2)

		// 存入数据库
		nowIMDBInfo := models.NewIMDBInfo(imdbInfo.ImdbId, t.Name, t.Year, t.Description, t.Languages, t.AKA)
		imdbInfos = make([]models.IMDBInfo, 0)
		imdbInfos = append(imdbInfos, *nowIMDBInfo)
		dao.GetDb().Create(nowIMDBInfo)

		log_helper.GetLogger().Debugln("GetVideoIMDBInfoFromLocal", 3)

		return nowIMDBInfo, nil
	} else {

		log_helper.GetLogger().Debugln("GetVideoIMDBInfoFromLocal", 4)
		// 找到
		return &imdbInfos[0], nil
	}
}

// IsChineseVideo 从 imdbID 去查询判断是否是中文视频
func IsChineseVideo(imdbInfo types.VideoIMDBInfo, _proxySettings ...*settings.ProxySettings) (bool, *models.IMDBInfo, error) {

	const chName0 = "chinese"
	const chName1 = "mandarin"

	localIMDBInfo, err := GetVideoIMDBInfoFromLocal(imdbInfo, _proxySettings...)
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
