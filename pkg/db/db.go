package db

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNoRows = errors.New("no results found")
)

var EnableSqlLogging = false

type contextKey string

const CorrelationKey contextKey = "correlation_id"

type Conn interface {
	Close() error
	FeatureDB
	ProjectDB
	EnvironmentDB
}

type DB struct {
	*sql.DB
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func InitDB(enableSqlLogging bool) (Conn, error) {
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
			id TEXT,
			name TEXT NOT NULL,
			label TEXT NOT NULL,
			description TEXT,
			environment_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			enabled INTEGER NOT NULL,
			json_value blob,
			is_deleted INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			PRIMARY KEY (id, environment_id, project_id),
			FOREIGN KEY (environment_id) REFERENCES environments(id),
			FOREIGN KEY (project_id) REFERENCES projects(id)
		);
	`
	_, err = db.Exec(createFeaturesTableSQL)
	if err != nil {
		return nil, err
	}
	log.Println("Created features table")

	// create indices for the tables
	createIndicesSQL := `
		BEGIN;
		CREATE INDEX IF NOT EXISTS idx_feature_project_id_environment_id ON features (project_id, environment_id);
		CREATE INDEX IF NOT EXISTS idx_feature_updated_at ON features (updated_at);
		CREATE INDEX IF NOT EXISTS idx_environment_project_id ON environments (project_id);
		COMMIT;
	`
	_, err = db.Exec(createIndicesSQL)
	if err != nil {
		return nil, err
	}
	log.Println("Created indices for the tables")

	EnableSqlLogging = enableSqlLogging

	return &DB{
		db,
	}, nil
}
