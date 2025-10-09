package controllers

import (
	"encoding/json"
	"log"
	"modulyn/pkg/models"
	"net/http"
	"slices"

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
		queryParams := r.URL.Query()

		searchTerm := queryParams.Get("search")

		features, err := c.conn.GetFeatures(r.Context(), projectID, searchTerm)
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

		environments, err := c.conn.GetEnvironments(r.Context(), projectID)
		if err != nil {
			log.Println("Error getting environments:", err)
			http.Error(w, "Failed to create feature", http.StatusInternalServerError)
			return
		}

		featureID, _ := uuid.NewRandom()

		if err := c.conn.CreateFeature(r.Context(), featureID.String(), projectID, environments, &createFeatureRequest); err != nil {
			log.Println("Error creating feature:", err)
			http.Error(w, "Failed to create feature", http.StatusInternalServerError)
			return
		}

		createdFeatures, err := c.conn.GetFeaturesByID(r.Context(), projectID, featureID.String())
		if err != nil {
			log.Println("Error getting features: ", err)
			return
		}

		for _, f := range createdFeatures {
			bytes, _ := json.Marshal(f)
			event := models.Event{
				Type: "feature_created",
				Data: bytes,
			}

			c.store.NotifyClients(event, f.EnvironmentID)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: featureID,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *controller) FeatureByIdController(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		projectID := r.PathValue("projectId")
		featureID := r.PathValue("featureId")

		features, err := c.conn.GetFeaturesByID(r.Context(), projectID, featureID)
		if err != nil {
			http.Error(w, "Failed to get features", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{
			Data: features,
		})
	case http.MethodPut:
		projectID := r.PathValue("projectId")
		featureID := r.PathValue("featureId")

		var updateFeaturesRequest []*models.UpdateFeatureRequest
		if err := json.NewDecoder(r.Body).Decode(&updateFeaturesRequest); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := c.conn.UpdateFeatures(r.Context(), projectID, featureID, updateFeaturesRequest); err != nil {
			log.Println("Error updating feature:", err)
			http.Error(w, "Failed to update feature", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

		newlyUpdatedFeatures, err := c.conn.GetFeaturesByID(r.Context(), projectID, featureID)
		if err != nil {
			log.Println("Error getting new feature:", err)
			return
		}

		for _, updateFeatureRequest := range updateFeaturesRequest {
			i := slices.IndexFunc(newlyUpdatedFeatures, func(f *models.Feature) bool {
				return f.EnvironmentID == updateFeatureRequest.EnvironmentID
			})
			if i < 0 {
				continue
			}

			bytes, _ := json.Marshal(newlyUpdatedFeatures[i])
			event := models.Event{
				Type: "feature_updated",
				Data: bytes,
			}

			c.store.NotifyClients(event, updateFeatureRequest.EnvironmentID)
		}
	case http.MethodDelete:
		projectID := r.PathValue("projectId")
		featureID := r.PathValue("featureId")

		existingFeature, _ := c.conn.GetFeaturesByID(r.Context(), projectID, featureID)

		if err := c.conn.DeleteFeature(r.Context(), projectID, featureID); err != nil {
			log.Println("Error deleting feature:", err)
			http.Error(w, "Failed to delete feature", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

		for _, feature := range existingFeature {
			bytes, _ := json.Marshal(feature)
			event := models.Event{
				Type: "feature_deleted",
				Data: bytes,
			}

			c.store.NotifyClients(event, feature.EnvironmentID)
		}
	}
}
