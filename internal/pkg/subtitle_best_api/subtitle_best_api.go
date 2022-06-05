package subtitle_best_api

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
)

type SubtitleBestApi struct {
	authKey       random_auth_key.AuthKey
	randomAuthKey *random_auth_key.RandomAuthKey
}

func NewSubtitleBestApi(inAuthKey random_auth_key.AuthKey) *SubtitleBestApi {
	return &SubtitleBestApi{
		randomAuthKey: random_auth_key.NewRandomAuthKey(5, inAuthKey),
		authKey:       inAuthKey,
	}
}

func (s *SubtitleBestApi) GetMediaInfo(id, source, videoType string, _proxySettings ...*settings.ProxySettings) (*MediaInfoReply, error) {

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
	httpClient, err := my_util.NewHttpClient(_proxySettings...)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	var mediaInfoReply MediaInfoReply
	_, err = httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetBody(MediaInfoReq{
			Id:        id,
			Source:    source,
			VideoType: videoType,
		}).
		SetResult(&mediaInfoReply).
		Post(postUrl)
	if err != nil {
		return nil, err
	}

	return &mediaInfoReply, nil
}

// AskFroUpload 在使用这个接口前，需要从 IMDB ID 获取到 TMDB ID
func (s *SubtitleBestApi) AskFroUpload(subSha256 string, trusted bool, ImdbId, TmdbId string, Season, Episode int, _proxySettings ...*settings.ProxySettings) (*AskForUploadReply, error) {

	postUrl := webUrlBase + "/v1/ask-for-upload"
	httpClient, err := my_util.NewHttpClient(_proxySettings...)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	var askReq AskForUploadReq
	if trusted == true {
		askReq = AskForUploadReq{
			SubSha256: subSha256,
			Trusted:   trusted,
			ImdbId:    ImdbId,
			TmdbId:    TmdbId,
			Season:    Season,
			Episode:   Episode,
		}
	} else {
		askReq = AskForUploadReq{
			SubSha256: subSha256,
		}
	}

	var askForUploadReply AskForUploadReply
	_, err = httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetBody(askReq).
		SetResult(&askForUploadReply).
		Post(postUrl)
	if err != nil {
		return nil, err
	}

	return &askForUploadReply, nil
}

// UploadSub 在使用这个接口前，需要从 IMDB ID 获取到 TMDB ID，其实在这一步应该默认就拿到了 TMDB ID，需要提前在 AskFroUpload 接口调用前就搞定这个
// year 这个也是从之前的接口拿到, 2019  or  2022
func (s *SubtitleBestApi) UploadSub(videoSubInfo *models.VideoSubInfo, subSaveRootDirPath string, tmdbId, year string, _proxySettings ...*settings.ProxySettings) (*UploadSubReply, error) {

	postUrl := webUrlBase + "/v1/upload-sub"
	httpClient, err := my_util.NewHttpClient(_proxySettings...)
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

	var uploadSubReply UploadSubReply
	_, err = httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFileReader("sub_file_context", videoSubInfo.SubName, bytes.NewReader(fd)).
		SetFormData(map[string]string{
			"sub_sha256":     videoSubInfo.SHA256,
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
		return nil, err
	}

	return &uploadSubReply, nil
}

func (s *SubtitleBestApi) UploadLowTrustSub(lowTrustVideoSubInfo *models.LowVideoSubInfo, subSaveRootDirPath string, tmdbId, year string, _proxySettings ...*settings.ProxySettings) (*UploadSubReply, error) {

	postUrl := webUrlBase + "/v1/upload-sub"
	httpClient, err := my_util.NewHttpClient(_proxySettings...)
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

	var uploadSubReply UploadSubReply
	_, err = httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetFileReader("sub_file_context", lowTrustVideoSubInfo.SubName, bytes.NewReader(fd)).
		SetFormData(map[string]string{
			"sub_sha256":     lowTrustVideoSubInfo.SHA256,
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
		}).
		SetResult(&uploadSubReply).
		Post(postUrl)
	if err != nil {
		return nil, err
	}

	return &uploadSubReply, nil
}

const (
	webUrlBase = "https://api.subtitle.best"
	//webUrlBase = "http://127.0.0.1:8889"
)
