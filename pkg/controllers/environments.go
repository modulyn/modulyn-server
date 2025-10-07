package controllers

import (
	"encoding/json"
	"log"
	"modulyn/pkg/models"
	"net/http"
)

func (c *controller) EnvironmentsController(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		projectID := r.PathValue("projectId")

		environments, err := c.conn.GetEnvironments(r.Context(), projectID)
		if err != nil {
			http.Error(w, "Failed to get environments", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: environments,
		})
	case http.MethodPost:
		projectID := r.PathValue("projectId")
		var createEnvironmentRequest models.CreateEnvironmentRequest
		if err := json.NewDecoder(r.Body).Decode(&createEnvironmentRequest); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		environmentID, err := c.conn.CreateEnvironment(r.Context(), projectID, &createEnvironmentRequest)
		if err != nil {
			log.Println("Error creating environment:", err)
			http.Error(w, "Failed to create environment", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.Response{
			Data: environmentID,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *controller) EnvironmentByIdControllers(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		projectID := r.PathValue("projectId")
		environmentID := r.PathValue("environmentId")
		environment, err := c.conn.GetEnvironment(r.Context(), projectID, environmentID)
		if err != nil {
			log.Println("Error fetching environment:", err)
			http.Error(w, "Failed to get environment", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: environment,
		})
	case http.MethodPut:
		projectID := r.PathValue("projectId")
		environmentID := r.PathValue("environmentId")
		var updateEnvironmentRequest models.UpdateEnvironmentRequest
		if err := json.NewDecoder(r.Body).Decode(&updateEnvironmentRequest); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := c.conn.UpdateEnvironment(r.Context(), projectID, environmentID, &updateEnvironmentRequest); err != nil {
			log.Println("Error updating environment:", err)
			http.Error(w, "Failed to update environment", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		// Handle DELETE request
		projectID := r.PathValue("projectId")
		environmentID := r.PathValue("environmentId")

		if err := c.conn.DeleteEnvironment(r.Context(), projectID, environmentID); err != nil {
			log.Println("Error deleting environment:", err)
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
