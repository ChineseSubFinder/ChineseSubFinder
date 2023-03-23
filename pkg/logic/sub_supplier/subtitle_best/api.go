package subtitle_best

import (
	"encoding/json"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/go-resty/resty/v2"
)

type Api struct {
	client *resty.Client
	token  string
	apiKey string
}

// QueryMovieSubtitle 查询电影的字幕
func (a *Api) QueryMovieSubtitle(imdbID string) (*SubtitleResponse, error) {
	// 构建请求体
	requestBody := SearchMovieSubtitleRequest{
		ImdbID: imdbID,
		ApiKey: a.apiKey,
	}

	// 发送请求
	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+a.token).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestSearchMovieUrl)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var subtitleResponse SubtitleResponse
	err = json.Unmarshal(resp.Body(), &subtitleResponse)
	if err != nil {
		return nil, err
	}

	return &subtitleResponse, nil
}

// QueryTVEpsSubtitle 查询连续剧 一季 一集的字幕
func (a *Api) QueryTVEpsSubtitle(imdbID string, season, episode int) (*SubtitleResponse, error) {
	// 构建请求体
	requestBody := SearchTVEpsSubtitleRequest{
		ImdbID:  imdbID,
		ApiKey:  a.apiKey,
		Season:  season,
		Episode: episode,
	}

	// 发送请求
	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+a.token).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestSearchTVEpsUrl)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var subtitleResponse SubtitleResponse
	err = json.Unmarshal(resp.Body(), &subtitleResponse)
	if err != nil {
		return nil, err
	}

	return &subtitleResponse, nil
}

// QueryTVSeasonPackages 查询连续剧 一季的字幕包 ID 列表
func (a *Api) QueryTVSeasonPackages(imdbID string, season int) (*SeasonPackagesResponse, error) {
	// 构建请求体
	requestBody := SearchTVSeasonPackagesRequest{
		ImdbID: imdbID,
		ApiKey: a.apiKey,
		Season: season,
	}

	// 发送请求
	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+a.token).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestSearchTVSeasonPackageUrl)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var seasonPackageResponse SeasonPackagesResponse
	err = json.Unmarshal(resp.Body(), &seasonPackageResponse)
	if err != nil {
		return nil, err
	}

	return &seasonPackageResponse, nil
}

// QueryTVSeasonPackageByID 查询连续剧 一季 一字幕包的字幕
func (a *Api) QueryTVSeasonPackageByID(imdbID string, seasonPackageId string) (*SubtitleResponse, error) {
	// 构建请求体
	requestBody := SearchTVSeasonPackageByIDRequest{
		ImdbID:          imdbID,
		ApiKey:          a.apiKey,
		SeasonPackageId: seasonPackageId,
	}

	// 发送请求
	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+a.token).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestSearchTVSeasonPackageByIDUrl)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var subtitleResponse SubtitleResponse
	err = json.Unmarshal(resp.Body(), &subtitleResponse)
	if err != nil {
		return nil, err
	}

	return &subtitleResponse, nil
}

// GetDownloadUrl 获取字幕下载地址
func (a *Api) GetDownloadUrl(subSha256, imdbID string,
	isMovie bool, season, episode int,
	seasonPackageId string, language int,
	token string) (*GetUrlResponse, error) {
	// 构建请求体
	requestBody := DownloadUrlConvertRequest{
		SubSha256:       subSha256,
		ImdbID:          imdbID,
		IsMovie:         isMovie,
		Season:          season,
		Episode:         episode,
		SeasonPackageId: seasonPackageId,
		Language:        language,
		ApiKey:          a.apiKey,
		Token:           token,
	}

	// 发送请求
	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+a.token).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestGetDlURLUrl)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var getUrlResponse GetUrlResponse
	err = json.Unmarshal(resp.Body(), &getUrlResponse)
	if err != nil {
		return nil, err
	}

	return &getUrlResponse, nil
}
