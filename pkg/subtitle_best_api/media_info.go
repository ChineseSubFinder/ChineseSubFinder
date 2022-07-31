package subtitle_best_api

/*
	{
		"id": "tt7278862",
		"source": "imdb",
		"video_type": "series"
	}

	{
		"id": "503235",
		"source": "tmdb",
		"video_type": "movie"
	}
*/
type MediaInfoReq struct {
	Id        string `json:"id"`
	Source    string `json:"source"`     // options=imdb|tmdb
	VideoType string `json:"video_type"` // ,options=movie|series
}

/*
	{
		"status": 1,
		"message": "",
		"tmdb_id": "503235",
		"original_title": "邪不压正",
		"original_language": "zh",
		"title_en": "Hidden Man",
		"title_cn": "邪不压正",
		"year": "2018-07-13"
	}

	{
		"status": 1,
		"message": "",
		"tmdb_id": "78154",
		"original_title": "L'amica geniale",
		"original_language": "it",
		"title_en": "My Brilliant Friend",
		"title_cn": "我的天才女友",
		"year": "2018-11-18"
	}
*/
type MediaInfoReply struct {
	Status           int    `json:"status"` // 0 失败，1 成功，2 在队列中等待查询
	Message          string `json:"message"`
	TMDBId           string `json:"tmdb_id,omitempty"`
	OriginalTitle    string `json:"original_title,omitempty"`
	OriginalLanguage string `json:"original_language,omitempty"`
	TitleEN          string `json:"title_en,omitempty"`
	TitleCN          string `json:"title_cn,omitempty"`
	Year             string `json:"year,omitempty"`
}

type IdConvertReq struct {
	Id        string `json:"id"`
	Source    string `json:"source"`     // options=imdb|tmdb
	VideoType string `json:"video_type"` // ,options=movie|series
}

type IdConvertReply struct {
	Status  int    `json:"status"` // 0 失败，1 成功
	Message string `json:"message"`
	IMDBId  string `json:"imdb_id,omitempty"`
	TVDBId  string `json:"tvdb_id,omitempty"`
}

type FeedReq struct {
	Id           string `json:"id"`             // 当前用户的id，这个需要在用户缓存中随机生成
	Version      string `json:"version"`        // 当前版本号
	MediaServer  string `json:"media_server"`   // 媒体服务的名称，没有使用则是 None
	EnableShare  bool   `json:"enable_share"`   // 是否开启了共享功能
	EnableApiKey bool   `json:"enable_api_key"` // 是否开启本地 http api 功能
}

type FeedReply struct {
	Status  int    `json:"status"` // 0 失败，1 成功
	Message string `json:"message"`
}
