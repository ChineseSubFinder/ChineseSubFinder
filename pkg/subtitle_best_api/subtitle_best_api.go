package subtitle_best_api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/allanpk716/ChineseSubFinder/pkg/global_value"

	"github.com/allanpk716/ChineseSubFinder/pkg/common"

	"github.com/sirupsen/logrus"

	"github.com/allanpk716/ChineseSubFinder/internal/models"

	"github.com/allanpk716/ChineseSubFinder/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/pkg/random_auth_key"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
)

type SubtitleBestApi struct {
	log           *logrus.Logger
	authKey       random_auth_key.AuthKey
	randomAuthKey *random_auth_key.RandomAuthKey
	proxySettings *settings.ProxySettings
}

func NewSubtitleBestApi(log *logrus.Logger, inAuthKey random_auth_key.AuthKey, proxySettings *settings.ProxySettings) *SubtitleBestApi {
	return &SubtitleBestApi{
		log:           log,
		randomAuthKey: random_auth_key.NewRandomAuthKey(5, inAuthKey),
		authKey:       inAuthKey,
		proxySettings: proxySettings,
	}
}

func (s *SubtitleBestApi) CheckAlive(proxySettings ...*settings.ProxySettings) error {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return errors.New("auth key is not set")
	}

	postUrl := webUrlBase + "/v1/subhd-code"
	httpClient, err := my_util.NewHttpClient(proxySettings...)
	if err != nil {
		return err
	}
	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return err
	}
	var codeReplyData CodeReplyData
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetQueryParams(map[string]string{
			"now_time": time.Now().Format("2006-01-02"),
		}).
		SetHeader("Accept", "application/json").
		SetResult(&codeReplyData).
		Get(postUrl)
	if err != nil {
		s.log.Errorln("get code error, status code:", resp.StatusCode(), "Error:", err)
		return err
	}

	return nil
}

func (s *SubtitleBestApi) GetCode() (string, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return "", errors.New("auth key is not set")
	}

	postUrl := webUrlBase + "/v1/subhd-code"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return "", err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return "", err
	}

	var codeReplyData CodeReplyData
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetQueryParams(map[string]string{
			"now_time": time.Now().Format("2006-01-02"),
		}).
		SetHeader("Accept", "application/json").
		SetResult(&codeReplyData).
		Get(postUrl)
	if err != nil {
		s.log.Errorln("get code error, status code:", resp.StatusCode(), "Error:", err)
		return "", err
	}

	if codeReplyData.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	if codeReplyData.Status == 0 {
		return "", errors.New(codeReplyData.Message)
	}

	decodeBytes, err := base64.StdEncoding.DecodeString(codeReplyData.Code)
	if err != nil {
		return "", err
	}

	return string(decodeBytes), nil
}

func (s *SubtitleBestApi) GetMediaInfo(id, source, videoType string) (*MediaInfoReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	if len(s.authKey.AESKey16) != 16 {
		return nil, errors.New(fmt.Sprintf("AESKey16 is not set, %s", s.authKey.AESKey16))
	}
	if len(s.authKey.AESIv16) != 16 {
		return nil, errors.New(fmt.Sprintf("AESIv16 is not set, %s", s.authKey.AESIv16))
	}

	postUrl := webUrlBase + "/v1/media-info"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	var mediaInfoReply MediaInfoReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetBody(MediaInfoReq{
			Id:        id,
			Source:    source,
			VideoType: videoType,
		}).
		SetResult(&mediaInfoReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("get media info error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if mediaInfoReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	return &mediaInfoReply, nil
}

// AskFroUpload 在使用这个接口前，需要从 IMDB ID 获取到 TMDB ID
func (s *SubtitleBestApi) AskFroUpload(subSha256 string, IsMovie, trusted bool, ImdbId, TmdbId string, Season, Episode int, VideoFeature string) (*AskForUploadReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	postUrl := webUrlBase + "/v1/ask-for-upload"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	isMovieStr := "false"
	if IsMovie == true {
		isMovieStr = "true"
	}

	trustedStr := "false"
	if trusted == true {
		trustedStr = "true"
	}

	formData := make(map[string]string)
	if trusted == true {
		formData["sub_sha256"] = subSha256
		formData["is_movie"] = isMovieStr
		formData["trusted"] = trustedStr
		formData["imdb_id"] = ImdbId
		formData["tmdb_id"] = TmdbId
		formData["season"] = strconv.Itoa(Season)
		formData["episode"] = strconv.Itoa(Episode)
		formData["video_feature"] = VideoFeature
	} else {
		formData["sub_sha256"] = subSha256
		formData["is_movie"] = isMovieStr
	}

	var askForUploadReply AskForUploadReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFormData(formData).
		SetResult(&askForUploadReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("ask for upload error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if askForUploadReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	return &askForUploadReply, nil
}

// UploadSub 在使用这个接口前，需要从 IMDB ID 获取到 TMDB ID，其实在这一步应该默认就拿到了 TMDB ID，需要提前在 AskFroUpload 接口调用前就搞定这个
// year 这个也是从之前的接口拿到, 2019  or  2022
func (s *SubtitleBestApi) UploadSub(videoSubInfo *models.VideoSubInfo, subSaveRootDirPath string, tmdbId, year string) (*UploadSubReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}

	postUrl := webUrlBase + "/v1/upload-sub"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	// 从相对路径转换为绝对路径
	subFileFPath := filepath.Join(subSaveRootDirPath, videoSubInfo.StoreRPath)
	if my_util.IsFile(subFileFPath) == false {
		return nil, errors.New(fmt.Sprintf("sub file not exist, %s", subFileFPath))
	}
	file, err := os.Open(subFileFPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("open sub file failed, %s", subFileFPath))
	}
	defer func() {
		_ = file.Close()
	}()
	fd, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("read sub file failed, %s", subFileFPath))
	}

	isDouble := "false"
	if videoSubInfo.IsDouble == true {
		isDouble = "true"
	}

	isMovieStr := "false"
	if videoSubInfo.IsMovie == true {
		isMovieStr = "true"
	}

	var uploadSubReply UploadSubReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFileReader("sub_file_context", videoSubInfo.SubName, bytes.NewReader(fd)).
		SetFormData(map[string]string{
			"sub_sha256":     videoSubInfo.SHA256,
			"is_movie":       isMovieStr,
			"season":         strconv.Itoa(videoSubInfo.Season),
			"episode":        strconv.Itoa(videoSubInfo.Episode),
			"is_double":      isDouble,
			"language_iso":   videoSubInfo.LanguageISO,
			"my_language":    videoSubInfo.MyLanguage,
			"extra_pre_name": videoSubInfo.ExtraPreName,
			"imdb_id":        videoSubInfo.IMDBInfoID,
			"tmdb_id":        tmdbId,
			"video_feature":  videoSubInfo.Feature,
			"year":           year,
		}).
		SetResult(&uploadSubReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("upload sub error, status code:", resp.StatusCode(), "Error:", err)
		if resp != nil && resp.StatusCode() == 413 {
			// 文件上传大小超限
			return nil, common.ErrorUpload413
		}

		return nil, err
	}

	if uploadSubReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}
	if resp.StatusCode() == 413 {
		// 文件上传大小超限
		return nil, common.ErrorUpload413
	}

	return &uploadSubReply, nil
}

func (s *SubtitleBestApi) UploadLowTrustSub(lowTrustVideoSubInfo *models.LowVideoSubInfo, subSaveRootDirPath string, tmdbId, year, taskID string) (*UploadSubReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}

	postUrl := webUrlBase + "/v1/upload-sub"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	// 从相对路径转换为绝对路径
	subFileFPath := filepath.Join(subSaveRootDirPath, lowTrustVideoSubInfo.StoreRPath)
	if my_util.IsFile(subFileFPath) == false {
		return nil, errors.New(fmt.Sprintf("sub file not exist, %s", subFileFPath))
	}
	file, err := os.Open(subFileFPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("open sub file failed, %s", subFileFPath))
	}
	defer func() {
		_ = file.Close()
	}()
	fd, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("read sub file failed, %s", subFileFPath))
	}

	isDouble := "false"
	if lowTrustVideoSubInfo.IsDouble == true {
		isDouble = "true"
	}

	isMovieStr := "false"
	if lowTrustVideoSubInfo.IsMovie == true {
		isMovieStr = "true"
	}

	var uploadSubReply UploadSubReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFileReader("sub_file_context", lowTrustVideoSubInfo.SubName, bytes.NewReader(fd)).
		SetFormData(map[string]string{
			"sub_sha256":     lowTrustVideoSubInfo.SHA256,
			"is_movie":       isMovieStr,
			"season":         strconv.Itoa(lowTrustVideoSubInfo.Season),
			"episode":        strconv.Itoa(lowTrustVideoSubInfo.Episode),
			"is_double":      isDouble,
			"language_iso":   lowTrustVideoSubInfo.LanguageISO,
			"my_language":    lowTrustVideoSubInfo.MyLanguage,
			"extra_pre_name": lowTrustVideoSubInfo.ExtraPreName,
			"imdb_id":        lowTrustVideoSubInfo.IMDBID,
			"tmdb_id":        tmdbId,
			"video_feature":  lowTrustVideoSubInfo.Feature,
			"year":           year,
			"low_trust":      "true",
			"task_id":        taskID,
		}).
		SetResult(&uploadSubReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("upload sub error, status code:", resp.StatusCode(), "Error:", err)

		if resp != nil && resp.StatusCode() == 413 {
			// 文件上传大小超限
			return nil, common.ErrorUpload413
		}
		return nil, err
	}

	if uploadSubReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	if resp.StatusCode() == 413 {
		// 文件上传大小超限
		return nil, common.ErrorUpload413
	}

	return &uploadSubReply, nil
}

func (s *SubtitleBestApi) AskFindSub(VideoFeature, ImdbId, TmdbId, Season, Episode, FindSubToken, ApiKey string) (*AskFindSubReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	postUrl := webUrlBase + "/v1/ask-find-sub"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	postData := map[string]string{
		"video_feature":  VideoFeature,
		"imdb_id":        ImdbId,
		"tmdb_id":        TmdbId,
		"season":         Season,
		"episode":        Episode,
		"find_sub_token": FindSubToken,
	}
	if ApiKey != "" {
		postData["api_key"] = ApiKey
	}
	var askFindSubReply AskFindSubReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFormData(postData).
		SetResult(&askFindSubReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("ask find sub error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if askFindSubReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	return &askFindSubReply, nil
}

func (s *SubtitleBestApi) FindSub(VideoFeature, ImdbId, TmdbId, Season, Episode, FindSubToken, ApiKey string) (*FindSubReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	postUrl := webUrlBase + "/v1/find-sub"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	postData := map[string]string{
		"video_feature":  VideoFeature,
		"imdb_id":        ImdbId,
		"tmdb_id":        TmdbId,
		"season":         Season,
		"episode":        Episode,
		"find_sub_token": FindSubToken,
	}
	if ApiKey != "" {
		postData["api_key"] = ApiKey
	}
	var findSubReply FindSubReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFormData(postData).
		SetResult(&findSubReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("find sub error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if findSubReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	return &findSubReply, nil
}

func (s *SubtitleBestApi) AskDownloadSub(SubSha256, DownloadToken, ApiKey string) (*AskForDownloadReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	postUrl := webUrlBase + "/v1/ask-for-download"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	postData := map[string]string{
		"sub_sha256":     SubSha256,
		"download_token": DownloadToken,
	}
	if ApiKey != "" {
		postData["api_key"] = ApiKey
	}
	var askDownloadReply AskForDownloadReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFormData(postData).
		SetResult(&askDownloadReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("ask download sub error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if askDownloadReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	return &askDownloadReply, nil
}

// DownloadSub 首先要确认 downloadFileDesFPath 这个文件是否存在，如果存在且跟需要下载的文件的 sha256 一样就要跳过，然后下载完毕后，也需要 check 这个文件是否存在，存在则需要判断是否是字幕
func (s *SubtitleBestApi) DownloadSub(SubSha256, DownloadToken, ApiKey, downloadFileDesFPath string) (*DownloadSubReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	postUrl := webUrlBase + "/v1/download-sub"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	postData := map[string]string{
		"sub_sha256":     SubSha256,
		"download_token": DownloadToken,
	}
	if ApiKey != "" {
		postData["api_key"] = ApiKey
	}

	if my_util.IsFile(downloadFileDesFPath) == true {
		err = os.Remove(downloadFileDesFPath)
		if err != nil {
			return nil, errors.New("remove file error: " + err.Error())
		}
	}

	var downloadReply DownloadSubReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFormData(postData).
		SetOutput(downloadFileDesFPath).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("download sub error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if downloadReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	readFile, err := ioutil.ReadFile(downloadFileDesFPath)
	if err != nil {
		return nil, errors.New("read file error: " + err.Error())
	}
	err = json.Unmarshal(readFile, &downloadReply)
	if err != nil {
		// 说明成功了，但是出去，还是要再次判断这个是不是字幕文件才行
		downloadReply.Status = 1
		downloadReply.Message = "success"
		return &downloadReply, nil
	}
	// 正常来说，只会获取到字幕，不会有这个 DownloadSubReply 结构的返回，上面获取到了字幕文件，也是伪造一个返回而已
	// 说明返回的这个文件是正常的 reply 文件，那么需要把下载的文件给删除了
	err = os.Remove(downloadFileDesFPath)
	if err != nil {
		return nil, errors.New("remove file error: " + err.Error())
	}
	return &downloadReply, nil
}

func (s *SubtitleBestApi) ConvertId(id, source, videoType string) (*IdConvertReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	if len(s.authKey.AESKey16) != 16 {
		return nil, errors.New(fmt.Sprintf("AESKey16 is not set, %s", s.authKey.AESKey16))
	}
	if len(s.authKey.AESIv16) != 16 {
		return nil, errors.New(fmt.Sprintf("AESIv16 is not set, %s", s.authKey.AESIv16))
	}

	postUrl := webUrlBase + "/v1/id-convert"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	var idConvertReply IdConvertReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetBody(IdConvertReq{
			Id:        id,
			Source:    source,
			VideoType: videoType,
		}).
		SetResult(&idConvertReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("convert id error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if idConvertReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	return &idConvertReply, nil
}

func (s *SubtitleBestApi) FeedBack(id, version, MediaServer string, EnableShare, EnableApiKey bool) (*FeedReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	if len(s.authKey.AESKey16) != 16 {
		return nil, errors.New(fmt.Sprintf("AESKey16 is not set, %s", s.authKey.AESKey16))
	}
	if len(s.authKey.AESIv16) != 16 {
		return nil, errors.New(fmt.Sprintf("AESIv16 is not set, %s", s.authKey.AESIv16))
	}

	postUrl := webUrlBase + "/v1/feedback"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	formData := make(map[string]string)
	formData["id"] = id
	formData["version"] = version
	formData["media_server"] = MediaServer
	formData["enable_share"] = strconv.FormatBool(EnableShare)
	formData["enable_api_key"] = strconv.FormatBool(EnableApiKey)
	var feedReply FeedReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFormData(formData).
		SetResult(&feedReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("feedback error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if feedReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	return &feedReply, nil
}

func (s *SubtitleBestApi) AskDownloadTask(id string) (*AskDownloadTaskReply, error) {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return nil, errors.New("auth key is not set")
	}
	if len(s.authKey.AESKey16) != 16 {
		return nil, errors.New(fmt.Sprintf("AESKey16 is not set, %s", s.authKey.AESKey16))
	}
	if len(s.authKey.AESIv16) != 16 {
		return nil, errors.New(fmt.Sprintf("AESIv16 is not set, %s", s.authKey.AESIv16))
	}

	postUrl := webUrlBase + "/v1/ask-download-task"
	httpClient, err := my_util.NewHttpClient(s.proxySettings)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	major, minor, patch := global_value.AppVersionInt()
	var askDownloadTaskReply AskDownloadTaskReply
	resp, err := httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFormData(map[string]string{
			"fid":               id,
			"app_version_major": fmt.Sprintf("%d", major),
			"app_version_minor": fmt.Sprintf("%d", minor),
			"app_version_patch": fmt.Sprintf("%d", patch),
		}).
		SetResult(&askDownloadTaskReply).
		Post(postUrl)
	if err != nil {
		s.log.Errorln("ask download task error, status code:", resp.StatusCode(), "Error:", err)
		return nil, err
	}

	if askDownloadTaskReply.Status == 0 {
		s.log.Warningln("status code:", resp.StatusCode())
	}

	return &askDownloadTaskReply, nil
}

const (
	webUrlBase = "https://api.subtitle.best"
	//webUrlBase = "http://127.0.0.1:8893"
)
