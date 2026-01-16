package media_info_dealers

import (
	"fmt"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/tmdb_api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Dealers struct {
	Logger     *logrus.Logger
	tmdbHelper *tmdb_api.TmdbApi
}

func NewDealers(log *logrus.Logger) *Dealers {
	return &Dealers{Logger: log}
}

func (d *Dealers) SetTmdbHelperInstance(tmdbHelper *tmdb_api.TmdbApi) {
	d.tmdbHelper = tmdbHelper
}

// ConvertId 目前仅仅支持 TMDB ID 转 IMDB ID， iD：TMDB ID，idType：tmdb
func (d *Dealers) ConvertId(iD string, idType string, isMovieOrSeries bool) (convertIdResult *tmdb_api.ConvertIdResult, err error) {

	if d.tmdbHelper != nil && settings.Get().AdvancedSettings.TmdbApiSettings.Enable == true && settings.Get().AdvancedSettings.TmdbApiSettings.ApiKey != "" {
		// 优先使用用户自己的 tmdb api
		return d.tmdbHelper.ConvertId(iD, idType, isMovieOrSeries)
	}

	return nil, errors.New("tmdb api is not configured")
}

func (d *Dealers) GetMediaInfo(id, source, videoType string) (*models.MediaInfo, error) {

	if d.tmdbHelper != nil && settings.Get().AdvancedSettings.TmdbApiSettings.Enable == true && settings.Get().AdvancedSettings.TmdbApiSettings.ApiKey != "" {
		// 优先使用用户自己的 tmdb api
		return d.getMediaInfoFromSelfApi(id, source, videoType)
	}

	return nil, errors.New("tmdb api is not configured")
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
