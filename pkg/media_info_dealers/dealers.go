package media_info_dealers

import (
	"fmt"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/subtitle_best_api"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/tmdb_api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Dealers struct {
	Logger          *logrus.Logger
	SubtitleBestApi *subtitle_best_api.SubtitleBestApi
	tmdbHelper      *tmdb_api.TmdbApi
}

func NewDealers(log *logrus.Logger, subtitleBestApi *subtitle_best_api.SubtitleBestApi) *Dealers {
	return &Dealers{Logger: log, SubtitleBestApi: subtitleBestApi}
}

func (d *Dealers) SetTmdbHelperInstance(tmdbHelper *tmdb_api.TmdbApi) {
	d.tmdbHelper = tmdbHelper
}

// ConvertId 目前仅仅支持 TMDB ID 转 IMDB ID， iD：TMDB ID，idType：tmdb
func (d *Dealers) ConvertId(iD string, idType string, isMovieOrSeries bool) (convertIdResult *tmdb_api.ConvertIdResult, err error) {

	if d.tmdbHelper != nil && settings.Get().AdvancedSettings.TmdbApiSettings.Enable == true && settings.Get().AdvancedSettings.TmdbApiSettings.ApiKey != "" {
		// 优先使用用户自己的 tmdb api
		return d.tmdbHelper.ConvertId(iD, idType, isMovieOrSeries)
	} else {
		// 使用默认的公用服务器 tmdb api
		videoType := ""
		if isMovieOrSeries == true {
			videoType = "movie"
		} else {
			videoType = "series"
		}
		idConvertReply, err := d.SubtitleBestApi.ConvertId(iD, idType, videoType)
		if err != nil {
			return nil, err
		}

		return &tmdb_api.ConvertIdResult{
			ImdbID: idConvertReply.IMDBId,
			TmdbID: iD,
			TvdbID: idConvertReply.TVDBId,
		}, nil
	}
}

func (d *Dealers) GetMediaInfo(id, source, videoType string) (*models.MediaInfo, error) {

	if d.tmdbHelper != nil && settings.Get().AdvancedSettings.TmdbApiSettings.Enable == true && settings.Get().AdvancedSettings.TmdbApiSettings.ApiKey != "" {
		// 优先使用用户自己的 tmdb api
		return d.getMediaInfoFromSelfApi(id, source, videoType)
	} else {
		// 使用默认的公用服务器 tmdb api
		return d.getMediaInfoFromSubtitleBestApi(id, source, videoType)
	}
}

// getMediaInfoFromSelfApi 通过用户自己的 tmdb api 查询媒体信息 "source"=imdb|tmdb  "video_type"=movie|series
func (d *Dealers) getMediaInfoFromSelfApi(id, source, videoType string) (*models.MediaInfo, error) {

	imdbId := ""
	var tmdbID int64
	idType := ""
	isMovieOrSeries := false
	if source == "imdb" {
		idType = tmdb_api.ImdbID
		imdbId = id
		if videoType == "movie" {
			isMovieOrSeries = true
		} else if videoType == "series" {
			isMovieOrSeries = false
		} else {
			return nil, errors.New("videoType is not movie or series")
		}
	} else if source == "tmdb" {

		if videoType == "movie" {
			idType = tmdb_api.TmdbID
			isMovieOrSeries = true
		} else if videoType == "series" {
			idType = tmdb_api.TmdbID
			isMovieOrSeries = false
		} else {
			return nil, errors.New("videoType is not movie or series")
		}
	} else {
		return nil, errors.New("source is not support")
	}
	// 先查询英文信息，然后再查询中文信息
	findByIDEn, err := d.tmdbHelper.GetInfo(id, idType, isMovieOrSeries, true)
	if err != nil {
		return nil, fmt.Errorf("error while getting info from TMDB: %v", err)
	}
	findByIDCn, err := d.tmdbHelper.GetInfo(id, idType, isMovieOrSeries, false)
	if err != nil {
		return nil, fmt.Errorf("error while getting info from TMDB: %v", err)
	}

	OriginalTitle := ""
	OriginalLanguage := ""
	TitleEn := ""
	TitleCn := ""
	Year := ""
	if isMovieOrSeries == true {
		// 电影
		if len(findByIDEn.MovieResults) < 1 {
			return nil, errors.New("not found movie info from tmdb")
		}
		tmdbID = findByIDEn.MovieResults[0].ID
		OriginalTitle = findByIDEn.MovieResults[0].OriginalTitle
		OriginalLanguage = findByIDEn.MovieResults[0].OriginalLanguage
		TitleEn = findByIDEn.MovieResults[0].Title
		TitleCn = findByIDCn.MovieResults[0].Title
		Year = findByIDEn.MovieResults[0].ReleaseDate

	} else {
		// 电视剧
		if len(findByIDEn.TvResults) < 1 {
			return nil, errors.New("not found series info from tmdb")
		}
		tmdbID = findByIDEn.TvResults[0].ID
		OriginalTitle = findByIDEn.TvResults[0].OriginalName
		OriginalLanguage = findByIDEn.TvResults[0].OriginalLanguage
		TitleEn = findByIDEn.TvResults[0].Name
		TitleCn = findByIDCn.TvResults[0].Name
		Year = findByIDEn.TvResults[0].FirstAirDate
	}

	mediaInfo := &models.MediaInfo{
		TmdbId:           fmt.Sprintf("%d", tmdbID),
		ImdbId:           imdbId,
		OriginalTitle:    OriginalTitle,
		OriginalLanguage: OriginalLanguage,
		TitleEn:          TitleEn,
		TitleCn:          TitleCn,
		Year:             Year,
	}

	return mediaInfo, nil
}

// getMediaInfoFromSubtitleBestApi 通过 subtitle.best api 查询媒体信息 "source"=imdb|tmdb  "video_type"=movie|series
func (d *Dealers) getMediaInfoFromSubtitleBestApi(id, source, videoType string) (*models.MediaInfo, error) {

	var mediaInfo *models.MediaInfo
	queryCount := 0
	for {
		queryCount++
		mediaInfoReply, err := d.SubtitleBestApi.GetMediaInfo(id, source, videoType)
		if err != nil {
			return nil, err
		}
		if mediaInfoReply.Status == 2 {
			// 说明进入了查询队列，可以等 30s 以上再次查询
			d.Logger.Infoln("query queue, sleep 30s")
			time.Sleep(30 * time.Second)

		} else if mediaInfoReply.Status == 1 {

			imdbId := ""
			if source == "imdb" {
				imdbId = id
			} else if source == "tmdb" {
				imdbId = ""
			}

			// 说明查询成功
			mediaInfo = &models.MediaInfo{
				TmdbId:           mediaInfoReply.TMDBId,
				ImdbId:           imdbId,
				OriginalTitle:    mediaInfoReply.OriginalTitle,
				OriginalLanguage: mediaInfoReply.OriginalLanguage,
				TitleEn:          mediaInfoReply.TitleEN,
				TitleCn:          mediaInfoReply.TitleCN,
				Year:             mediaInfoReply.Year,
			}
		} else {
			// 说明查询失败
			return nil, errors.New("SubtitleBestApi.GetMediaInfo failed, Message: " + mediaInfoReply.Message)
		}

		if queryCount > 9 {
			break
		}
	}

	return mediaInfo, nil
}
