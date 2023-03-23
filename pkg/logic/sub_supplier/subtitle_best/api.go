package subtitle_best

import (
	"encoding/json"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/go-resty/resty/v2"
	"strconv"
)

type Api struct {
	headerToken string
	apiKey      string
}

func NewApi(headerToken string, apiKey string) *Api {
	return &Api{headerToken: headerToken, apiKey: apiKey}
}

// QueryMovieSubtitle 查询电影的字幕
func (a *Api) QueryMovieSubtitle(client *resty.Client, imdbID string) (*SubtitleResponse, *LimitInfo, error) {
	// 构建请求体
	requestBody := SearchMovieSubtitleRequest{
		ImdbID: imdbID,
		ApiKey: a.apiKey,
	}

	// 发送请求
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+a.headerToken).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestSearchMovieUrl)
	if err != nil {
		return nil, nil, err
	}

	// 解析响应
	var subtitleResponse SubtitleResponse
	err = json.Unmarshal(resp.Body(), &subtitleResponse)
	if err != nil {
		return nil, nil, err
	}

	return &subtitleResponse, NewHeaderInfo(resp), nil
}

// QueryTVEpsSubtitle 查询连续剧 一季 一集的字幕
func (a *Api) QueryTVEpsSubtitle(client *resty.Client, imdbID string, season, episode int) (*SubtitleResponse, *LimitInfo, error) {
	// 构建请求体
	requestBody := SearchTVEpsSubtitleRequest{
		ImdbID:  imdbID,
		ApiKey:  a.apiKey,
		Season:  season,
		Episode: episode,
	}

	// 发送请求
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+a.headerToken).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestSearchTVEpsUrl)
	if err != nil {
		return nil, nil, err
	}

	// 解析响应
	var subtitleResponse SubtitleResponse
	err = json.Unmarshal(resp.Body(), &subtitleResponse)
	if err != nil {
		return nil, nil, err
	}

	return &subtitleResponse, NewHeaderInfo(resp), nil
}

// QueryTVSeasonPackages 查询连续剧 一季的字幕包 ID 列表
func (a *Api) QueryTVSeasonPackages(client *resty.Client, imdbID string, season int) (*SeasonPackagesResponse, *LimitInfo, error) {
	// 构建请求体
	requestBody := SearchTVSeasonPackagesRequest{
		ImdbID: imdbID,
		ApiKey: a.apiKey,
		Season: season,
	}

	// 发送请求
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+a.headerToken).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestSearchTVSeasonPackageUrl)
	if err != nil {
		return nil, nil, err
	}

	// 解析响应
	var seasonPackageResponse SeasonPackagesResponse
	err = json.Unmarshal(resp.Body(), &seasonPackageResponse)
	if err != nil {
		return nil, nil, err
	}

	return &seasonPackageResponse, NewHeaderInfo(resp), nil
}

// QueryTVSeasonPackageByID 查询连续剧 一季 一字幕包的字幕
func (a *Api) QueryTVSeasonPackageByID(client *resty.Client, imdbID string, seasonPackageId string) (*SubtitleResponse, *LimitInfo, error) {
	// 构建请求体
	requestBody := SearchTVSeasonPackageByIDRequest{
		ImdbID:          imdbID,
		ApiKey:          a.apiKey,
		SeasonPackageId: seasonPackageId,
	}

	// 发送请求
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+a.headerToken).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestSearchTVSeasonPackageByIDUrl)
	if err != nil {
		return nil, nil, err
	}

	// 解析响应
	var subtitleResponse SubtitleResponse
	err = json.Unmarshal(resp.Body(), &subtitleResponse)
	if err != nil {
		return nil, nil, err
	}

	return &subtitleResponse, NewHeaderInfo(resp), nil
}

// GetDownloadUrl 获取字幕下载地址
func (a *Api) GetDownloadUrl(client *resty.Client, subSha256, imdbID string,
	isMovie bool, season, episode int,
	seasonPackageId string, language int,
	token string) (*GetUrlResponse, *LimitInfo, error) {
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
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+a.headerToken).
		SetBody(requestBody).
		Post(common.SubSubtitleBestRootUrlDef + common.SubSubtitleBestGetDlURLUrl)
	if err != nil {
		return nil, nil, err
	}

	// 解析响应
	var getUrlResponse GetUrlResponse
	err = json.Unmarshal(resp.Body(), &getUrlResponse)
	if err != nil {
		return nil, nil, err
	}

	return &getUrlResponse, NewHeaderInfo(resp), nil
}

type LimitInfo struct {
	dailyLimit         string
	dailyCount         string
	rateLimitLimit     string
	rateLimitRemaining string
	rateLimitReset     string
}

func NewHeaderInfo(resp *resty.Response) *LimitInfo {
	return &LimitInfo{
		dailyLimit:         resp.Header().Get("X-Daily-Limit"),
		dailyCount:         resp.Header().Get("X-Daily-Count"),
		rateLimitLimit:     resp.Header().Get("X-RateLimit-Limit"),
		rateLimitRemaining: resp.Header().Get("X-RateLimit-Remaining"),
		rateLimitReset:     resp.Header().Get("X-RateLimit-Reset"),
	}
}

func (h LimitInfo) DailyLimit() int {
	dailyLimit, _ := strconv.Atoi(h.dailyLimit)
	return dailyLimit
}

func (h LimitInfo) DailyCount() int {
	dailyCount, _ := strconv.Atoi(h.dailyCount)
	return dailyCount
}

func (h LimitInfo) RateLimitLimit() int {
	rateLimitLimit, _ := strconv.Atoi(h.rateLimitLimit)
	return rateLimitLimit
}

func (h LimitInfo) RateLimitRemaining() int {
	rateLimitRemaining, _ := strconv.Atoi(h.rateLimitRemaining)
	return rateLimitRemaining
}

func (h LimitInfo) RateLimitReset() int {
	rateLimitReset, _ := strconv.Atoi(h.rateLimitReset)
	return rateLimitReset
}
