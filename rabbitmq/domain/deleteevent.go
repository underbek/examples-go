package domain

type DeleteEvent struct {
	UserId int      `json:"user_id"`
	Urls   []string `json:"urls"`
}
