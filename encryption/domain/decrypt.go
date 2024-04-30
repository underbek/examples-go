package domain

type DecryptRequest struct {
	EncryptedValue string `json:"encrypted_value"`
	EncryptorID    string `json:"encryptor_id"`
}

type DecryptResponse struct {
	Value string        `json:"value,omitempty"`
	Type  EncryptorType `json:"type,omitempty"`
}
