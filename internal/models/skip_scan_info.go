package models

import (
	"crypto/sha256"
	"fmt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/WQGroup/logger"
	PTN "github.com/middelink/go-parse-torrent-name"
	"path/filepath"
)

type SkipScanInfo struct {
	/*
		这里的 UID 计算方式有两种:
		1. 电影，由电影的文件夹路径计算 sha256 得到，X:\电影\Three Thousand Years of Longing (2022)
		2. 连续剧，由连续剧的文件夹路径计算 sha256 得到，只能具体到一集（S01E01 这里是拼接出来的不是真正的文件名）
			X:\连续剧\绝命毒师S01E01
	*/
	UID    string `gorm:"type:varchar(64);primarykey"`
	Skip   bool   `gorm:"type:bool;default:false"`
	season int    `gorm:"-"`
	eps    int    `gorm:"-"`
}

func (s *SkipScanInfo) Season() int {
	return s.season
}

func (s *SkipScanInfo) Eps() int {
	return s.eps
}

func NewSkipScanInfoByUID(uid string, skip bool) *SkipScanInfo {

	var skipScanInfo SkipScanInfo
	skipScanInfo.UID = uid
	skipScanInfo.Skip = skip

	return &skipScanInfo
}

func GenerateUID4Movie(movieFPath string) string {
	movieDirPath := filepath.Dir(movieFPath)
	fileUID := fmt.Sprintf("%x", sha256.Sum256([]byte(movieDirPath)))
	return fileUID
}

func GenerateUID4Series(seriesDirFPath string, season, eps int) string {

	mixInfo := fmt.Sprintf("%sS%02dE%02d", seriesDirFPath, season, eps)
	fileUID := fmt.Sprintf("%x", sha256.Sum256([]byte(mixInfo)))
	return fileUID
}

func NewSkipScanInfoByMovie(movieFPath string, skip bool) *SkipScanInfo {

	var skipScanInfo SkipScanInfo
	skipScanInfo.UID = GenerateUID4Movie(movieFPath)
	skipScanInfo.Skip = skip

	return &skipScanInfo
}

func NewSkipScanInfoBySeries(seriesDirFPath string, season, eps int, skip bool) *SkipScanInfo {

	var skipScanInfo SkipScanInfo
	skipScanInfo.UID = GenerateUID4Series(seriesDirFPath, season, eps)
	skipScanInfo.Skip = skip
	skipScanInfo.season = season
	skipScanInfo.eps = eps

	return &skipScanInfo
}

func NewSkipScanInfoBySeriesEx(oneEpsFPath string, skip bool) *SkipScanInfo {

	seriesDirFPath := filepath.Dir(filepath.Dir(oneEpsFPath))
	var season, eps int
	oneSeriesEpisode, err := decode.GetVideoNfoInfo4OneSeriesEpisode(oneEpsFPath)
	if err != nil {
		// 换一种方式获取 Season 和 Eps 信息
		var parse *PTN.TorrentInfo
		parse, err = PTN.Parse(oneEpsFPath)
		if err != nil {
			season = -1
			eps = -1
			logger.GetLogger().Errorln("NewSkipScanInfoBySeriesEx Parse Error: ", err)
		}
		season = parse.Season
		eps = parse.Episode
	} else {
		season = oneSeriesEpisode.Season
		eps = oneSeriesEpisode.Episode
	}
	var skipScanInfo SkipScanInfo
	skipScanInfo.UID = GenerateUID4Series(seriesDirFPath, season, eps)
	skipScanInfo.Skip = skip
	skipScanInfo.season = season
	skipScanInfo.eps = eps

	return &skipScanInfo
}
