package domain

type Context struct {
	ID   uint64     `json:"id" db:"id"`
	Meta Attributes `json:"meta" db:"meta"`
}
