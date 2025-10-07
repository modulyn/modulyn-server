package db

import (
	"encoding/json"
	"log"
	"modulyn/pkg/models"
	"time"
)

type FeatureDB interface {
	CreateFeature(featureID, projectID, environmentID string, createFeatureRequest *models.CreateFeatureRequest) error
	GetFeatures(projectID, environmentID string) ([]*models.Feature, error)
	GetFeaturesByEnvironmentID(environmentID string) ([]*models.Feature, error)
	GetFeaturesByProjectID(projectID string) ([]*models.Feature, error)
	UpdateFeature(projectID, environmentID, featureID string, updateFeatureRequest *models.UpdateFeatureRequest) error
	DeleteFeature(projectID, environmentID, featureID string) error
	GetFeature(projectID, environmentID, featureID string) (*models.Feature, error)
}

func (db *DB) GetFeatures(projectID, environmentID string) ([]*models.Feature, error) {
	// Query the database for flags associated with the given SDK key
	query := `
		SELECT f.id, f.name, f.enabled, f.json_value, f.created_at, f.updated_at, f.deleted_at, f.environment_id, e.name, f.project_id, p.name
		FROM features f
		INNER JOIN environments e ON f.environment_id = e.id
		INNER JOIN projects p ON f.project_id = p.id
		WHERE f.environment_id = ? AND f.project_id = ? AND f.is_deleted = 0
		ORDER BY f.updated_at DESC`
	rows, err := db.Query(query, environmentID, projectID)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return nil, err
	}
	defer rows.Close()

	features := make([]*models.Feature, 0)

	for rows.Next() {
		var id, name, environmentID, projectID, environmentName, projectName string
		var enabled int
		var jsonValue []byte
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time

		if err := rows.Scan(&id, &name, &enabled, &jsonValue, &createdAt, &updatedAt, &deletedAt, &environmentID, &environmentName, &projectID, &projectName); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		var jsonVal models.JsonValue
		json.Unmarshal(jsonValue, &jsonVal)

		feature := &models.Feature{
			ID:              id,
			Name:            name,
			Enabled:         enabled == 1,
			JsonValue:       jsonVal,
			CreatedAt:       createdAt.Format(time.RFC3339),
			UpdatedAt:       updatedAt.Format(time.RFC3339),
			EnvironmentID:   environmentID,
			EnvironmentName: environmentName,
			ProjectID:       projectID,
			ProjectName:     projectName,
		}
		if deletedAt != nil {
			feature.DeletedAt = deletedAt.Format(time.RFC3339)
		}

		features = append(features, feature)
	}

	return features, nil
}

func (db *DB) GetFeaturesByEnvironmentID(environmentID string) ([]*models.Feature, error) {
	// Query the database for flags associated with the given SDK key
	query := `
		SELECT f.id, f.name, f.enabled, f.json_value, f.created_at, f.updated_at, f.deleted_at, f.environment_id, e.name, f.project_id, p.name
		FROM features f
		INNER JOIN environments e ON f.environment_id = e.id
		INNER JOIN projects p ON f.project_id = p.id
		WHERE f.environment_id = ? AND f.is_deleted = 0
		ORDER BY f.updated_at DESC`
	rows, err := db.Query(query, environmentID)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return nil, err
	}
	defer rows.Close()

	features := make([]*models.Feature, 0)

	for rows.Next() {
		var id, name, environmentID, projectID, environmentName, projectName string
		var enabled int
		var jsonValue []byte
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time

		if err := rows.Scan(&id, &name, &enabled, &jsonValue, &createdAt, &updatedAt, &deletedAt, &environmentID, &environmentName, &projectID, &projectName); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		var jsonVal models.JsonValue
		json.Unmarshal(jsonValue, &jsonVal)

		feature := &models.Feature{
			ID:              id,
			Name:            name,
			Enabled:         enabled == 1,
			JsonValue:       jsonVal,
			CreatedAt:       createdAt.Format(time.RFC3339),
			UpdatedAt:       updatedAt.Format(time.RFC3339),
			EnvironmentID:   environmentID,
			EnvironmentName: environmentName,
			ProjectID:       projectID,
			ProjectName:     projectName,
		}
		if deletedAt != nil {
			feature.DeletedAt = deletedAt.Format(time.RFC3339)
		}

		features = append(features, feature)
	}

	return features, nil
}

func (db *DB) GetFeaturesByProjectID(projectID string) ([]*models.Feature, error) {
	// Query the database for flags associated with the given SDK key
	query := `
		SELECT f.id, f.name, f.enabled, f.json_value, f.created_at, f.updated_at, f.deleted_at, f.environment_id, e.name, f.project_id, p.name
		FROM features f
		INNER JOIN environments e ON f.environment_id = e.id
		INNER JOIN projects p ON f.project_id = p.id
		WHERE f.project_id = ? AND f.is_deleted = 0
		ORDER BY f.updated_at DESC`
	rows, err := db.Query(query, projectID)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return nil, err
	}
	defer rows.Close()

	features := make([]*models.Feature, 0)

	for rows.Next() {
		var id, name, environmentID, projectID, environmentName, projectName string
		var enabled int
		var jsonValue []byte
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time

		if err := rows.Scan(&id, &name, &enabled, &jsonValue, &createdAt, &updatedAt, &deletedAt, &environmentID, &environmentName, &projectID, &projectName); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		var jsonVal models.JsonValue
		json.Unmarshal(jsonValue, &jsonVal)

		feature := &models.Feature{
			ID:              id,
			Name:            name,
			Enabled:         enabled == 1,
			JsonValue:       jsonVal,
			CreatedAt:       createdAt.Format(time.RFC3339),
			UpdatedAt:       updatedAt.Format(time.RFC3339),
			EnvironmentID:   environmentID,
			EnvironmentName: environmentName,
			ProjectID:       projectID,
			ProjectName:     projectName,
		}
		if deletedAt != nil {
			feature.DeletedAt = deletedAt.Format(time.RFC3339)
		}

		features = append(features, feature)
	}

	return features, nil
}

func (db *DB) CreateFeature(featureID, projectID, environmentID string, createFeatureRequest *models.CreateFeatureRequest) error {
	query := `
		INSERT INTO features (id, name, enabled, json_value, environment_id, project_id)
		VALUES (?, ?, ?, ?, ?, ?);
	`

	_, err := db.Exec(query, featureID, createFeatureRequest.Name, false, nil, environmentID, projectID)
	if err != nil {
		log.Println("Error inserting feature in database:", err)
		return err
	}

	return nil
}

func (db *DB) UpdateFeature(projectID, environmentID, featureID string, updateFeatureRequest *models.UpdateFeatureRequest) error {
	// Update an existing flag in the database
	jsonValueBytes, _ := json.Marshal(updateFeatureRequest.JsonValue)
	query := `
		UPDATE features
		SET enabled = ?, json_value = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND environment_id = ? AND project_id = ?`
	_, err := db.Exec(query, updateFeatureRequest.Enabled, jsonValueBytes, featureID, environmentID, projectID)
	if err != nil {
		log.Println("Error updating feature in database:", err)
		return err
	}
	return nil
}

func (db *DB) DeleteFeature(projectID, environmentID, featureID string) error {
	// Soft delete a flag by marking it as deleted
	query := `
		UPDATE features
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND environment_id = ? AND project_id = ?`
	_, err := db.Exec(query, featureID, environmentID, projectID)
	if err != nil {
		log.Println("Error deleting feature in database:", err)
		return err
	}
	return nil
}

func (db *DB) GetFeature(projectID, environmentID, featureID string) (*models.Feature, error) {
	query := `
		SELECT f.id, f.name, f.enabled, f.json_value, f.created_at, f.updated_at, f.deleted_at, f.environment_id, e.name, f.project_id, p.name
		FROM features f
		INNER JOIN environments e ON f.environment_id = e.id
		INNER JOIN projects p ON f.project_id = p.id
		WHERE f.id = ? AND f.environment_id = ? AND f.project_id = ?
	`
	var id, name, environmentId, projectId, environmentName, projectName string
	var enabled int
	var jsonValue []byte
	var createdAt, updatedAt time.Time
	var deletedAt *time.Time

	row := db.QueryRow(query, featureID, environmentID, projectID)
	if err := row.Scan(&id, &name, &enabled, &jsonValue, &createdAt, &updatedAt, &deletedAt, &environmentId, &environmentName, &projectId, &projectName); err != nil {
		if err.Error() == "sql: no rows in result set" {
			log.Println("No rows found")
			return nil, ErrNoRows
		}
		log.Println("Error scanning row:", err)
		return nil, err
	}

	var jsonVal models.JsonValue
	json.Unmarshal(jsonValue, &jsonVal)

	feature := &models.Feature{
		ID:              id,
		Name:            name,
		Enabled:         enabled == 1,
		JsonValue:       jsonVal,
		CreatedAt:       createdAt.Format(time.RFC3339),
		UpdatedAt:       updatedAt.Format(time.RFC3339),
		EnvironmentID:   environmentID,
		EnvironmentName: environmentName,
		ProjectID:       projectID,
		ProjectName:     projectName,
	}
	if deletedAt != nil {
		feature.DeletedAt = deletedAt.Format(time.RFC3339)
	}

	return feature, nil
}
