package models

type Feature struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Enabled         bool      `json:"enabled"`
	JsonValue       JsonValue `json:"jsonValue"`
	CreatedAt       string    `json:"createdAt"`
	UpdatedAt       string    `json:"updatedAt"`
	DeletedAt       string    `json:"deletedAt"`
	EnvironmentID   string    `json:"environmentId"`
	EnvironmentName string    `json:"environmentName"`
	ProjectID       string    `json:"projectId"`
	ProjectName     string    `json:"projectName"`
}

type CreateFeatureRequest struct {
	Name string `json:"name"`
}

type UpdateFeatureRequest struct {
	EnvironmentID string    `json:"environmentId"`
	Enabled       bool      `json:"enabled"`
	JsonValue     JsonValue `json:"jsonValue,omitempty"`
}

type JsonValue struct {
	Key     string   `json:"key"`
	Values  []string `json:"values"`
	Enabled bool     `json:"enabled"`
}
