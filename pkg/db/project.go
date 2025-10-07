package db

import (
	"fmt"
	"log"
	"modulyn/pkg/models"
	"strings"

	"github.com/google/uuid"
)

type ProjectDB interface {
	CreateProject(createProjectRequest *models.CreateProjectRequest) (string, error)
	GetProjects() ([]*models.Project, error)
	UpdateProject(projectID string, updateProjectRequest *models.UpdateProjectRequest) error
	DeleteProject(projectID string) error
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

func (db *DB) UpdateProject(projectID string, updateProjectRequest *models.UpdateProjectRequest) error {
	query := `
		UPDATE projects
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := db.Exec(query, updateProjectRequest.Name, projectID)
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
