package downloader

import (
	"path/filepath"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"
	vsh "github.com/ChineseSubFinder/ChineseSubFinder/pkg/video_scan_and_refresh_helper"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/strcut_json"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
)

func (d *Downloader) SetMovieAndSeasonInfo(movieInfos []backend2.MovieInfo, seasonInfos []backend2.SeasonInfo) {
	d.cacheLocker.Lock()
	defer d.cacheLocker.Unlock()

	d.setMovieAndSeasonInfo(movieInfos, seasonInfos)
}

func (d *Downloader) GetMovieInfoAndSeasonInfo() ([]backend2.MovieInfo, []backend2.SeasonInfo) {
	// 需要把本实例中的缓存 map 转换到 Web 传递的结构体中
	d.cacheLocker.Lock()
	defer d.cacheLocker.Unlock()
	// 全部都获取
	return d.getMovieInfoAndSeasonInfo(0)
}

func (d *Downloader) UpdateInfo(job task_queue.OneJob) {
	d.cacheLocker.Lock()
	defer d.cacheLocker.Unlock()

	// 搜索字幕
	matchedSubFileByOneVideo, err := sub_helper.SearchMatchedSubFileByOneVideo(d.log, job.VideoFPath)
	if err != nil {
		d.log.Errorln("SearchMatchedSubFileByOneVideo", err)
		return
	}

	if job.VideoType == common.Movie {
		// 更新 movieInfo
		// 更新缓存, 存在，更新, 不存在，跳过
		if oneMovieInfo, ok := d.movieInfoMap[job.VideoFPath]; ok == true {

			oneMovieInfo.MediaServerInsideVideoID = job.MediaServerInsideVideoID
			oneMovieInfo.SubFPathList = matchedSubFileByOneVideo
			d.movieInfoMap[job.VideoFPath] = oneMovieInfo
			// 写入本地缓存
			backendMovieInfo, _ := d.getMovieInfoAndSeasonInfo(1)
			err = d.saveVideoListCache(backendMovieInfo, nil)
			if err != nil {
				d.log.Errorln("saveVideoListCache", job.VideoFPath, err)
				return
			}
		}
	} else if job.VideoType == common.Series {
		// 更新 seasonInfo
		// 更新缓存, 存在，更新, 不存在，跳过
		if oneSeasonInfo, ok := d.seasonInfoMap[job.SeriesRootDirPath]; ok == true {

			if nowOneVideoInfo, ok := oneSeasonInfo.OneVideoInfoMap[job.VideoFPath]; ok == true {
				nowOneVideoInfo.MediaServerInsideVideoID = job.MediaServerInsideVideoID
				nowOneVideoInfo.SubFPathList = matchedSubFileByOneVideo
				d.seasonInfoMap[job.SeriesRootDirPath].OneVideoInfoMap[job.VideoFPath] = nowOneVideoInfo
				// 写入本地缓存
				_, backendSeasonInfo := d.getMovieInfoAndSeasonInfo(2)
				err = d.saveVideoListCache(nil, backendSeasonInfo)
				if err != nil {
					d.log.Errorln("saveVideoListCache", job.VideoFPath, err)
					return
				}
			}
		}
	}
}

func (d *Downloader) setMovieAndSeasonInfo(movieInfos []backend2.MovieInfo, seasonInfos []backend2.SeasonInfo, skip ...bool) {
	// 需要把 Web 传递的结构体 转换到 本实例中的缓存 map

	// 清空
	d.movieInfoMap = make(map[string]MovieInfo)
	d.seasonInfoMap = make(map[string]SeasonInfo)

	for _, movieInfo := range movieInfos {

		nowMovieInfo := MovieInfo{
			Name:                     movieInfo.Name,
			DirRootUrl:               movieInfo.DirRootUrl,
			VideoFPath:               movieInfo.VideoFPath,
			VideoUrl:                 movieInfo.VideoUrl,
			MediaServerInsideVideoID: movieInfo.MediaServerInsideVideoID,
			SubFPathList:             make([]string, 0),
		}
		nowMovieInfo.SubFPathList = append(nowMovieInfo.SubFPathList, movieInfo.SubFPathList...)
		d.movieInfoMap[movieInfo.VideoFPath] = nowMovieInfo
	}

	for _, seasonInfo := range seasonInfos {

		nowSeasonInfo := SeasonInfo{
			Name:            seasonInfo.Name,
			RootDirPath:     seasonInfo.RootDirPath,
			DirRootUrl:      seasonInfo.DirRootUrl,
			OneVideoInfoMap: make(map[string]OneVideoInfo),
		}

		for _, oneVideoInfo := range seasonInfo.OneVideoInfos {

			nowOneVideoInfo := OneVideoInfo{
				Name:                     oneVideoInfo.Name,
				VideoFPath:               oneVideoInfo.VideoFPath,
				VideoUrl:                 oneVideoInfo.VideoUrl,
				Season:                   oneVideoInfo.Season,
				Episode:                  oneVideoInfo.Episode,
				MediaServerInsideVideoID: oneVideoInfo.MediaServerInsideVideoID,
				SubFPathList:             make([]string, 0),
			}
			nowOneVideoInfo.SubFPathList = append(nowOneVideoInfo.SubFPathList, oneVideoInfo.SubFPathList...)

			nowSeasonInfo.OneVideoInfoMap[oneVideoInfo.VideoFPath] = nowOneVideoInfo
		}

		d.seasonInfoMap[seasonInfo.RootDirPath] = nowSeasonInfo
	}

	if len(skip) > 0 && skip[0] == true {

	} else {
		err := d.saveVideoListCache(movieInfos, seasonInfos)
		if err != nil {
			d.log.Errorln("saveVideoListCache err:", err)
			return
		}
	}
}

func (d *Downloader) getMovieInfoAndSeasonInfo(AllorFrontorEnd int) ([]backend2.MovieInfo, []backend2.SeasonInfo) {

	outMovieInfos := make([]backend2.MovieInfo, 0)
	outSeasonInfo := make([]backend2.SeasonInfo, 0)
	// AllorFrontorEnd == 0, 全部, AllorFrontorEnd == 1, MovieInfo, AllorFrontorEnd == 2, SeasonInfo
	if AllorFrontorEnd == 0 || AllorFrontorEnd == 1 {

		for _, movieInfo := range d.movieInfoMap {

			nowMovieInfo := backend2.MovieInfo{
				Name:                     movieInfo.Name,
				DirRootUrl:               movieInfo.DirRootUrl,
				VideoFPath:               movieInfo.VideoFPath,
				VideoUrl:                 movieInfo.VideoUrl,
				MediaServerInsideVideoID: movieInfo.MediaServerInsideVideoID,
				SubFPathList:             make([]string, 0),
			}
			nowMovieInfo.SubFPathList = append(nowMovieInfo.SubFPathList, movieInfo.SubFPathList...)
			outMovieInfos = append(outMovieInfos, nowMovieInfo)
		}
	}

	if AllorFrontorEnd == 0 || AllorFrontorEnd == 2 {

		for _, seasonInfo := range d.seasonInfoMap {

			nowSeasonInfo := backend2.SeasonInfo{
				Name:          seasonInfo.Name,
				RootDirPath:   seasonInfo.RootDirPath,
				DirRootUrl:    seasonInfo.DirRootUrl,
				OneVideoInfos: make([]backend2.OneVideoInfo, 0),
			}

			for _, oneVideoInfo := range seasonInfo.OneVideoInfoMap {

				nowOneVideoInfo := backend2.OneVideoInfo{

					Name:                     oneVideoInfo.Name,
					VideoFPath:               oneVideoInfo.VideoFPath,
					VideoUrl:                 oneVideoInfo.VideoUrl,
					Season:                   oneVideoInfo.Season,
					Episode:                  oneVideoInfo.Episode,
					MediaServerInsideVideoID: oneVideoInfo.MediaServerInsideVideoID,
					SubFPathList:             make([]string, 0),
				}
				nowOneVideoInfo.SubFPathList = append(nowOneVideoInfo.SubFPathList, oneVideoInfo.SubFPathList...)
				nowSeasonInfo.OneVideoInfos = append(nowSeasonInfo.OneVideoInfos, nowOneVideoInfo)
			}

			outSeasonInfo = append(outSeasonInfo, nowSeasonInfo)
		}
	}

	return outMovieInfos, outSeasonInfo
}

func (d *Downloader) saveVideoListCache(movieInfos []backend2.MovieInfo, seasonInfos []backend2.SeasonInfo) error {

	// 缓存下来
	cacheCenterFolder, err := pkg.GetRootCacheCenterFolder()
	if err != nil {
		return err
	}

	movieInfosFileName := filepath.Join(cacheCenterFolder, "movie_infos.json")
	seasonInfosFileName := filepath.Join(cacheCenterFolder, "season_infos.json")

	if movieInfos != nil && len(movieInfos) > 0 {
		err = strcut_json.ToFile(movieInfosFileName, movieInfos)
		if err != nil {
			return err
		}
	}

	if seasonInfos != nil && len(seasonInfos) > 0 {
		err = strcut_json.ToFile(seasonInfosFileName, seasonInfos)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Downloader) loadVideoListCache() error {

	// 缓存下来
	cacheCenterFolder, err := pkg.GetRootCacheCenterFolder()
	if err != nil {
		return err
	}

	movieInfosFileName := filepath.Join(cacheCenterFolder, "movie_infos.json")
	seasonInfosFileName := filepath.Join(cacheCenterFolder, "season_infos.json")

	movieInfos := make([]backend2.MovieInfo, 0)
	seasonInfos := make([]backend2.SeasonInfo, 0)

	if pkg.IsFile(movieInfosFileName) == true {
		err = strcut_json.ToStruct(movieInfosFileName, &movieInfos)
		if err != nil {
			return err
		}
	}

	if pkg.IsFile(seasonInfosFileName) == true {
		err = strcut_json.ToStruct(seasonInfosFileName, &seasonInfos)
		if err != nil {
			return err
		}
	}

	d.setMovieAndSeasonInfo(movieInfos, seasonInfos, true)

	return nil
}

// SetMovieAndSeasonInfoV2 只把第一级目录的信息给缓存下来，比如 x:\电影\壮志凌云\壮志凌云.mp4 或者是连续剧的 x:\连续剧\绝命毒师 根目录
func (d *Downloader) SetMovieAndSeasonInfoV2(mainList *vsh.NormalScanVideoResult) error {

	d.cacheLocker.Lock()
	defer d.cacheLocker.Unlock()
	// 缓存下来
	cacheCenterFolder, err := pkg.GetRootCacheCenterFolder()
	if err != nil {
		return err
	}
	var movieInfos = make([]backend2.MovieInfoV2, 0)
	var seasonInfos = make([]backend2.SeasonInfoV2, 0)

	movieInfosFileName := filepath.Join(cacheCenterFolder, "movie_main_list.json")
	seasonInfosFileName := filepath.Join(cacheCenterFolder, "season_main_list.json")

	if mainList != nil && mainList.MoviesDirMap != nil {

		mainList.MoviesDirMap.Any(func(movieDirRootPath interface{}, moviesFPath interface{}) bool {

			oneMovieDirRootPath := movieDirRootPath.(string)
			for _, movieFPath := range moviesFPath.([]string) {
				movieInfos = append(movieInfos, backend2.MovieInfoV2{
					Name:             filepath.Base(movieFPath),
					MainRootDirFPath: oneMovieDirRootPath,
					VideoFPath:       movieFPath,
				})
			}
			return false
		})

		err = strcut_json.ToFile(movieInfosFileName, movieInfos)
		if err != nil {
			return err
		}
	}

	if mainList != nil && mainList.SeriesDirMap != nil {

		mainList.SeriesDirMap.Any(func(seriesRootPathName interface{}, seriesNames interface{}) bool {

			oneSeriesRootPathName := seriesRootPathName.(string)
			for _, oneSeriesRootDir := range seriesNames.([]string) {
				seasonInfos = append(seasonInfos, backend2.SeasonInfoV2{
					Name:             filepath.Base(oneSeriesRootDir),
					MainRootDirFPath: oneSeriesRootPathName,
					RootDirPath:      oneSeriesRootDir,
				})
			}
			return false
		})

		err = strcut_json.ToFile(seasonInfosFileName, seasonInfos)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetMovieInfoAndSeasonInfoV2 只把第一级目录的信息给缓存下来，比如 x:\电影\壮志凌云\壮志凌云.mp4 或者是连续剧的 x:\连续剧\绝命毒师 根目录
func (d *Downloader) GetMovieInfoAndSeasonInfoV2() ([]backend2.MovieInfoV2, []backend2.SeasonInfoV2, error) {
	// 需要把本实例中的缓存 map 转换到 Web 传递的结构体中
	d.cacheLocker.Lock()
	defer d.cacheLocker.Unlock()

	// 缓存下来
	cacheCenterFolder, err := pkg.GetRootCacheCenterFolder()
	if err != nil {
		return nil, nil, err
	}

	movieInfosFileName := filepath.Join(cacheCenterFolder, "movie_main_list.json")
	seasonInfosFileName := filepath.Join(cacheCenterFolder, "season_main_list.json")

	movieInfos := make([]backend2.MovieInfoV2, 0)
	seasonInfos := make([]backend2.SeasonInfoV2, 0)

	if pkg.IsFile(movieInfosFileName) == true {
		err = strcut_json.ToStruct(movieInfosFileName, &movieInfos)
		if err != nil {
			return nil, nil, err
		}
	}

	if pkg.IsFile(seasonInfosFileName) == true {
		err = strcut_json.ToStruct(seasonInfosFileName, &seasonInfos)
		if err != nil {
			return nil, nil, err
		}
	}

	return movieInfos, seasonInfos, nil
}

type MovieInfo struct {
	Name                     string   `json:"name"`
	DirRootUrl               string   `json:"dir_root_url"`
	VideoFPath               string   `json:"video_f_path"`
	VideoUrl                 string   `json:"video_url"`
	MediaServerInsideVideoID string   `json:"media_server_inside_video_id"`
	SubFPathList             []string `json:"sub_f_path_list"`
}

type SeasonInfo struct {
	Name            string                  `json:"name"`
	RootDirPath     string                  `json:"root_dir_path"`
	DirRootUrl      string                  `json:"dir_root_url"`
	OneVideoInfoMap map[string]OneVideoInfo `json:"one_video_info"` // Key VideoFPath
}

type OneVideoInfo struct {
	Name                     string   `json:"name"`
	VideoFPath               string   `json:"video_f_path"`
	VideoUrl                 string   `json:"video_url"`
	Season                   int      `json:"season"`
	Episode                  int      `json:"episode"`
	SubFPathList             []string `json:"sub_f_path_list"`
	MediaServerInsideVideoID string   `json:"media_server_inside_video_id"`
}
