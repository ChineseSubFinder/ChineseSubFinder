package models

type IMDBAKA struct {
	AKA        string `json:"aka" binding:"required"`
	IMDBInfoID uint   `json:"imdb_info_id" binding:"required"`
}
