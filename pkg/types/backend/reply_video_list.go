package backend

type ReplyVideoList struct {
	MovieInfos  []MovieInfo  `json:"movie_infos"`
	SeasonInfos []SeasonInfo `json:"season_infos"`
}

type ReplyMainList struct {
	MovieInfos  []MovieInfoV2  `json:"movie_infos_v2"`
	SeasonInfos []SeasonInfoV2 `json:"season_infos_v2"`
}
