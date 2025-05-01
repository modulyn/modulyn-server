package db

import (
	"database/sql"
	"fmt"
	"log"
	"modulyn/pkg/models"
	"time"

	"github.com/google/uuid"

	_ "github.com/mattn/go-sqlite3"
)

type Conn interface {
	Close() error
	GetFeatures(sdkKey string) ([]*models.Feature, error)
	CreateFeature(createFeatureRequest *models.CreateFeatureRequest) error
	CreateProject(createProjectRequest *models.CreateProjectRequest) (string, error)
	CreateEnvironment(createEnvironmentRequest *models.CreateEnvironmentRequest) (string, error)
	UpdateFeature(updateFeatureRequest *models.UpdateFeatureRequest) error
	DeleteFeature(featureID string) error
}

type DB struct {
	*sql.DB
}

func InitDB() (Conn, error) {
	db, err := sql.Open("sqlite3", "./modulyn.db")
	if err != nil {
		return nil, err
	}

	// Create the projects table if it doesn't exist
	createProjectTableSQL := `
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
	`
	_, err = db.Exec(createProjectTableSQL)
	if err != nil {
		return nil, err
	}
	log.Println("Created projects table")

	// Create the projects table if it doesn't exist
	createEnvironmentsTableSQL := `
		CREATE TABLE IF NOT EXISTS environments (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			project_id TEXT NOT NULL,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		);
	`
	_, err = db.Exec(createEnvironmentsTableSQL)
	if err != nil {
		return nil, err
	}
	log.Println("Created environments table")

	// Create the flags table if it doesn't exist
	createFeaturesTableSQL := `
		CREATE TABLE IF NOT EXISTS features (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			environment_id TEXT NOT NULL,
			enabled INTEGER NOT NULL,
			json_value blob,
			is_deleted INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (environment_id) REFERENCES environments(id)
		);
	`
	_, err = db.Exec(createFeaturesTableSQL)
	if err != nil {
		return nil, err
	}
	log.Println("Created flags table")

	// create indices for the tables
	createIndicesSQL := `
		BEGIN;
		CREATE INDEX IF NOT EXISTS idx_feature_environment_id ON features (environment_id);
		CREATE INDEX IF NOT EXISTS idx_environment_project_id ON environments (project_id);
		COMMIT;
	`
	_, err = db.Exec(createIndicesSQL)
	if err != nil {
		return nil, err
	}
	log.Println("Created indices for the tables")

	return &DB{
		db,
	}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) GetFeatures(sdkKey string) ([]*models.Feature, error) {
	// Query the database for flags associated with the given SDK key
	query := `
		SELECT f.id, f.name, f.enabled, f.json_value, f.created_at, f.updated_at
		FROM features f
		JOIN environments e ON f.environment_id = e.id
		WHERE e.id = ? AND f.is_deleted = 0`
	rows, err := db.Query(query, sdkKey)
	if err != nil {
		log.Println("Error querying features from database:", err)
		return nil, err
	}
	defer rows.Close()

	features := make([]*models.Feature, 0)

	for rows.Next() {
		var id, name string
		var enabled int
		var jsonValue []byte
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &name, &enabled, &jsonValue, &createdAt, &updatedAt); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		features = append(features, &models.Feature{
			ID:        id,
			Name:      name,
			Enabled:   enabled == 1,
			JsonValue: string(jsonValue),
			CreatedAt: createdAt.Format(time.RFC3339),
			UpdatedAt: updatedAt.Format(time.RFC3339),
		})
	}

	return features, nil
}

func (db *DB) CreateFeature(createFeatureRequest *models.CreateFeatureRequest) error {
	// Insert a new flag into the database
	newID, _ := uuid.NewRandom()
	query := `
		INSERT INTO features (id, name, enabled, json_value, environment_id)
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, newID.String(), createFeatureRequest.Name, createFeatureRequest.Value, createFeatureRequest.JsonValue, createFeatureRequest.EnvironmentID)
	if err != nil {
		log.Println("Error inserting features in database:", err)
		return err
	}
	return nil
}

func (db *DB) UpdateFeature(updateFeatureRequest *models.UpdateFeatureRequest) error {
	// Update an existing flag in the database
	query := `
		UPDATE features
		SET enabled = ?, json_value = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := db.Exec(query, updateFeatureRequest.Value, updateFeatureRequest.JsonValue, updateFeatureRequest.ID)
	if err != nil {
		log.Println("Error updating feature in database:", err)
		return err
	}
	return nil
}

func (db *DB) DeleteFeature(featureID string) error {
	// Soft delete a flag by marking it as deleted
	query := `
		UPDATE features
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := db.Exec(query, featureID)
	if err != nil {
		log.Println("Error deleting feature in database:", err)
		return err
	}
	return nil
}

func (db *DB) CreateProject(createProjectRequest *models.CreateProjectRequest) (string, error) {
	// Insert a new project into the database
	newID, _ := uuid.NewRandom()
	query := `
		INSERT INTO projects (id, name)
		VALUES (?, ?)`
	_, err := db.Exec(query, newID.String(), createProjectRequest.Name)
	if err != nil {
		log.Println("Error inserting project in database:", err)
		return "", err
	}
	return newID.String(), nil
}

func (db *DB) CreateEnvironment(createEnvironmentRequest *models.CreateEnvironmentRequest) (string, error) {
	// Insert a new environment into the database
	newID, _ := uuid.NewRandom()
	sdkKey := fmt.Sprintf("sdk-%s", newID.String())
	query := `
		INSERT INTO environments (id, name, project_id)
		VALUES (?, ?, ?)`
	_, err := db.Exec(query, sdkKey, createEnvironmentRequest.Name, createEnvironmentRequest.ProjectID)
	if err != nil {
		log.Println("Error inserting environment in database:", err)
		return "", err
	}
	return sdkKey, nil
}
