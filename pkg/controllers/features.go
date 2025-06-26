package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"modulyn/pkg/db"
	"modulyn/pkg/models"
	"net/http"
)

func (c *controller) FeaturesController(w http.ResponseWriter, r *http.Request) {
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

		features, err := c.conn.GetFeatures(projectID, environmentID)
		if err != nil {
			http.Error(w, "Failed to get features", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: features,
		})
	case http.MethodPost:
		projectID := r.PathValue("projectId")
		environmentID := r.PathValue("environmentId")
		var createFeatureRequest models.CreateFeatureRequest
		if err := json.NewDecoder(r.Body).Decode(&createFeatureRequest); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		featureID, err := c.conn.CreateFeature(projectID, environmentID, &createFeatureRequest)
		if err != nil {
			log.Println("Error creating feature:", err)
			http.Error(w, "Failed to create feature", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: featureID,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *controller) FeatureByIdControllers(w http.ResponseWriter, r *http.Request) {
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
		featureID := r.PathValue("featureId")
 
		feature, err := c.conn.GetFeature(projectID, environmentID, featureID)
		if err != nil {
			if errors.Is(err, db.ErrNoRows) {
				http.Error(w, "No feature found", http.StatusNoContent)
				return
			}
			http.Error(w, "Failed to get feature", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: feature,
		})
	case http.MethodPut:
		projectID := r.PathValue("projectId")
		environmentID := r.PathValue("environmentId")
		featureID := r.PathValue("featureId")
		var updateFeatureRequest models.UpdateFeatureRequest
		if err := json.NewDecoder(r.Body).Decode(&updateFeatureRequest); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := c.conn.UpdateFeature(projectID, environmentID, featureID, &updateFeatureRequest); err != nil {
			log.Println("Error updating feature:", err)
			http.Error(w, "Failed to update feature", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		projectID := r.PathValue("projectId")
		environmentID := r.PathValue("environmentId")
		featureID := r.PathValue("featureId")

		if err := c.conn.DeleteFeature(projectID, environmentID, featureID); err != nil {
			log.Println("Error deleting feature:", err)
			http.Error(w, "Failed to delete feature", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
