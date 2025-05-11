package models

type Feature struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	JsonValue JsonValue `json:"jsonValue"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type CreateFeatureRequest struct {
	Name string `json:"name"`
}

type UpdateFeatureRequest struct {
	Enabled   bool      `json:"enabled"`
	JsonValue JsonValue `json:"jsonValue,omitempty"`
}

type JsonValue struct {
	Key     string   `json:"key"`
	Values  []string `json:"values"`
	Enabled bool     `json:"enabled"`
}
