package downloader

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/strcut_json"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"path/filepath"
)

func (d *Downloader) SetMovieAndSeasonInfo(movieInfos []backend.MovieInfo, seasonInfos []backend.SeasonInfo) {
	d.cacheLocker.Lock()
	defer d.cacheLocker.Unlock()

	d.setMovieAndSeasonInfo(movieInfos, seasonInfos)
}

func (d *Downloader) GetMovieInfoAndSeasonInfo() ([]backend.MovieInfo, []backend.SeasonInfo) {
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

func (d *Downloader) setMovieAndSeasonInfo(movieInfos []backend.MovieInfo, seasonInfos []backend.SeasonInfo, skip ...bool) {
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

func (d *Downloader) getMovieInfoAndSeasonInfo(AllorFrontorEnd int) ([]backend.MovieInfo, []backend.SeasonInfo) {

	outMovieInfos := make([]backend.MovieInfo, 0)
	outSeasonInfo := make([]backend.SeasonInfo, 0)
	// AllorFrontorEnd == 0, 全部, AllorFrontorEnd == 1, MovieInfo, AllorFrontorEnd == 2, SeasonInfo
	if AllorFrontorEnd == 0 || AllorFrontorEnd == 1 {

		for _, movieInfo := range d.movieInfoMap {

			nowMovieInfo := backend.MovieInfo{
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

			nowSeasonInfo := backend.SeasonInfo{
				Name:          seasonInfo.Name,
				RootDirPath:   seasonInfo.RootDirPath,
				DirRootUrl:    seasonInfo.DirRootUrl,
				OneVideoInfos: make([]backend.OneVideoInfo, 0),
			}

			for _, oneVideoInfo := range seasonInfo.OneVideoInfoMap {

				nowOneVideoInfo := backend.OneVideoInfo{

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

func (d *Downloader) saveVideoListCache(movieInfos []backend.MovieInfo, seasonInfos []backend.SeasonInfo) error {

	// 缓存下来
	cacheCenterFolder, err := my_folder.GetRootCacheCenterFolder()
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
	cacheCenterFolder, err := my_folder.GetRootCacheCenterFolder()
	if err != nil {
		return err
	}

	movieInfosFileName := filepath.Join(cacheCenterFolder, "movie_infos.json")
	seasonInfosFileName := filepath.Join(cacheCenterFolder, "season_infos.json")

	movieInfos := make([]backend.MovieInfo, 0)
	seasonInfos := make([]backend.SeasonInfo, 0)

	if my_util.IsFile(movieInfosFileName) == true {
		err = strcut_json.ToStruct(movieInfosFileName, &movieInfos)
		if err != nil {
			return err
		}
	}

	if my_util.IsFile(seasonInfosFileName) == true {
		err = strcut_json.ToStruct(seasonInfosFileName, &seasonInfos)
		if err != nil {
			return err
		}
	}

	d.setMovieAndSeasonInfo(movieInfos, seasonInfos, true)

	return nil
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
