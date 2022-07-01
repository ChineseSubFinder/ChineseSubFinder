package backend

type ReplyVideoList struct {
	MovieInfos  []MovieInfo  `json:"movie_infos"`
	SeasonInfos []SeasonInfo `json:"season_infos"`
}
