package domain

type EncryptRequest struct {
	Value string        `json:"value"`
	Type  EncryptorType `json:"type"`
}

type EncryptResponse struct {
	EncryptedValue string `json:"encrypted_value"`
	EncryptorID    string `json:"encryptor_id"`
}
