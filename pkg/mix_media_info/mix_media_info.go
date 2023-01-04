package mix_media_info

import (
	"errors"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/media_info_dealers"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"gorm.io/gorm"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/imdb_helper"
)

func GetMixMediaInfo(
	dealers *media_info_dealers.Dealers,
	videoFPath string, isMovie bool) (*models.MediaInfo, error) {
	// 从本地读取 IMDB ID 信息，找到基本 ID 信息后，也会去 IMDB web 找到对应的额外信息填充
	imdbInfo, err := imdb_helper.GetIMDBInfoFromVideoFile(dealers, videoFPath, isMovie)
	if err != nil {
		return nil, err
	}

	source := "imdb"
	videoType := "movie"
	if isMovie == false {
		videoType = "series"
	}

	// TMDB ID 是否存在
	if imdbInfo.TmdbId == "" {
		// 需要去 web 查询
		source = "imdb"
		return GetMediaInfoAndSave(dealers, imdbInfo, imdbInfo.IMDBID, source, videoType)
	} else {
		// 已经存在，从本地拿去信息
		// 首先从数据库中查找是否存在这个 IMDB 信息，如果不存在再使用 Web 查找，且写入数据库
		var mediaInfos []models.MediaInfo
		// 把嵌套关联的 has many 的信息都查询出来
		dao.GetDb().Limit(1).Where(&models.MediaInfo{TmdbId: imdbInfo.TmdbId}).Find(&mediaInfos)

		if len(mediaInfos) > 0 {
			// 找到
			return &mediaInfos[0], nil
		} else {
			// 没有找到本地缓存的 TMDB ID 信息，需要去 web 查询
			source = "imdb"
			return GetMediaInfoAndSave(dealers, imdbInfo, imdbInfo.IMDBID, source, videoType)
		}
	}
}

// GetMediaInfoAndSave 通过 IMDB ID 查询媒体信息，并保存到数据库，IMDB 和 MediaInfo 都会进行保存 // source，options=imdb|tmdb  videoType，options=movie|series
func GetMediaInfoAndSave(dealers *media_info_dealers.Dealers, imdbInfo *models.IMDBInfo, id, source, videoType string) (*models.MediaInfo, error) {

	mediaInfo, err := dealers.GetMediaInfo(id, source, videoType)
	if err != nil {
		return nil, err
	}
	if mediaInfo == nil {
		// 超过 9次 30s 等待都没有查询到，返回错误
		return nil, errors.New("can't get media info from subtitle.best api")
	}
	// 更新 ID
	imdbInfo.TmdbId = mediaInfo.TmdbId
	err = dao.GetDb().Transaction(func(tx *gorm.DB) error {

		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		if err := tx.Save(imdbInfo).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		if err := tx.Save(mediaInfo).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		// 返回 nil 提交事务
		return nil
	})
	if err != nil {
		return nil, err
	}

	return mediaInfo, nil
}

// KeyWordSelect keyWordType cn, 中文， en，英文，org，原始名称
func KeyWordSelect(mediaInfo *models.MediaInfo, videoFPath string, isMovie bool, keyWordType string) (string, error) {

	keyWord := ""

	if keyWordType == "cn" {
		keyWord = mediaInfo.TitleCn
		if keyWord == "" {
			return "", errors.New("TitleCn is empty")
		}
	} else if keyWordType == "en" {
		keyWord = mediaInfo.TitleEn
		if keyWord == "" {
			return "", errors.New("TitleEn is empty")
		}
	} else if keyWordType == "org" {
		keyWord = mediaInfo.OriginalTitle
		if keyWord == "" {
			return "", errors.New("OriginalTitle is empty")
		}
	} else {
		return "", errors.New("keyWordType is not cn, en, org")
	}

	if isMovie == false {
		// 连续剧需要额外补充 S01E01 这样的信息
		epsVideoNfoInfo, err := decode.GetVideoNfoInfo4OneSeriesEpisode(videoFPath)
		if err != nil {
			return "", err
		}
		keyWord += " " + pkg.GetEpisodeKeyName(epsVideoNfoInfo.Season, epsVideoNfoInfo.Episode, true)
	}

	return keyWord, nil
}
