package models

type CreateEnvironmentRequest struct {
	Name      string `json:"name"`
	ProjectID string `json:"projectId"`
}
