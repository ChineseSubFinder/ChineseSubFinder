package emby

import (
	"strings"
	"time"
)

type EmbyRecentlyItems struct {
	Items            []EmbyRecentlyItem `json:"Items,omitempty"`
	TotalRecordCount int                `json:"TotalRecordCount,omitempty"`
}

type EmbyRecentlyItem struct {
	Name              string `json:"Name,omitempty"`
	Id                string `json:"Id,omitempty"`
	IndexNumber       int    `json:"IndexNumber,omitempty"`
	ParentIndexNumber int    `json:"ParentIndexNumber,omitempty"`
	Type              string `json:"Type,omitempty"`
	UserData          struct {
		PlaybackPositionTicks int  `json:"PlaybackPositionTicks"`
		PlayCount             int  `json:"PlayCount"`
		IsFavorite            bool `json:"IsFavorite"`
		Played                bool `json:"Played"`
	} `json:"UserData"`
	SeriesName string `json:"SeriesName,omitempty"`
}

type EmbyItemsAncestors struct {
	Name string `json:"Name,omitempty"`
	ID   string `json:"Id,omitempty"`
	Path string `json:"Path,omitempty"`
	Type string `json:"Type,omitempty"`
}

type EmbyVideoInfo struct {
	Name          string    `json:"Name,omitempty"`
	OriginalTitle string    `json:"OriginalTitle,omitempty"`
	Id            string    `json:"Id,omitempty"`
	DateCreated   time.Time `json:"DateCreated,omitempty"`
	PremiereDate  time.Time `json:"PremiereDate,omitempty"`
	SortName      string    `json:"SortName,omitempty"`
	Path          string    `json:"Path"`
	MediaSources  []struct {
		Protocol             string `json:"Protocol"`
		Id                   string `json:"Id"`
		Path                 string `json:"Path"`
		Type                 string `json:"Type"`
		Container            string `json:"Container"`
		Size                 int    `json:"Size"`
		Name                 string `json:"Name"`
		IsRemote             bool   `json:"IsRemote"`
		RunTimeTicks         int64  `json:"RunTimeTicks"`
		SupportsTranscoding  bool   `json:"SupportsTranscoding"`
		SupportsDirectStream bool   `json:"SupportsDirectStream"`
		SupportsDirectPlay   bool   `json:"SupportsDirectPlay"`
		IsInfiniteStream     bool   `json:"IsInfiniteStream"`
		RequiresOpening      bool   `json:"RequiresOpening"`
		RequiresClosing      bool   `json:"RequiresClosing"`
		RequiresLooping      bool   `json:"RequiresLooping"`
		SupportsProbing      bool   `json:"SupportsProbing"`
		MediaStreams         []struct {
			Codec                  string  `json:"Codec"`
			TimeBase               string  `json:"TimeBase,omitempty"`
			CodecTimeBase          string  `json:"CodecTimeBase,omitempty"`
			VideoRange             string  `json:"VideoRange,omitempty"`
			DisplayTitle           string  `json:"DisplayTitle"`
			NalLengthSize          string  `json:"NalLengthSize,omitempty"`
			IsInterlaced           bool    `json:"IsInterlaced"`
			IsAVC                  bool    `json:"IsAVC,omitempty"`
			BitRate                int     `json:"BitRate,omitempty"`
			BitDepth               int     `json:"BitDepth,omitempty"`
			RefFrames              int     `json:"RefFrames,omitempty"`
			IsDefault              bool    `json:"IsDefault"`
			IsForced               bool    `json:"IsForced"`
			Height                 int     `json:"Height,omitempty"`
			Width                  int     `json:"Width,omitempty"`
			AverageFrameRate       float64 `json:"AverageFrameRate,omitempty"`
			RealFrameRate          float64 `json:"RealFrameRate,omitempty"`
			Profile                string  `json:"Profile,omitempty"`
			Type                   string  `json:"Type"`
			AspectRatio            string  `json:"AspectRatio,omitempty"`
			Index                  int     `json:"Index"`
			IsExternal             bool    `json:"IsExternal"`
			IsTextSubtitleStream   bool    `json:"IsTextSubtitleStream"`
			SupportsExternalStream bool    `json:"SupportsExternalStream"`
			Protocol               string  `json:"Protocol"`
			PixelFormat            string  `json:"PixelFormat,omitempty"`
			Level                  int     `json:"Level,omitempty"`
			IsAnamorphic           bool    `json:"IsAnamorphic,omitempty"`
			Language               string  `json:"Language,omitempty"`
			DisplayLanguage        string  `json:"DisplayLanguage,omitempty"`
			ChannelLayout          string  `json:"ChannelLayout,omitempty"`
			Channels               int     `json:"Channels,omitempty"`
			SampleRate             int     `json:"SampleRate,omitempty"`
			Title                  string  `json:"Title,omitempty"`
			Path                   string  `json:"Path,omitempty"`
		} `json:"MediaStreams"`
		Formats             []interface{} `json:"Formats"`
		Bitrate             int           `json:"Bitrate"`
		RequiredHttpHeaders struct {
		} `json:"RequiredHttpHeaders"`
		ReadAtNativeFramerate      bool `json:"ReadAtNativeFramerate"`
		DefaultAudioStreamIndex    int  `json:"DefaultAudioStreamIndex"`
		DefaultSubtitleStreamIndex int  `json:"DefaultSubtitleStreamIndex"`
	} `json:"MediaSources"`
	MediaStreams []struct {
		Codec                  string `json:"Codec"`
		Language               string `json:"Language"`
		DisplayTitle           string `json:"DisplayTitle"`
		Index                  int    `json:"Index"`
		IsExternal             bool   `json:"IsExternal"`
		IsTextSubtitleStream   bool   `json:"IsTextSubtitleStream"`
		SupportsExternalStream bool   `json:"SupportsExternalStream"`
		Path                   string `json:"Path"`
		Protocol               string `json:"Protocol"`
	} `json:"MediaStreams"`
	ProviderIds struct {
		Tmdb string `json:"Tmdb"`
		Imdb string `json:"Imdb"`
	} `json:"ProviderIds"`
}

type EmbyUsers struct {
	Items []struct {
		Name string `json:"Name"`
		Id   string `json:"Id"`
	} `json:"Items"`
	TotalRecordCount int `json:"TotalRecordCount"`
}

type EmbyVideoInfoByUserId struct {
	Name          string    `json:"Name"`
	OriginalTitle string    `json:"OriginalTitle"`
	Id            string    `json:"Id"`
	DateCreated   time.Time `json:"DateCreated,omitempty"`
	PremiereDate  time.Time `json:"PremiereDate,omitempty"`
	SortName      string    `json:"SortName,omitempty"`
	Path          string    `json:"Path"`
	UserData      struct {
		PlaybackPositionTicks int  `json:"PlaybackPositionTicks"`
		PlayCount             int  `json:"PlayCount"`
		IsFavorite            bool `json:"IsFavorite"`
		Played                bool `json:"Played"`
	} `json:"UserData"`
	MediaSources []struct {
		Path                       string `json:"Path"`
		DefaultAudioStreamIndex    int    `json:"DefaultAudioStreamIndex,omitempty"`
		DefaultSubtitleStreamIndex int    `json:"DefaultSubtitleStreamIndex,omitempty"`
	} `json:"MediaSources"`
}

// GetDefaultSubIndex 获取匹配视频字幕的索引，默认值是0，不应该是0，0 就是没有选择或者说关闭
func (info EmbyVideoInfoByUserId) GetDefaultSubIndex() int {

	for _, mediaSource := range info.MediaSources {
		if info.Path == mediaSource.Path {
			return mediaSource.DefaultSubtitleStreamIndex
		}
	}

	return 0
}

type EmbyMixInfo struct {
	IMDBId                    string // 这个视频的 IMDB ID，注意，连续剧一集是没有 IMDB ID 这个概念的，所以会向上获取到 series 这个级别再取拿 IMDB ID 的
	TMDBId                    string // 这个视频的 TMDb ID
	VideoFolderName           string // 电影就是电影的文件夹名称，连续剧就是对应的剧集的 root 文件夹
	VideoFileName             string // 视频文件名
	PhysicalVideoFileFullPath string // 视频的物理路径（这里指的物理路径是相对于本程序而言，如果是用 docker 使用的话，那么就是映射容器内的路径，如果是用物理机器比如 Windows 使用的话，那么就是相对于物理机器的路径）
	PhysicalRootPath          string // 不是 Emby 扫描的情况，无需关注。视频在那个物理根目录中（这里指的物理路径是相对于本程序而言，如果是用 docker 使用的话，那么就是映射容器内的路径，如果是用物理机器比如 Windows 使用的话，那么就是相对于物理机器的路径）
	PhysicalSeriesRootDir     string // 当前视频的连续剧文件夹根目录
	Ancestors                 []EmbyItemsAncestors
	VideoInfo                 EmbyVideoInfo
}

type UserPlayedItems struct {
	UserName string
	UserID   string
	Items    []EmbyRecentlyItem
}

type Time time.Time

const (
	embyTimeFormart = "2006-01-02T15:04:05"
)

func (t *Time) UnmarshalJSON(data []byte) (err error) {

	orgString := string(data)
	orgString = strings.ReplaceAll(orgString, "\"", "")
	fixTimeString := orgString
	if strings.Contains(orgString, ".") == true {
		strList := strings.Split(orgString, ".")
		if len(strList) > 1 {
			fixTimeString = strList[0]
		}
	}

	now, err := time.ParseInLocation(embyTimeFormart, fixTimeString, time.Local)
	if err != nil {
		return err
	}
	*t = Time(now)
	return
}
func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(embyTimeFormart)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, embyTimeFormart)
	b = append(b, '"')
	return b, nil
}
func (t Time) String() string {
	return time.Time(t).Format(embyTimeFormart)
}
