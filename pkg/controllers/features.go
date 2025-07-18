package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"modulyn/pkg/db"
	"modulyn/pkg/models"
	"net/http"

	"github.com/google/uuid"
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

		features, err := c.conn.GetFeaturesByProjectID(projectID)
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
		var createFeatureRequest models.CreateFeatureRequest
		if err := json.NewDecoder(r.Body).Decode(&createFeatureRequest); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		environments, err := c.conn.GetEnvironments(projectID)
		if err != nil {
			log.Println("Error getting environments:", err)
			http.Error(w, "Failed to create feature", http.StatusInternalServerError)
			return
		}

		featureID, _ := uuid.NewRandom()

		for _, env := range environments {
			if err := c.conn.CreateFeature(featureID.String(), projectID, env.ID, &createFeatureRequest); err != nil {
				log.Println("Error creating feature:", err)
				http.Error(w, "Failed to create feature", http.StatusInternalServerError)
				return
			}

			newFeature, err := c.conn.GetFeature(projectID, env.ID, featureID.String())
			if err != nil {
				log.Println("Error getting new feature:", err)
				return
			}
			bytes, _ := json.Marshal(newFeature)
			event := models.Event{
				Type: "feature_created",
				Data: bytes,
			}

			c.store.NotifyClients(event, env.ID)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: featureID,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *controller) FeaturesByEnvironmentIDController(w http.ResponseWriter, r *http.Request) {
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

		updatedFeature, err := c.conn.GetFeature(projectID, environmentID, featureID)
		if err != nil {
			log.Println("Error getting new feature:", err)
			return
		}
		bytes, _ := json.Marshal(updatedFeature)
		event := models.Event{
			Type: "feature_updated",
			Data: bytes,
		}

		c.store.NotifyClients(event, environmentID)
	case http.MethodDelete:
		projectID := r.PathValue("projectId")
		environmentID := r.PathValue("environmentId")
		featureID := r.PathValue("featureId")

		existingFeature, _ := c.conn.GetFeature(projectID, environmentID, featureID)

		if err := c.conn.DeleteFeature(projectID, environmentID, featureID); err != nil {
			log.Println("Error deleting feature:", err)
			http.Error(w, "Failed to delete feature", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

		bytes, _ := json.Marshal(existingFeature)
		event := models.Event{
			Type: "feature_deleted",
			Data: bytes,
		}

		c.store.NotifyClients(event, environmentID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
