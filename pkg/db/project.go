package db

import (
	"context"
	"fmt"
	"log"
	"modulyn/pkg/models"

	"github.com/google/uuid"
)

type ProjectDB interface {
	CreateProject(ctx context.Context, createProjectRequest *models.CreateProjectRequest) (string, error)
	GetProjects(ctx context.Context) ([]*models.Project, error)
	UpdateProject(ctx context.Context, projectID string, updateProjectRequest *models.UpdateProjectRequest) error
	DeleteProject(ctx context.Context, projectID string) error
}

func handleTxCommitOrRollback(tx *LoggerTx, err error) {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Println("Error rolling back transaction:", rollbackErr)
		}
		return
	}
	if commitErr := tx.Commit(); commitErr != nil {
		log.Println("Error committing transaction:", commitErr)
		err = commitErr
	}
}

func (db *DB) CreateProject(ctx context.Context, createProjectRequest *models.CreateProjectRequest) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return "", err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	newID, _ := uuid.NewRandom()
	projectID := newID.String()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO projects 
		(id, name) 
		VALUES 
		(?, ?)
	`, projectID, createProjectRequest.Name)
	if err != nil {
		log.Println("Error inserting project in database:", err)
		return "", err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO environments 
		(id, name, project_id) 
		VALUES 
		(?, ?, ?)
	`, fmt.Sprintf("sdk-%s", projectID), "Default", projectID)
	if err != nil {
		log.Println("Error inserting default environment in database:", err)
		return "", err
	}

	return projectID, nil
}

func (db *DB) GetProjects(ctx context.Context) ([]*models.Project, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	// Query the database for all projects
	query := `
		SELECT id, name
		FROM projects
		WHERE is_deleted = 0
	`
	rows, err := tx.QueryContext(ctx, query, nil)
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

func (db *DB) UpdateProject(ctx context.Context, projectID string, updateProjectRequest *models.UpdateProjectRequest) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	query := `
		UPDATE projects
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, query, updateProjectRequest.Name, projectID)
	if err != nil {
		log.Println("Error updating project in database:", err)
		return err
	}
	return nil
}

func (db *DB) DeleteProject(ctx context.Context, projectID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		handleTxCommitOrRollback(tx, err)
	}()

	getEnvironmentsQuery := `
		SELECT id 
		FROM environments 
		WHERE project_id = ? AND is_deleted = 0
	`
	rows, err := tx.QueryContext(ctx, getEnvironmentsQuery, projectID)
	if err != nil {
		log.Println("Error querying environments from database:", err)
		return err
	}

	var environmentIDs []string
	for rows.Next() {
		var environmentID string
		if err := rows.Scan(&environmentID); err != nil {
			log.Println("Error scanning row:", err)
			return err
		}
		environmentIDs = append(environmentIDs, environmentID)
	}
	rows.Close()

	for _, environmentID := range environmentIDs {
		updateFeatureQuery := `
			UPDATE features
			SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
			WHERE environment_id = ? AND project_id = ?
		`
		_, err := tx.ExecContext(ctx, updateFeatureQuery, environmentID, projectID)
		if err != nil {
			log.Println("Error deleting features in database:", err)
			return err
		}

		updateEnvironmentQuery := `
			UPDATE environments
			SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
			WHERE id = ? AND project_id = ?
		`
		_, err = tx.ExecContext(ctx, updateEnvironmentQuery, environmentID, projectID)
		if err != nil {
			log.Println("Error deleting environment in database:", err)
			return err
		}
	}

	updateProjectQuery := `
		UPDATE projects
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, updateProjectQuery, projectID)
	if err != nil {
		log.Println("Error deleting project in database:", err)
		return err
	}

	return nil
}
