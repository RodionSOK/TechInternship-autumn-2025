package domain

type UserStatistics struct {
	UserId       string `json:"user_id"`
	Username     string `json:"username"`
	TotalReviews int    `json:"total_reviews"`
	OpenReviews  int    `json:"open_reviews"`
	MergedReviews int   `json:"merged_reviews"`
}

type Statistics struct {
	TotalPRs      int               `json:"total_prs"`
	OpenPRs       int               `json:"open_prs"`
	MergedPRs     int               `json:"merged_prs"`
	UserStatistics []*UserStatistics `json:"user_statistics"`
}

