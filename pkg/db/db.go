package db

import (
	"database/sql"
	"fmt"
	"log"
	"modulyn/pkg/models"
	"strings"
	"time"

	"github.com/google/uuid"

	_ "github.com/mattn/go-sqlite3"
)

type Conn interface {
	Close() error
	CreateFeature(createFeatureRequest *models.CreateFeatureRequest) error
	GetFeatures(sdkKey string) ([]*models.Feature, error)
	UpdateFeature(updateFeatureRequest *models.UpdateFeatureRequest) error
	DeleteFeature(featureID string) error
	CreateProject(createProjectRequest *models.CreateProjectRequest) (string, error)
	GetProjects() ([]*models.Project, error)
	UpdateProject(updateProjectRequest *models.UpdateProjectRequest) error
	DeleteProject(projectID string) error
	CreateEnvironment(createEnvironmentRequest *models.CreateEnvironmentRequest) (string, error)
	GetEnvironments(projectID string) ([]*models.Environment, error)
	UpdateEnvironment(updateEnvironmentRequest *models.UpdateEnvironmentRequest) error
	DeleteEnvironment(environmentID string) error
}

type DB struct {
	*sql.DB
}

func (db *DB) Close() error {
	return db.DB.Close()
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
			name TEXT NOT NULL,
			is_deleted INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
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
			is_deleted INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
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

func (db *DB) GetFeatures(sdkKey string) ([]*models.Feature, error) {
	// Query the database for flags associated with the given SDK key
	query := `
		SELECT f.id, f.name, f.enabled, f.json_value, f.created_at, f.updated_at
		FROM features f
		JOIN environments e ON f.environment_id = e.id
		WHERE e.id = ? AND f.is_deleted = 0
		ORDER BY f.updated_at DESC`
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
	// Insert a new feature into the database
	query := `
		SELECT id
		FROM environments
		WHERE project_id = ? AND is_deleted = 0
	`

	rows, err := db.Query(query, createFeatureRequest.ProjectID)
	if err != nil {
		log.Println("Error querying environments from database:", err)
		return err
	}
	defer rows.Close()

	environments := make([]string, 0)

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Println("Error scanning row:", err)
			return err
		}

		environments = append(environments, id)
	}

	parameters := make([]any, 0)
	createQuery := strings.Builder{}
	createQuery.WriteString(`BEGIN;`)
	for _, environmentID := range environments {
		newID, _ := uuid.NewRandom()
		createQuery.WriteString(`
			INSERT INTO features (id, name, enabled, json_value, environment_id)
			VALUES (?, ?, ?, ?, ?);
		`)
		parameters = append(parameters, newID.String(), createFeatureRequest.Name, createFeatureRequest.Value, createFeatureRequest.JsonValue, environmentID)
	}
	createQuery.WriteString(`COMMIT;`)

	_, err = db.Exec(createQuery.String(), parameters...)
	if err != nil {
		log.Println("Error inserting feature in database:", err)
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
	query := strings.Builder{}
	parameters := make([]any, 0)
	query.WriteString(`BEGIN;`)
	newID, _ := uuid.NewRandom()
	projectID := newID.String()
	query.WriteString(`
		INSERT INTO projects (id, name)
		VALUES (?, ?);`)
	parameters = append(parameters, projectID, createProjectRequest.Name)
	query.WriteString(`
		INSERT INTO environments (id, name, project_id)
		VALUES (?, ?, ?);
	`)
	parameters = append(parameters, fmt.Sprintf("sdk-%s", projectID), "Default", projectID)
	query.WriteString(`COMMIT;`)
	_, err := db.Exec(query.String(), parameters...)
	if err != nil {
		log.Println("Error inserting project in database:", err)
		return "", err
	}
	return projectID, nil
}

func (db *DB) GetProjects() ([]*models.Project, error) {
	// Query the database for all projects
	query := `
		SELECT id, name
		FROM projects
		WHERE is_deleted = 0`
	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error querying projects from database:", err)
		return nil, err
	}
	defer rows.Close()

	projects := make([]*models.Project, 0)

	for rows.Next() {
		var id, name string

		if err := rows.Scan(&id, &name); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		projects = append(projects, &models.Project{
			ID:   id,
			Name: name,
		})
	}
	return projects, nil
}

func (db *DB) UpdateProject(updateProjectRequest *models.UpdateProjectRequest) error {
	// Update an existing project in the database
	query := `
		UPDATE projects
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := db.Exec(query, updateProjectRequest.Name, updateProjectRequest.ID)
	if err != nil {
		log.Println("Error updating project in database:", err)
		return err
	}
	return nil
}

func (db *DB) DeleteProject(projectID string) error {
	// Soft delete a project by marking it as deleted
	query := `
		UPDATE projects
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := db.Exec(query, projectID)
	if err != nil {
		log.Println("Error deleting project in database:", err)
		return err
	}
	return nil
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

func (db *DB) GetEnvironments(projectID string) ([]*models.Environment, error) {
	// Query the database for environments associated with the given project ID
	query := `
		SELECT id, name
		FROM environments
		WHERE project_id = ? and is_deleted = 0`
	rows, err := db.Query(query, projectID)
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

func (db *DB) UpdateEnvironment(updateEnvironmentRequest *models.UpdateEnvironmentRequest) error {
	// Update an existing environment in the database
	query := `
		UPDATE environments
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := db.Exec(query, updateEnvironmentRequest.Name, updateEnvironmentRequest.ID)
	if err != nil {
		log.Println("Error updating environment in database:", err)
		return err
	}
	return nil
}

func (db *DB) DeleteEnvironment(environmentID string) error {
	// Soft delete an environment by marking it as deleted
	query := `
		UPDATE environments
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := db.Exec(query, environmentID)
	if err != nil {
		log.Println("Error deleting environment in database:", err)
		return err
	}
	return nil
}
