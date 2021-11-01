package media_system

type RecentlyItems struct {
	Items            []RecentlyItem `json:"Items,omitempty"`
	TotalRecordCount int            `json:"TotalRecordCount,omitempty"`
}

type RecentlyItem struct {
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
