package models

type Feature struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	JsonValue string `json:"jsonValue"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type CreateFeatureRequest struct {
	Name string `json:"name"`
}

type UpdateFeatureRequest struct {
	Enabled   bool    `json:"enabled"`
	JsonValue *string `json:"jsonValue,omitempty"`
}
