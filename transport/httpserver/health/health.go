package health

// HealthCheckPath defines default health check path to services
const HealthCheckPath = "/health_check"

// HealthResponse service standard health check response.
type HealthResponse struct {
	Status     string `json:"status,omitempty" `
	CommitHash string `json:"commit_hash" validation:"required"`
}
