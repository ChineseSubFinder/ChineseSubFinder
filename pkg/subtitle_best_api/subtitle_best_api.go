package subtitle_best_api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/sirupsen/logrus"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"
)

type SubtitleBestApi struct {
	log           *logrus.Logger
	authKey       random_auth_key.AuthKey
	randomAuthKey *random_auth_key.RandomAuthKey
}

func NewSubtitleBestApi(log *logrus.Logger, inAuthKey random_auth_key.AuthKey) *SubtitleBestApi {
	return &SubtitleBestApi{
		log:           log,
		randomAuthKey: random_auth_key.NewRandomAuthKey(5, inAuthKey),
		authKey:       inAuthKey,
	}
}

func (s *SubtitleBestApi) CheckAlive() error {

	if s.authKey.BaseKey == random_auth_key.BaseKey || s.authKey.AESKey16 == random_auth_key.AESKey16 || s.authKey.AESIv16 == random_auth_key.AESIv16 {
		return errors.New("auth key is not set")
	}

	postUrl := webUrlBase + "/v1/subhd-code"
	httpClient, err := pkg.NewHttpClient()
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
	httpClient, err := pkg.NewHttpClient()
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
	httpClient, err := pkg.NewHttpClient()
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

// ConvertId 目前仅仅支持 TMDB ID 转 IMDB ID
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
	httpClient, err := pkg.NewHttpClient()
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
	httpClient, err := pkg.NewHttpClient()
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

const (
	webUrlBase = "https://api.subtitle.best"
)
