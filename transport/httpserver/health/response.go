package health

// ServiceError service standard error.
type ServiceError struct {
	Message string `json:"message" example:"some error message"`
}
