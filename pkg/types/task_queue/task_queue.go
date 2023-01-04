package task_queue

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/emby"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_file_hash"
	"github.com/araddon/dateparse"
)

type OneJob struct {
	Id                       string           `json:"id"`                           // 任务的唯一 ID
	VideoType                common.VideoType `json:"video_type"`                   // 视频的类型
	VideoFPath               string           `json:"video_f_path"`                 // 视频的全路径
	VideoName                string           `json:"video_name"`                   // 视频的名称
	Feature                  string           `json:"feature"`                      // 视频的特征码，蓝光的时候可能是空
	SeriesRootDirPath        string           `json:"series_root_dir_path"`         // 连续剧的目录
	Season                   int              `json:"season"`                       // 如果对应的是电影则可能是 0，没有
	Episode                  int              `json:"episode"`                      // 如果对应的是电影则可能是 0，没有
	JobStatus                JobStatus        `json:"job_status"`                   // 任务的状态
	TaskPriority             int              `json:"task_priority" default:"5"`    // 任务的优先级，0 - 10 个级别，0 是最高，10 是最低
	RetryTimes               int              `json:"retry_times"`                  // 重试了多少次
	CreatedTime              emby.Time        `json:"created_time"`                 // 视频的发布时间或者是文件的创建时间
	AddedTime                emby.Time        `json:"added_time"`                   // 任务添加的时间
	UpdateTime               emby.Time        `json:"update_time"`                  // 任务更新的时间
	MediaServerInsideVideoID string           `json:"media_server_inside_video_id"` // 媒体服务器中，这个视频的 ID，如果是 Emby 就对应它内部这个视频的 ID，后续用于指定刷新视频信息
	ErrorInfo                string           `json:"error_info"`                   // 这个任务的错误信息
	DownloadTimes            int              `json:"download_times"`               // 下载的次数，用于统计下载过几次
}

func NewOneJob(videoType common.VideoType, videoFPath string, taskPriority int, MediaServerInsideVideoID ...string) *OneJob {

	ob := &OneJob{VideoType: videoType, VideoFPath: videoFPath, TaskPriority: taskPriority}

	sha256FilePathID := func() string {
		return fmt.Sprintf("%x", sha256.Sum256([]byte(videoFPath)))
	}

	/*
		sub_file_hash.Calculate 现在支持内部的 fake 蓝光视频地址了，会解析到 BDMV 中最大的那个视频流文件来计算
		所以上面这个函数如果 errors 了，才需要使用这个伪造的路径进行 sha256 加密即可
	*/
	sha256String, err := sub_file_hash.Calculate(videoFPath)
	if err != nil {
		ob.Id = sha256FilePathID()
	} else {
		ob.Id = sha256String
	}

	ob.VideoName = filepath.Base(videoFPath)
	// -------------------------------------------------
	// 使用本程序的 hash 的算法，得到视频的唯一 ID
	ob.Feature = sha256String
	// -------------------------------------------------
	ob.JobStatus = Waiting
	nTime := time.Now()
	ob.AddedTime = emby.Time(nTime)
	ob.UpdateTime = emby.Time(nTime)
	// 需要获取这个视频的创建时间或者发布时间
	if ob.VideoType == common.Movie {

		imdbInfo4Movie, err := decode.GetVideoNfoInfo4Movie(videoFPath)
		if err == nil {
			createTime, _ := dateparse.ParseAny(imdbInfo4Movie.ReleaseDate)
			ob.CreatedTime = emby.Time(createTime)
		}
	} else if ob.VideoType == common.Series {
		imdbInfo4Eps, err := decode.GetVideoNfoInfo4OneSeriesEpisode(videoFPath)
		if err == nil {
			createTime, _ := dateparse.ParseAny(imdbInfo4Eps.ReleaseDate)
			ob.CreatedTime = emby.Time(createTime)
		}
	}

	if len(MediaServerInsideVideoID) > 0 && MediaServerInsideVideoID[0] != "" {
		ob.MediaServerInsideVideoID = MediaServerInsideVideoID[0]
	}

	return ob
}
