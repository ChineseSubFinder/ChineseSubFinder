package emby

import (
	"strings"
	"time"
)

type EmbyRecentlyItems struct {
	Items []struct {
		Name                string   `json:"Name,omitempty"`
		Id                  string   `json:"Id,omitempty"`
		IndexNumber         int      `json:"IndexNumber,omitempty"`
		ParentIndexNumber   int      `json:"ParentIndexNumber,omitempty"`
		Type        		string   `json:"Type,omitempty"`
		SeriesName          string   `json:"SeriesName,omitempty"`
	} `json:"Items,omitempty"`
	TotalRecordCount int `json:"TotalRecordCount,omitempty"`
}

type EmbyItemsAncestors struct {
	Name                    string `json:"Name,omitempty"`
	Path                    string `json:"Path,omitempty"`
	Type                    string `json:"Type,omitempty"`
}

type EmbyVideoInfo struct {
	Name                         string    `json:"Name,omitempty"`
	OriginalTitle                string    `json:"OriginalTitle,omitempty"`
	Id                           string    `json:"Id,omitempty"`
	DateCreated                  time.Time `json:"DateCreated,omitempty"`
	PremiereDate				 time.Time `json:"PremiereDate,omitempty"`
	SortName                     string    `json:"SortName,omitempty"`
	Path                		 string   `json:"Path"`
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
}

type EmbyMixInfo struct {
	VideoFolderName       string			// 电影就是电影的文件夹名称，连续剧就是对应的剧集的 root 文件夹
	VideoFileName         string			// 视频文件名
	VideoFileRelativePath string	// 视频文件的相对路径（注意，这里还是需要补齐 x:/电影 这样的 root 路径的，仅仅算相对路径）
	VideoFileFullPath     string
	Ancestors             []EmbyItemsAncestors
	VideoInfo             EmbyVideoInfo
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
