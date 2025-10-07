package db

import (
	"fmt"
	"log"
	"modulyn/pkg/models"

	"github.com/google/uuid"
)

type EnvironmentDB interface {
	CreateEnvironment(projectID string, createEnvironmentRequest *models.CreateEnvironmentRequest) (string, error)
	GetEnvironments(projectID string) ([]*models.Environment, error)
	GetEnvironment(projectID, environmentID string) (*models.Environment, error)
	UpdateEnvironment(projectID, environmentID string, updateEnvironmentRequest *models.UpdateEnvironmentRequest) error
	DeleteEnvironment(projectID, environmentID string) error
}

func (db *DB) CreateEnvironment(projectID string, createEnvironmentRequest *models.CreateEnvironmentRequest) (string, error) {
	newID, _ := uuid.NewRandom()
	sdkKey := fmt.Sprintf("sdk-%s", newID.String())
	query := `
		INSERT INTO environments (id, name, project_id)
		VALUES (?, ?, ?)`
	_, err := db.Exec(query, sdkKey, createEnvironmentRequest.Name, projectID)
	if err != nil {
		log.Println("Error inserting environment in database:", err)
		return "", err
	}
	return sdkKey, nil
}

func (db *DB) GetEnvironments(projectID string) ([]*models.Environment, error) {
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

func (db *DB) GetEnvironment(projectID, environmentID string) (*models.Environment, error) {
	query := `
		SELECT e.id, e.name
		FROM environments e
		WHERE e.id = ? AND e.project_id = ?
	`
	var id, name string
	row := db.QueryRow(query, environmentID, projectID)
	if err := row.Scan(&id, &name); err != nil {
		if err.Error() == "sql: no rows in result set" {
			log.Println("No rows found")
			return nil, ErrNoRows
		}
		log.Println("Error scanning row:", err)
		return nil, err
	}

	return &models.Environment{
		ID:   id,
		Name: name,
	}, nil
}

func (db *DB) UpdateEnvironment(projectID, environmentID string, updateEnvironmentRequest *models.UpdateEnvironmentRequest) error {
	query := `
		UPDATE environments
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND project_id = ?`
	_, err := db.Exec(query, updateEnvironmentRequest.Name, environmentID, projectID)
	if err != nil {
		log.Println("Error updating environment in database:", err)
		return err
	}
	return nil
}

func (db *DB) DeleteEnvironment(projectID, environmentID string) error {
	query := `
		UPDATE environments
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND project_id = ?`
	_, err := db.Exec(query, environmentID, projectID)
	if err != nil {
		log.Println("Error deleting environment in database:", err)
		return err
	}
	return nil
}
