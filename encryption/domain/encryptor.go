package domain

/*
ENUM(
CARD
CVV
REQUISITE
SECRET
)
*/
type EncryptorType int

type EncryptorData struct {
	ID            int64         `db:"id"`
	Engine        string        `json:"engine" db:"engine"`
	EncryptorType EncryptorType `json:"encryptor_type" db:"encryptor_type"`
	Additional    Attributes    `json:"additional,omitempty" db:"additional"`
}
