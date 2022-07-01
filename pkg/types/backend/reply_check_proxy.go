package backend

type ReplyCheckStatus struct {
	SubSiteStatus []SiteStatus `json:"sub_site_status"`
}

type SiteStatus struct {
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
	Speed int64  `json:"speed"`
}
