package imdb_helper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StalkR/imdb"
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/sirupsen/logrus"
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

// GetVideoIMDBInfoFromLocal 从本地获取 IMDB 信息，注意，如果需要跳过，那么返回 Error == common.SkipCreateInDB
func GetVideoIMDBInfoFromLocal(log *logrus.Logger, imdbInfo types.VideoIMDBInfo, skipCreate ...bool) (*models.IMDBInfo, error) {

	/*
		这里需要注意一个细节，之前理想情况下是从 Web 获取完整的 IMDB Info 回来，放入本地存储
		获取的时候如果本地有就拿本地的，没有则从 Web 获取，然后写入本地存储
		但是实际的使用中发现，其实只有在判断视频是否是中文的时候才有必要去获取 Web 上完整的 IMDB Info 信息
		如果一开始就需要从 Web 上获取，那么这个过程非常的缓慢，很耽误时间，所以还是切换为两层逻辑
		1. 第一次，优先获取本地的 IMDB Info，写入数据库缓存
		2. 如果需要判断视频是否是中文的时候，再去获取 Web 上完整的 IMDB Info 信息，更新写入本地存储
		3. 因为现在默认是不跳过中文视频扫描的，所以如果开启后，则会再判断的时候访问外网获取，然后写入本地，过程会比较慢
		4. 同时，再发送字幕和 IMDB Info 到服务器的时候，也需要判断是否 IMDB Info 信息是否齐全，否则就需要从外网获取齐全后再上传
	*/

	log.Debugln("GetVideoIMDBInfoFromLocal", 0)

	// 首先从数据库中查找是否存在这个 IMDB 信息，如果不存在再使用 Web 查找，且写入数据库
	var imdbInfos []models.IMDBInfo
	// 把嵌套关联的 has many 的信息都查询出来
	dao.GetDb().
		Preload("VideoSubInfos").
		Limit(1).Where(&models.IMDBInfo{IMDBID: imdbInfo.ImdbId}).Find(&imdbInfos)

	log.Debugln("GetVideoIMDBInfoFromLocal", 1)

	if len(imdbInfos) <= 0 {

		if len(skipCreate) > 0 && skipCreate[0] == true {
			log.Debugln(fmt.Sprintf("skip insert, imdbInfo.ImdbId = %v", imdbInfo.ImdbId))
			return nil, common.SkipCreateInDB
		}

		// 没有找到，新增，存储本地，但是信息肯定是不完整的，需要在判断是否是中文的时候再次去外网获取补全信息
		log.Debugln("GetVideoIMDBInfoFromLocal", 2)
		// 存入数据库
		nowIMDBInfo := models.NewIMDBInfo(imdbInfo.ImdbId, "", 0, "", []string{}, []string{})
		dao.GetDb().Create(nowIMDBInfo)

		log.Debugln("GetVideoIMDBInfoFromLocal", 3)

		return nowIMDBInfo, nil
	} else {

		log.Debugln("GetVideoIMDBInfoFromLocal", 4)
		// 找到
		return &imdbInfos[0], nil
	}
}

// IsChineseVideo 从 imdbID 去查询判断是否是中文视频
func IsChineseVideo(log *logrus.Logger, imdbInfo types.VideoIMDBInfo, _proxySettings ...*settings.ProxySettings) (bool, *models.IMDBInfo, error) {

	const chName0 = "chinese"
	const chName1 = "mandarin"

	log.Debugln("IsChineseVideo", 0)

	localIMDBInfo, err := GetVideoIMDBInfoFromLocal(log, imdbInfo)
	if err != nil {
		return false, nil, err
	}
	if len(localIMDBInfo.Description) <= 0 {
		// 需要去外网获去补全信息，然后更新本地的信息
		log.Debugln("IsChineseVideo", 1)

		t, err := GetVideoInfoFromIMDBWeb(imdbInfo, _proxySettings...)
		if err != nil {
			log.Errorln("IsChineseVideo.getVideoInfoFromIMDBWeb,", imdbInfo.Title, err)
			return false, nil, err
		}

		log.Debugln("IsChineseVideo", 2)
		localIMDBInfo.Year = t.Year
		localIMDBInfo.Name = t.Name
		localIMDBInfo.Year = t.Year
		localIMDBInfo.AKA = t.AKA
		localIMDBInfo.Description = t.Description
		localIMDBInfo.Languages = t.Languages

		log.Debugln("IsChineseVideo", 3)

		dao.GetDb().Save(localIMDBInfo)

		log.Debugln("IsChineseVideo", 4)
	}

	if len(localIMDBInfo.Languages) < 1 {
		return false, localIMDBInfo, nil
	}

	firstLangLowCase := strings.ToLower(localIMDBInfo.Languages[0])

	log.Debugln("IsChineseVideo", 5)

	// 判断第一语言是否是中文
	switch firstLangLowCase {
	case chName0, chName1:
		return true, localIMDBInfo, nil
	default:
		return false, localIMDBInfo, nil
	}
}

// GetIMDBInfo 先从本地拿缓存，如果没有就从 Web 获取
func GetIMDBInfo(log *logrus.Logger, videoFPath string, isMovie bool, _proxySettings ...*settings.ProxySettings) (*models.IMDBInfo, error) {

	var err error
	var imdbInfo4Video types.VideoIMDBInfo
	if isMovie == true {
		imdbInfo4Video, err = decode.GetImdbInfo4Movie(videoFPath)
	} else {
		imdbInfo4Video, err = decode.GetSeriesSeasonImdbInfoFromEpisode(videoFPath)
	}
	if err != nil {
		// 如果找不到当前电影的 IMDB Info 本地文件，那么就跳过
		log.Warningln("getSubListFromFile", videoFPath, err)
		return nil, err
	}
	imdbInfo, err := GetVideoIMDBInfoFromLocal(log, imdbInfo4Video)
	if err != nil {
		log.Warningln("GetVideoIMDBInfoFromLocal", videoFPath, err)
		return nil, err
	}
	if len(imdbInfo.Description) <= 0 {
		// 需要去外网获去补全信息，然后更新本地的信息
		t, err := GetVideoInfoFromIMDBWeb(imdbInfo4Video, _proxySettings...)
		if err != nil {
			log.Errorln("GetVideoInfoFromIMDBWeb,", imdbInfo4Video.Title, err)
			return nil, err
		}
		imdbInfo.Year = t.Year
		imdbInfo.AKA = t.AKA
		imdbInfo.Description = t.Description
		imdbInfo.Languages = t.Languages

		dao.GetDb().Save(imdbInfo)
	}

	return imdbInfo, nil
}
