package db

import (
	"context"
	"encoding/json"
	"log"
	"modulyn/pkg/models"
	"time"
)

type FeatureDB interface {
	CreateFeature(ctx context.Context, featureID, projectID string, environments []*models.Environment, createFeatureRequest *models.CreateFeatureRequest) error
	GetFeatures(ctx context.Context, projectID string) ([]*models.Feature, error)
	GetFeaturesByID(ctx context.Context, projectID, featureID string) ([]*models.Feature, error)
	UpdateFeatures(ctx context.Context, projectID, featureID string, updateFeaturesRequest []*models.UpdateFeatureRequest) error
	DeleteFeature(ctx context.Context, projectID, featureID string) error
	GetFeaturesByEnvironmentID(ctx context.Context, environmentID string) ([]*models.Feature, error)
}

func (db *DB) CreateFeature(ctx context.Context, featureID, projectID string, environments []*models.Environment, createFeatureRequest *models.CreateFeatureRequest) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	for _, environment := range environments {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO features 
			(id, name, description, enabled, json_value, environment_id, project_id)
			VALUES 
			(?, ?, ?, ?, ?, ?, ?)
		`, featureID, createFeatureRequest.Name, createFeatureRequest.Description, false, nil, environment.ID, projectID)
		if err != nil {
			log.Println("Error inserting feature in database:", err)
			return err
		}
	}

	return nil
}

func (db *DB) GetFeatures(ctx context.Context, projectID string) ([]*models.Feature, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	// Query the database for flags associated with the given SDK key
	rows, err := tx.QueryContext(ctx, `
		SELECT f.id, f.name, f.description, f.enabled, f.json_value, f.created_at, f.updated_at, f.deleted_at, f.environment_id, e.name, f.project_id, p.name
		FROM features f
		INNER JOIN environments e ON f.environment_id = e.id
		INNER JOIN projects p ON f.project_id = p.id
		WHERE f.project_id = ? AND f.is_deleted = 0
		ORDER BY f.name, e.name
	`, projectID)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return nil, err
	}
	defer rows.Close()

	features := make([]*models.Feature, 0)

	for rows.Next() {
		var id, name, environmentID, projectID, environmentName, projectName string
		var description *string
		var enabled int
		var jsonValue []byte
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time

		if err := rows.Scan(&id, &name, &description, &enabled, &jsonValue, &createdAt, &updatedAt, &deletedAt, &environmentID, &environmentName, &projectID, &projectName); err != nil {
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
		if description != nil {
			feature.Description = *description
		}

		features = append(features, feature)
	}

	return features, nil
}

func (db *DB) GetFeaturesByEnvironmentID(ctx context.Context, environmentID string) ([]*models.Feature, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	// Query the database for flags associated with the given SDK key
	rows, err := tx.QueryContext(ctx, `
		SELECT f.id, f.name, f.description, f.enabled, f.json_value, f.created_at, f.updated_at, f.deleted_at, f.environment_id, e.name, f.project_id, p.name
		FROM features f
		INNER JOIN environments e ON f.environment_id = e.id
		INNER JOIN projects p ON f.project_id = p.id
		WHERE f.environment_id = ? AND f.is_deleted = 0
		ORDER BY f.name, e.name
	`, environmentID)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return nil, err
	}
	defer rows.Close()

	features := make([]*models.Feature, 0)

	for rows.Next() {
		var id, name, environmentID, projectID, environmentName, projectName string
		var description *string
		var enabled int
		var jsonValue []byte
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time

		if err := rows.Scan(&id, &name, &description, &enabled, &jsonValue, &createdAt, &updatedAt, &deletedAt, &environmentID, &environmentName, &projectID, &projectName); err != nil {
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
		if description != nil {
			feature.Description = *description
		}

		features = append(features, feature)
	}

	return features, nil
}

func (db *DB) GetFeaturesByID(ctx context.Context, projectID, featureID string) ([]*models.Feature, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	// Query the database for flags associated with the given SDK key
	rows, err := tx.QueryContext(ctx, `
		SELECT f.id, f.name, f.description, f.enabled, f.json_value, f.created_at, f.updated_at, f.deleted_at, f.environment_id, e.name, f.project_id, p.name
		FROM features f
		INNER JOIN environments e ON f.environment_id = e.id
		INNER JOIN projects p ON f.project_id = p.id
		WHERE f.project_id = ? AND f.id = ? AND f.is_deleted = 0
		ORDER BY f.name, e.name
	`, projectID, featureID)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return nil, err
	}
	defer rows.Close()

	features := make([]*models.Feature, 0)

	for rows.Next() {
		var id, name, environmentID, projectID, environmentName, projectName string
		var description *string
		var enabled int
		var jsonValue []byte
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time

		if err := rows.Scan(&id, &name, &description, &enabled, &jsonValue, &createdAt, &updatedAt, &deletedAt, &environmentID, &environmentName, &projectID, &projectName); err != nil {
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
		if description != nil {
			feature.Description = *description
		}

		features = append(features, feature)
	}

	return features, nil
}

func (db *DB) UpdateFeatures(ctx context.Context, projectID, featureID string, updateFeaturesRequest []*models.UpdateFeatureRequest) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	for _, updateFeatureRequest := range updateFeaturesRequest {
		jsonValueBytes, _ := json.Marshal(updateFeatureRequest.JsonValue)
		_, err = tx.ExecContext(ctx, `
			UPDATE features
			SET enabled = ?, json_value = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND environment_id = ? AND project_id = ?
		`, updateFeatureRequest.Enabled, jsonValueBytes, featureID, updateFeatureRequest.EnvironmentID, projectID)
		if err != nil {
			log.Println("Error updating feature in database:", err)
			return err
		}
	}

	return nil
}

func (db *DB) DeleteFeature(ctx context.Context, projectID, featureID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE features
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND project_id = ?
	`, featureID, projectID)
	if err != nil {
		log.Println("Error deleting feature in database:", err)
		return err
	}
	return nil
}
