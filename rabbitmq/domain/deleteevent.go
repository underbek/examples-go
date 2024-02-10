package domain

type DeleteEvent struct {
	UserID int      `json:"user_id"`
	Urls   []string `json:"urls"`
}
