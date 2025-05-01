package models

type CreateEnvironmentRequest struct {
	Name      string `json:"name"`
	ProjectID string `json:"projectId"`
}

type UpdateEnvironmentRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Environment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
