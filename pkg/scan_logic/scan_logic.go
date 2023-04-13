package scan_logic

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
)

type ScanLogic struct {
	l            *logrus.Logger
	scanLogicMap sync.Map
}

func NewScanLogic(l *logrus.Logger) *ScanLogic {

	s := &ScanLogic{
		l: l,
	}
	// 那么尝试读取数据库，进行缓存，仅执行一次
	var skipInfos []*models.SkipScanInfo
	dao.GetDb().Find(&skipInfos)
	for _, skipInfo := range skipInfos {
		s.scanLogicMap.Store(skipInfo.UID, skipInfo.Skip)
	}

	return s
}

// Set 设置跳过扫描的信息
func (s *ScanLogic) Set(skipInfo *models.SkipScanInfo) {

	s.scanLogicMap.Store(skipInfo.UID, skipInfo.Skip)
	dao.GetDb().Save(skipInfo)
}

// Get 是否跳过，获取跳过扫描的信息设置，带有缓存。电影就是具体的视频文件全路径，连续剧就是具体一集视频文件的全路径
func (s *ScanLogic) Get(videoType int, videoPath string) bool {

	var uid string
	if videoType == 0 {
		// 电影
		uid = models.GenerateUID4Movie(videoPath)
	} else {
		// 电视剧
		skipInfo := models.NewSkipScanInfoBySeriesEx(videoPath, true)
		uid = skipInfo.UID
	}

	value, found := s.scanLogicMap.Load(uid)
	if found == false {
		// 缓存没有找到那么就从数据库查询
		var skipInfos []models.SkipScanInfo
		dao.GetDb().Where("uid = ?", uid).Find(&skipInfos)
		if len(skipInfos) < 1 {
			// 数据库中没有找到，但是也需要写入一份到数据库中，默认是扫描
			skipInfo := models.NewSkipScanInfoByUID(uid, false)
			dao.GetDb().Save(skipInfo)
			// 缓存下来
			s.scanLogicMap.Store(uid, false)
			return false
		} else {
			// 数据库中找到了，缓存下来
			s.scanLogicMap.Store(uid, skipInfos[0].Skip)
			return skipInfos[0].Skip
		}
	} else {
		return value.(bool)
	}
}
