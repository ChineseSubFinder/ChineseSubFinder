package imdb_helper

import (
	"errors"
	"github.com/jinzhu/now"
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/media_info_dealers"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types"
)

// GetIMDBInfoFromVideoFile 先从本地拿缓存，如果没有就从 Web 获取
func GetIMDBInfoFromVideoFile(dealers *media_info_dealers.Dealers, videoFPath string, isMovie bool) (*models.IMDBInfo, error) {

	var err error
	var videoNfoInfo types.VideoNfoInfo
	if isMovie == true {
		videoNfoInfo, err = decode.GetVideoNfoInfo4Movie(videoFPath)
	} else {
		videoNfoInfo, err = decode.GetVideoNfoInfoFromEpisode(videoFPath)
	}
	if err != nil {
		// 如果找不到当前电影的 IMDB Info 本地文件，那么就跳过
		dealers.Logger.Warningln("getSubListFromFile", videoFPath, err)
		return nil, err
	}
	// 这里一定会保存本地的 IMDB 信息
	imdbInfo, err := GetIMDBInfoFromVideoNfoInfo(dealers, videoNfoInfo)
	if err != nil {
		dealers.Logger.Warningln("GetIMDBInfoFromVideoNfoInfo", videoFPath, err)
		return nil, err
	}
	if len(imdbInfo.Description) <= 0 {
		// 需要去外网获去补全信息，然后更新本地的信息
		if imdbInfo.IMDBID != "" && videoNfoInfo.ImdbId == "" {
			// 可能本地没有获取到 IMDB ID 信息，那么从上面的 GetIMDBInfoFromVideoNfoInfo 可以从 TMDB ID 获取到 IMDB ID，那么需要传递下去
			videoNfoInfo.ImdbId = imdbInfo.IMDBID
		}
		videoType := ""
		if isMovie == true {
			videoType = "movie"
		} else {
			videoType = "series"
		}
		mediaInfo, err := dealers.GetMediaInfo(videoNfoInfo.ImdbId, "imdb", videoType)
		if err != nil {
			return nil, err
		}
		parseTime, err := now.Parse(mediaInfo.Year)
		if err != nil {
			return nil, err
		}
		imdbInfo.Year = parseTime.Year()
		imdbInfo.AKA = []string{mediaInfo.TitleCn, mediaInfo.TitleEn, mediaInfo.OriginalTitle}
		imdbInfo.Description = mediaInfo.OriginalTitle
		imdbInfo.Languages = []string{mediaInfo.OriginalLanguage}

		dao.GetDb().Save(imdbInfo)
	}

	return imdbInfo, nil
}

// GetIMDBInfoFromVideoNfoInfo 从本地获取 IMDB 信息，注意，如果需要跳过，那么返回 Error == common.SkipCreateInDB
func GetIMDBInfoFromVideoNfoInfo(dealers *media_info_dealers.Dealers, videoNfoInfo types.VideoNfoInfo) (*models.IMDBInfo, error) {

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
	dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", "IMDBID:", videoNfoInfo.ImdbId, "TMDBID:", videoNfoInfo.TmdbId, videoNfoInfo.Title, videoNfoInfo.Season, videoNfoInfo.Episode)
	dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 0)

	if videoNfoInfo.ImdbId != "" {
		// 优先从 IMDB ID 去查找本地的信息
		// 首先从数据库中查找是否存在这个 IMDB 信息，如果不存在再使用 Web 查找，且写入数据库
		var imdbInfos []models.IMDBInfo
		// 把嵌套关联的 has many 的信息都查询出来
		dao.GetDb().
			Preload("VideoSubInfos").
			Limit(1).Where(&models.IMDBInfo{IMDBID: videoNfoInfo.ImdbId}).Find(&imdbInfos)

		dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 1)

		if len(imdbInfos) <= 0 {
			// 没有找到，新增，存储本地，但是信息肯定是不完整的，需要在判断是否是中文的时候再次去外网获取补全信息
			dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 2)
			// 存入数据库
			nowIMDBInfo := models.NewIMDBInfo(videoNfoInfo.ImdbId, "", 0, "", []string{}, []string{})
			dao.GetDb().Create(nowIMDBInfo)

			dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 3)

			return nowIMDBInfo, nil
		} else {

			dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 4)
			// 找到
			return &imdbInfos[0], nil
		}
	} else if videoNfoInfo.TmdbId != "" {
		// 如果 IMDB ID 在本地没有获取到，但是 TMDB ID 获取到了，那么就从 Web 去查询 IMDB ID 出来
		var imdbInfos []models.IMDBInfo
		// 把嵌套关联的 has many 的信息都查询出来
		dao.GetDb().
			Preload("VideoSubInfos").
			Limit(1).Where(&models.IMDBInfo{TmdbId: videoNfoInfo.TmdbId}).Find(&imdbInfos)

		dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 1)

		if len(imdbInfos) <= 0 {
			// 没有找到那么就从 Web 端获取 imdb id 信息
		} else if len(imdbInfos) > 1 {
			// 如果找到多个，那么就应该删除这些，因为这些都是重复的，然后再次从 Web 去获取 imdb id 信息
			dao.GetDb().Where(&models.IMDBInfo{TmdbId: videoNfoInfo.TmdbId}).Delete(&models.IMDBInfo{})
		} else {
			dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 4)
			// 找到
			return &imdbInfos[0], nil
		}
		// 确定需要从 Web 端获取 imdb id 信息
		dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 2)
		// 联网查询
		idConvertReply, err := dealers.ConvertId(videoNfoInfo.TmdbId, "tmdb_id", videoNfoInfo.IsMovie)
		if err != nil {
			return nil, err
		}
		dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 3)
		// 存入数据库
		nowIMDBInfo := models.NewIMDBInfo(idConvertReply.ImdbID, "", 0, "", []string{}, []string{})
		dao.GetDb().Create(nowIMDBInfo)
		dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo", 4)
		return nowIMDBInfo, nil

	} else {
		// 都没有，那么就报错
		dealers.Logger.Debugln("GetIMDBInfoFromVideoNfoInfo IMDB TMDB ID is empty")
		return nil, errors.New("IMDB TMDB ID is empty")
	}
}

// IsChineseVideo 从 imdbID 去查询判断是否是中文视频
func IsChineseVideo(dealers *media_info_dealers.Dealers, videoNfoInfo types.VideoNfoInfo) (bool, *models.IMDBInfo, error) {

	const chName0 = "chinese"
	const chName1 = "mandarin"
	const chName2 = "zh"

	dealers.Logger.Debugln("IsChineseVideo", 0)

	localIMDBInfo, err := GetIMDBInfoFromVideoNfoInfo(dealers, videoNfoInfo)
	if err != nil {
		return false, nil, err
	}
	if len(localIMDBInfo.Description) <= 0 {
		// 需要去外网获去补全信息，然后更新本地的信息
		dealers.Logger.Debugln("IsChineseVideo", 1)

		videoType := ""
		if videoNfoInfo.IsMovie == true {
			videoType = "movie"
		} else {
			videoType = "series"
		}
		mediaInfo, err := dealers.GetMediaInfo(videoNfoInfo.ImdbId, "imdb", videoType)
		if err != nil {
			return false, nil, err
		}
		parseTime, err := now.Parse(mediaInfo.Year)
		if err != nil {
			return false, nil, err
		}

		dealers.Logger.Debugln("IsChineseVideo", 2)

		name := mediaInfo.TitleCn
		if name == "" {
			name = mediaInfo.OriginalTitle
		}

		localIMDBInfo.Year = parseTime.Year()
		localIMDBInfo.Name = name
		localIMDBInfo.AKA = []string{mediaInfo.TitleCn, mediaInfo.TitleEn, mediaInfo.OriginalTitle}
		localIMDBInfo.Description = mediaInfo.OriginalTitle
		localIMDBInfo.Languages = []string{mediaInfo.OriginalLanguage}

		dealers.Logger.Debugln("IsChineseVideo", 3)

		dao.GetDb().Save(localIMDBInfo)

		dealers.Logger.Debugln("IsChineseVideo", 4)
	}

	if len(localIMDBInfo.Languages) < 1 {
		return false, localIMDBInfo, nil
	}

	firstLangLowCase := strings.ToLower(localIMDBInfo.Languages[0])

	dealers.Logger.Debugln("IsChineseVideo", 5)

	// 判断第一语言是否是中文
	switch firstLangLowCase {
	case chName0, chName1, chName2:
		return true, localIMDBInfo, nil
	default:
		return false, localIMDBInfo, nil
	}
}

//// getVideoInfoFromIMDBWeb 从 IMDB 网站 ID 查询影片的信息
//func getVideoInfoFromIMDBWeb(videoNfoInfo types.VideoNfoInfo) (*imdb.Title, error) {
//
//	client, err := pkg.NewHttpClient()
//	if err != nil {
//		return nil, err
//	}
//
//	t, err := imdb.NewTitle(client.GetClient(), videoNfoInfo.ImdbId)
//	if err != nil {
//		notify_center.Notify.Add("imdb model - imdb.NewTitle :", err.Error())
//		return nil, err
//	}
//	if t.Year == 0 {
//		// IMDB 信息获取的库(1.0.7)，目前有bug，比如，tt6856242 年份为 0
//		if videoNfoInfo.Year != "" {
//			year, err := strconv.Atoi(videoNfoInfo.Year)
//			if err != nil {
//				return nil, err
//			}
//			t.Year = year
//		}
//	}
//
//	return t, nil
//}
