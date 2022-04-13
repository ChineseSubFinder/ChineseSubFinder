package task_queue

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_file_hash"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"path/filepath"
	"time"
)

type OneJob struct {
	Id           string           `json:"id"`                        // 任务的唯一 ID
	VideoType    common.VideoType `json:"video_type"`                // 视频的类型
	VideoFPath   string           `json:"video_f_path"`              // 视频的全路径
	VideoName    string           `json:"video_name"`                // 视频的名称
	Feature      string           `json:"feature"`                   // 视频的特征码，蓝光的时候可能是空
	Season       int              `json:"season"`                    // 如果对应的是电影则可能是 0，没有
	Episode      int              `json:"episode"`                   // 如果对应的是电影则可能是 0，没有
	JobStatus    JobStatus        `json:"job_status"`                // 任务的状态
	TaskPriority int              `json:"task_priority" default:"5"` // 任务的优先级，0 - 10 个级别，0 是最高，10 是最低
	RetryTimes   int              `json:"retry_times"`               // 重试了多少次
	AddedTime    time.Time        `json:"added_time"`                // 任务添加的时间
	UpdateTime   time.Time        `json:"update_time"`               // 任务更新的时间
}

func NewOneJob(videoType common.VideoType, videoFPath string, taskPriority int) *OneJob {

	ob := &OneJob{VideoType: videoType, VideoFPath: videoFPath, TaskPriority: taskPriority}
	ob.Id = my_util.Get2UUID()
	ob.VideoName = filepath.Base(videoFPath)
	// -------------------------------------------------
	// 使用本程序的 hash 的算法，得到视频的唯一 ID
	ob.Feature, _ = sub_file_hash.Calculate(videoFPath)
	// -------------------------------------------------
	if videoType == common.Series {
		// 连续剧的时候，如果可能应该获取是 第几季  第几集
		torrentInfo, _, err := decode.GetVideoInfoFromFileFullPath(videoFPath)
		if err == nil {
			ob.Season = torrentInfo.Season
			ob.Episode = torrentInfo.Episode
		}
	}
	// -------------------------------------------------
	ob.JobStatus = Waiting
	nTime := time.Now()
	ob.AddedTime = nTime
	ob.UpdateTime = nTime

	return ob
}
