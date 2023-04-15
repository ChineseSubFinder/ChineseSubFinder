package tmdb_api

import (
	"fmt"
	"strconv"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/sirupsen/logrus"
)

type TmdbApi struct {
	l          *logrus.Logger
	apiKey     string
	tmdbClient *tmdb.Client
}

func NewTmdbHelper(l *logrus.Logger, apiKey string, useAlternateBaseURL bool) (*TmdbApi, error) {

	tmdbClient, err := tmdb.Init(apiKey)
	if err != nil {
		err = fmt.Errorf("error initializing tmdb client: %s", err)
		return nil, err
	}
	if useAlternateBaseURL == true {
		tmdbClient.SetAlternateBaseURL()
	}
	t := TmdbApi{
		l:          l,
		apiKey:     apiKey,
		tmdbClient: tmdbClient,
	}
	t.setClientConfig()
	return &t, nil
}

func (t *TmdbApi) Alive() bool {

	options := make(map[string]string)
	options["language"] = "en-US"
	searchMulti, err := t.tmdbClient.GetSearchMulti("Dexter", options)
	if err != nil {
		t.l.Errorln("GetSearchMulti", err)
		return false
	}
	t.l.Infoln("Tmdb Api is Alive", searchMulti.TotalResults)
	return true
}

// GetInfo 获取视频的信息 idType: imdb_id or tmdb_id
func (t *TmdbApi) GetInfo(iD string, idType string, isMovieOrSeries, isQueryEnOrCNInfo bool) (outFindByID *tmdb.FindByID, err error) {

	// 查询的参数
	options := make(map[string]string)
	if isQueryEnOrCNInfo == true {
		options["language"] = "en-US"
	} else {
		options["language"] = "zh-CN"
	}
	if idType == ImdbID {

		options["external_source"] = "imdb_id"
		outFindByID, err = t.tmdbClient.GetFindByID(iD, options)
		if err != nil {
			return nil, fmt.Errorf("error getting tmdb info by id = %s: %s", iD, err)
		}
	} else if idType == TmdbID {

		intVar, err := strconv.Atoi(iD)
		if err != nil {
			return nil, fmt.Errorf("error converting tmdb id = %s to int: %s", iD, err)
		}

		if isMovieOrSeries == true {
			movieDetails, err := t.tmdbClient.GetMovieDetails(intVar, options)
			if err != nil {
				return nil, fmt.Errorf("error getting tmdb movie details by id = %s: %s", iD, err)
			}
			outFindByID = &tmdb.FindByID{
				MovieResults: []struct {
					Adult            bool    `json:"adult"`
					BackdropPath     string  `json:"backdrop_path"`
					GenreIDs         []int64 `json:"genre_ids"`
					ID               int64   `json:"id"`
					OriginalLanguage string  `json:"original_language"`
					OriginalTitle    string  `json:"original_title"`
					Overview         string  `json:"overview"`
					PosterPath       string  `json:"poster_path"`
					ReleaseDate      string  `json:"release_date"`
					Title            string  `json:"title"`
					Video            bool    `json:"video"`
					VoteAverage      float32 `json:"vote_average"`
					VoteCount        int64   `json:"vote_count"`
					Popularity       float32 `json:"popularity"`
				}{
					{
						Adult:            movieDetails.Adult,
						BackdropPath:     movieDetails.BackdropPath,
						ID:               movieDetails.ID,
						OriginalLanguage: movieDetails.OriginalLanguage,
						OriginalTitle:    movieDetails.OriginalTitle,
						Overview:         movieDetails.Overview,
						PosterPath:       movieDetails.PosterPath,
						ReleaseDate:      movieDetails.ReleaseDate,
						Title:            movieDetails.Title,
						Video:            movieDetails.Video,
						VoteAverage:      movieDetails.VoteAverage,
						VoteCount:        movieDetails.VoteCount,
						Popularity:       movieDetails.Popularity,
					},
				},
			}
		} else {
			tvDetails, err := t.tmdbClient.GetTVDetails(intVar, options)
			if err != nil {
				return nil, fmt.Errorf("error getting tmdb tv details by id = %s: %s", iD, err)
			}
			outFindByID = &tmdb.FindByID{
				TvResults: []struct {
					OriginalName     string   `json:"original_name"`
					ID               int64    `json:"id"`
					Name             string   `json:"name"`
					VoteCount        int64    `json:"vote_count"`
					VoteAverage      float32  `json:"vote_average"`
					FirstAirDate     string   `json:"first_air_date"`
					PosterPath       string   `json:"poster_path"`
					GenreIDs         []int64  `json:"genre_ids"`
					OriginalLanguage string   `json:"original_language"`
					BackdropPath     string   `json:"backdrop_path"`
					Overview         string   `json:"overview"`
					OriginCountry    []string `json:"origin_country"`
					Popularity       float32  `json:"popularity"`
				}{
					{
						OriginalName:     tvDetails.OriginalName,
						ID:               tvDetails.ID,
						Name:             tvDetails.Name,
						VoteCount:        tvDetails.VoteCount,
						VoteAverage:      tvDetails.VoteAverage,
						FirstAirDate:     tvDetails.FirstAirDate,
						PosterPath:       tvDetails.PosterPath,
						OriginalLanguage: tvDetails.OriginalLanguage,
						BackdropPath:     tvDetails.BackdropPath,
						Overview:         tvDetails.Overview,
						OriginCountry:    tvDetails.OriginCountry,
						Popularity:       tvDetails.Popularity,
					},
				},
			}
		}

	}

	return outFindByID, nil
}

// ConvertId 目前仅仅支持 TMDB ID 转 IMDB ID
func (t *TmdbApi) ConvertId(iD string, idType string, isMovieOrSeries bool) (convertIdResult *ConvertIdResult, err error) {

	if idType == ImdbID {
		return nil, fmt.Errorf("imdb id type is not supported")
	} else if idType == TmdbID {
		var intVar int
		intVar, err = strconv.Atoi(iD)
		if err != nil {
			return nil, fmt.Errorf("error converting tmdb id = %s to int: %s", iD, err)
		}
		options := make(map[string]string)
		if isMovieOrSeries == true {
			movieExternalIDs, err := t.tmdbClient.GetMovieExternalIDs(intVar, options)
			if err != nil {
				return nil, err
			}
			convertIdResult = &ConvertIdResult{
				ImdbID: movieExternalIDs.IMDbID,
				TmdbID: iD,
			}

			return convertIdResult, nil
		} else {
			tvExternalIDs, err := t.tmdbClient.GetTVExternalIDs(intVar, options)
			if err != nil {
				return nil, err
			}

			convertIdResult = &ConvertIdResult{
				ImdbID: tvExternalIDs.IMDbID,
				TmdbID: iD,
				TvdbID: fmt.Sprintf("%d", tvExternalIDs.TVDBID),
			}

			return convertIdResult, nil
		}
	} else {
		return nil, fmt.Errorf("id type is not supported: " + idType)
	}
}

func (t *TmdbApi) setClientConfig() {
	// 获取 http client 实例
	restyClient, err := pkg.NewHttpClient()
	if err != nil {
		err = fmt.Errorf("error initializing resty client: %s", err)
		return
	}
	t.tmdbClient.SetClientConfig(*restyClient.GetClient())
	t.tmdbClient.SetClientAutoRetry()
}

const (
	ImdbID = "imdb_id"
	TmdbID = "tmdb_id"
)

type ConvertIdResult struct {
	ImdbID string `json:"imdb_id"`
	TmdbID string `json:"tmdb_id"`
	TvdbID string `json:"tvdb_id"`
}

type Req struct {
	ProxySettings       settings.ProxySettings `json:"proxy_settings"  binding:"required"`
	ApiKey              string                 `json:"api_key"`
	UseAlternateBaseURL bool                   `json:"use_alternate_base_url"`
}
