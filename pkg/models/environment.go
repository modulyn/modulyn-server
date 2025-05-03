package models

type CreateEnvironmentRequest struct {
	Name string `json:"name"`
}

type UpdateEnvironmentRequest struct {
	Name string `json:"name"`
}

type Environment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
