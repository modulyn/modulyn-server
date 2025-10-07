package db

import (
	"context"
	"fmt"
	"log"
	"modulyn/pkg/models"

	"github.com/google/uuid"
)

type EnvironmentDB interface {
	CreateEnvironment(ctx context.Context, projectID string, createEnvironmentRequest *models.CreateEnvironmentRequest) (string, error)
	GetEnvironments(ctx context.Context, projectID string) ([]*models.Environment, error)
	GetEnvironment(ctx context.Context, projectID, environmentID string) (*models.Environment, error)
	UpdateEnvironment(ctx context.Context, projectID, environmentID string, updateEnvironmentRequest *models.UpdateEnvironmentRequest) error
	DeleteEnvironment(ctx context.Context, projectID, environmentID string) error
}

func (db *DB) CreateEnvironment(ctx context.Context, projectID string, createEnvironmentRequest *models.CreateEnvironmentRequest) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return "", err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	newEnvironmentId, _ := uuid.NewRandom()
	sdkKey := fmt.Sprintf("sdk-%s", newEnvironmentId.String())

	rows, err := tx.QueryContext(ctx, `
		SELECT distinct f.id, f.name 
		FROM features f 
		WHERE f.project_id = ? AND f.is_deleted = 0
	`, projectID)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return "", err
	}
	defer rows.Close()

	var features []struct {
		id   string
		name string
	}
	for rows.Next() {
		var feature struct {
			id   string
			name string
		}
		if err := rows.Scan(&feature.id, &feature.name); err != nil {
			log.Println("Error scanning row:", err)
			return "", err
		}
		features = append(features, feature)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO environments 
		(id, name, project_id) 
		VALUES 
		(?, ?, ?)
	`, sdkKey, createEnvironmentRequest.Name, projectID)
	if err != nil {
		log.Println("Error inserting environment:", err)
		return "", err
	}

	for _, feature := range features {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO features 
			(id, name, enabled, json_value, environment_id, project_id) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, feature.id, feature.name, false, nil, sdkKey, projectID)
		if err != nil {
			log.Println("Error inserting feature for new environment:", err)
			return "", err
		}
	}

	return sdkKey, nil
}

func (db *DB) GetEnvironments(ctx context.Context, projectID string) ([]*models.Environment, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	rows, err := db.QueryContext(ctx, `
		SELECT id, name 
		FROM environments 
		WHERE project_id = ? and is_deleted = 0
	`, projectID)
	if err != nil {
		log.Println("Error querying environments from database:", err)
		return nil, err
	}
	defer rows.Close()

	environments := make([]*models.Environment, 0)

	for rows.Next() {
		var id, name string

		if err := rows.Scan(&id, &name); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		environments = append(environments, &models.Environment{
			ID:   id,
			Name: name,
		})
	}

	return environments, nil
}

func (db *DB) GetEnvironment(ctx context.Context, projectID, environmentID string) (*models.Environment, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	var id, name string
	rows, err := tx.QueryContext(ctx, `
		SELECT e.id, e.name 
		FROM environments e 
		WHERE e.id = ? AND e.project_id = ?
	`, environmentID, projectID)
	if err != nil {
		log.Println("Error querying environment from database:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			if err.Error() == "sql: no rows in result set" {
				log.Println("No rows found")
				return nil, ErrNoRows
			}
			log.Println("Error scanning row:", err)
			return nil, err
		}
	}

	return &models.Environment{
		ID:   id,
		Name: name,
	}, nil
}

func (db *DB) UpdateEnvironment(ctx context.Context, projectID, environmentID string, updateEnvironmentRequest *models.UpdateEnvironmentRequest) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE environments 
		SET name = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ? AND project_id = ?
	`, updateEnvironmentRequest.Name, environmentID, projectID)
	if err != nil {
		log.Println("Error updating environment in database:", err)
		return err
	}
	return nil
}

func (db *DB) DeleteEnvironment(ctx context.Context, projectID, environmentID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM features 
		WHERE environment_id = ? AND project_id = ? AND is_deleted = 0
	`, environmentID, projectID)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return err
	}
	defer rows.Close()

	var featureIDs []string
	for rows.Next() {
		var featureID string
		if err := rows.Scan(&featureID); err != nil {
			log.Println("Error scanning row:", err)
			return err
		}
		featureIDs = append(featureIDs, featureID)
	}

	for _, featureID := range featureIDs {
		_, err := tx.ExecContext(ctx, `
			UPDATE features
			SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
			WHERE feature_id = ? AND environment_id = ? AND project_id = ?
		`, featureID, environmentID, projectID)
		if err != nil {
			log.Println("Error deleting features in database:", err)
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE environments
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND project_id = ?
	`, environmentID, projectID)
	if err != nil {
		log.Println("Error deleting environment in database:", err)
		return err
	}

	return nil
}
