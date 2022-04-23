package models

type IMDBLanguage struct {
	Language   string `json:"language" binding:"required"`
	IMDBInfoID uint   `json:"imdb_info_id" binding:"required"`
}
