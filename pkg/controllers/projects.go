package controllers

import (
	"encoding/json"
	"log"
	"modulyn/pkg/models"
	"net/http"
)

func (c *controller) ProjectsController(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		projects, err := c.conn.GetProjects(r.Context())
		if err != nil {
			log.Println("Error getting projects:", err)
			http.Error(w, "Failed to get projects", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: projects,
		})
	case http.MethodPost:
		var createProjectRequest models.CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&createProjectRequest); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		projectID, err := c.conn.CreateProject(r.Context(), &createProjectRequest)
		if err != nil {
			log.Println("Error creating project:", err)
			http.Error(w, "Failed to create project", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.Response{
			Data: projectID,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *controller) ProjectByIdControllers(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPut:
		projectID := r.PathValue("projectId")
		var updateProjectRequest models.UpdateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&updateProjectRequest); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := c.conn.UpdateProject(r.Context(), projectID, &updateProjectRequest); err != nil {
			log.Println("Error updating project:", err)
			http.Error(w, "Failed to update project", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		projectID := r.PathValue("projectId")

		if err := c.conn.DeleteProject(r.Context(), projectID); err != nil {
			log.Println("Error deleting project:", err)
			http.Error(w, "Failed to delete project", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
