package emby

import "time"

type T struct {
	Name                  string    `json:"Name"`
	OriginalTitle         string    `json:"OriginalTitle"`
	ServerId              string    `json:"ServerId"`
	Id                    string    `json:"Id"`
	Etag                  string    `json:"Etag"`
	DateCreated           time.Time `json:"DateCreated"`
	CanDelete             bool      `json:"CanDelete"`
	CanDownload           bool      `json:"CanDownload"`
	PresentationUniqueKey string    `json:"PresentationUniqueKey"`
	Container             string    `json:"Container"`
	SortName              string    `json:"SortName"`
	PremiereDate          time.Time `json:"PremiereDate"`
	ExternalUrls          []struct {
		Name string `json:"Name"`
		Url  string `json:"Url"`
	} `json:"ExternalUrls"`
	MediaSources []struct {
		Protocol             string `json:"Protocol"`
		Id                   string `json:"Id"`
		Path                 string `json:"Path"`
		Type                 string `json:"Type"`
		Container            string `json:"Container"`
		Size                 int64  `json:"Size"`
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
			Codec                  string `json:"Codec"`
			TimeBase               string `json:"TimeBase"`
			VideoRange             string `json:"VideoRange,omitempty"`
			DisplayTitle           string `json:"DisplayTitle"`
			NalLengthSize          string `json:"NalLengthSize,omitempty"`
			IsInterlaced           bool   `json:"IsInterlaced"`
			BitRate                int    `json:"BitRate,omitempty"`
			BitDepth               int    `json:"BitDepth,omitempty"`
			RefFrames              int    `json:"RefFrames,omitempty"`
			IsDefault              bool   `json:"IsDefault"`
			IsForced               bool   `json:"IsForced"`
			Height                 int    `json:"Height,omitempty"`
			Width                  int    `json:"Width,omitempty"`
			AverageFrameRate       int    `json:"AverageFrameRate,omitempty"`
			RealFrameRate          int    `json:"RealFrameRate,omitempty"`
			Profile                string `json:"Profile,omitempty"`
			Type                   string `json:"Type"`
			AspectRatio            string `json:"AspectRatio,omitempty"`
			Index                  int    `json:"Index"`
			IsExternal             bool   `json:"IsExternal"`
			IsTextSubtitleStream   bool   `json:"IsTextSubtitleStream"`
			SupportsExternalStream bool   `json:"SupportsExternalStream"`
			Protocol               string `json:"Protocol"`
			PixelFormat            string `json:"PixelFormat,omitempty"`
			Level                  int    `json:"Level,omitempty"`
			IsAnamorphic           bool   `json:"IsAnamorphic,omitempty"`
			CodecTag               string `json:"CodecTag,omitempty"`
			Language               string `json:"Language,omitempty"`
			DisplayLanguage        string `json:"DisplayLanguage,omitempty"`
			ChannelLayout          string `json:"ChannelLayout,omitempty"`
			Channels               int    `json:"Channels,omitempty"`
			SampleRate             int    `json:"SampleRate,omitempty"`
		} `json:"MediaStreams"`
		Formats             []interface{} `json:"Formats"`
		Bitrate             int           `json:"Bitrate"`
		RequiredHttpHeaders struct {
		} `json:"RequiredHttpHeaders"`
		ReadAtNativeFramerate bool `json:"ReadAtNativeFramerate"`
	} `json:"MediaSources"`
	ProductionLocations []string      `json:"ProductionLocations"`
	Path                string        `json:"Path"`
	Overview            string        `json:"Overview"`
	Taglines            []interface{} `json:"Taglines"`
	Genres              []string      `json:"Genres"`
	CommunityRating     float64       `json:"CommunityRating"`
	RunTimeTicks        int64         `json:"RunTimeTicks"`
	ProductionYear      int           `json:"ProductionYear"`
	RemoteTrailers      []interface{} `json:"RemoteTrailers"`
	ProviderIds         struct {
		Tmdb string `json:"Tmdb"`
		Imdb string `json:"Imdb"`
	} `json:"ProviderIds"`
	IsFolder bool   `json:"IsFolder"`
	ParentId string `json:"ParentId"`
	Type     string `json:"Type"`
	People   []struct {
		Name            string `json:"Name"`
		Id              string `json:"Id"`
		Role            string `json:"Role"`
		Type            string `json:"Type"`
		PrimaryImageTag string `json:"PrimaryImageTag,omitempty"`
	} `json:"People"`
	Studios []struct {
		Name string `json:"Name"`
		Id   int    `json:"Id"`
	} `json:"Studios"`
	GenreItems []struct {
		Name string `json:"Name"`
		Id   int    `json:"Id"`
	} `json:"GenreItems"`
	TagItems                []interface{} `json:"TagItems"`
	LocalTrailerCount       int           `json:"LocalTrailerCount"`
	DisplayPreferencesId    string        `json:"DisplayPreferencesId"`
	PrimaryImageAspectRatio float64       `json:"PrimaryImageAspectRatio"`
	MediaStreams            []struct {
		Codec                  string `json:"Codec"`
		TimeBase               string `json:"TimeBase"`
		VideoRange             string `json:"VideoRange,omitempty"`
		DisplayTitle           string `json:"DisplayTitle"`
		NalLengthSize          string `json:"NalLengthSize,omitempty"`
		IsInterlaced           bool   `json:"IsInterlaced"`
		BitRate                int    `json:"BitRate,omitempty"`
		BitDepth               int    `json:"BitDepth,omitempty"`
		RefFrames              int    `json:"RefFrames,omitempty"`
		IsDefault              bool   `json:"IsDefault"`
		IsForced               bool   `json:"IsForced"`
		Height                 int    `json:"Height,omitempty"`
		Width                  int    `json:"Width,omitempty"`
		AverageFrameRate       int    `json:"AverageFrameRate,omitempty"`
		RealFrameRate          int    `json:"RealFrameRate,omitempty"`
		Profile                string `json:"Profile,omitempty"`
		Type                   string `json:"Type"`
		AspectRatio            string `json:"AspectRatio,omitempty"`
		Index                  int    `json:"Index"`
		IsExternal             bool   `json:"IsExternal"`
		IsTextSubtitleStream   bool   `json:"IsTextSubtitleStream"`
		SupportsExternalStream bool   `json:"SupportsExternalStream"`
		Protocol               string `json:"Protocol"`
		PixelFormat            string `json:"PixelFormat,omitempty"`
		Level                  int    `json:"Level,omitempty"`
		IsAnamorphic           bool   `json:"IsAnamorphic,omitempty"`
		CodecTag               string `json:"CodecTag,omitempty"`
		Language               string `json:"Language,omitempty"`
		DisplayLanguage        string `json:"DisplayLanguage,omitempty"`
		ChannelLayout          string `json:"ChannelLayout,omitempty"`
		Channels               int    `json:"Channels,omitempty"`
		SampleRate             int    `json:"SampleRate,omitempty"`
	} `json:"MediaStreams"`
	ImageTags struct {
		Primary string `json:"Primary"`
	} `json:"ImageTags"`
	BackdropImageTags []string `json:"BackdropImageTags"`
	Chapters          []struct {
		StartPositionTicks int64  `json:"StartPositionTicks"`
		Name               string `json:"Name"`
	} `json:"Chapters"`
	MediaType    string        `json:"MediaType"`
	LockedFields []interface{} `json:"LockedFields"`
	LockData     bool          `json:"LockData"`
	Width        int           `json:"Width"`
	Height       int           `json:"Height"`
}
