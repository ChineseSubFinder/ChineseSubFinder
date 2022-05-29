package subtitle_best_api

import (
	"errors"
	"fmt"

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

	const postUrl = webUrlBase + "/v1/media-info"
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

func (s SubtitleBestApi) AskFroUpload(subSha256 string, _proxySettings ...*settings.ProxySettings) (*AskForUploadReply, error) {

	const postUrl = webUrlBase + "/v1/media-info"
	httpClient, err := my_util.NewHttpClient(_proxySettings...)
	if err != nil {
		return nil, err
	}

	authKey, err := s.randomAuthKey.GetAuthKey()
	if err != nil {
		return nil, err
	}

	var askForUploadReply AskForUploadReply
	_, err = httpClient.R().
		SetHeader("Authorization", "beer "+authKey).
		SetBody(AskForUploadReq{
			SubSha256: subSha256,
		}).
		SetResult(&askForUploadReply).
		Post(postUrl)
	if err != nil {
		return nil, err
	}

	return &askForUploadReply, nil
}

const (
	webUrlBase = "https://api.subtitle.best"
)
