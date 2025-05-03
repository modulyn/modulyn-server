package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"modulyn/pkg/db"
	"modulyn/pkg/models"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	conn, err := db.InitDB()
	if err != nil {
		log.Fatalln("Failed to initialize database: ", err)
	}
	defer conn.Close()

	// events
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		sdkKey := r.URL.Query().Get("sdk_key")
		if sdkKey == "" {
			http.Error(w, "Missing sdk_key parameter", http.StatusBadRequest)
			return
		}

		ticker := time.NewTicker(5 * time.Second)
		go func() {
			for range ticker.C {
				// Fetch the features from the database
				features, err := conn.GetFeatures(sdkKey)
				if err != nil {
					log.Printf("Error fetching features for sdk: %s - %+v", sdkKey, err)
					return
				}
				out, err := json.Marshal(features)
				if err != nil {
					log.Printf("Error marshalling features for sdk: %s - %+v", sdkKey, err)
					return
				}
				fmt.Fprintf(w, "data: %+v\n\n", string(out))
				w.(http.Flusher).Flush()
			}
		}()

		closeNotify := r.Context().Done()
		<-closeNotify
		ticker.Stop()
	})

	// features
	http.HandleFunc("/api/v1/features", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			// Handle GET request
			sdkKey := r.URL.Query().Get("sdk_key")
			if sdkKey == "" {
				http.Error(w, "Missing sdk_key parameter", http.StatusBadRequest)
				return
			}

			features, err := conn.GetFeatures(sdkKey)
			if err != nil {
				http.Error(w, "Failed to get features", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(models.Response{
				Data: features,
			})
		case http.MethodPut:
			// Handle PUT request
			var updateFeatureRequest models.UpdateFeatureRequest
			if err := json.NewDecoder(r.Body).Decode(&updateFeatureRequest); err != nil {
				log.Println("Error decoding request body:", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			if err := conn.UpdateFeature(&updateFeatureRequest); err != nil {
				log.Println("Error updating feature:", err)
				http.Error(w, "Failed to update feature", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			// Handle POST request
			var createFeatureRequest models.CreateFeatureRequest
			if err := json.NewDecoder(r.Body).Decode(&createFeatureRequest); err != nil {
				log.Println("Error decoding request body:", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			if err := conn.CreateFeature(&createFeatureRequest); err != nil {
				log.Println("Error creating feature:", err)
				http.Error(w, "Failed to create feature", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			// Handle DELETE request
			featureID := r.URL.Query().Get("id")
			if featureID == "" {
				http.Error(w, "Missing feature ID", http.StatusBadRequest)
				return
			}

			if err := conn.DeleteFeature(featureID); err != nil {
				log.Println("Error deleting feature:", err)
				http.Error(w, "Failed to delete feature", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// get feature
	http.HandleFunc("/api/v1/environments/{environmentId}/features/{featureId}", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			environmentID := r.PathValue("environmentId")
			featureID := r.PathValue("featureId")

			feature, err := conn.GetFeature(featureID, environmentID)
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
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// projects
	http.HandleFunc("/api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			// Handle GET request
			projects, err := conn.GetProjects()
			if err != nil {
				log.Println("Error getting projects:", err)
				http.Error(w, "Failed to get projects", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(models.Response{
				Data: projects,
			})
		case http.MethodPut:
			// Handle PUT request
			var updateProjectRequest models.UpdateProjectRequest
			if err := json.NewDecoder(r.Body).Decode(&updateProjectRequest); err != nil {
				log.Println("Error decoding request body:", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			if err := conn.UpdateProject(&updateProjectRequest); err != nil {
				log.Println("Error updating project:", err)
				http.Error(w, "Failed to update project", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			// Handle POST request
			var createProjectRequest models.CreateProjectRequest
			if err := json.NewDecoder(r.Body).Decode(&createProjectRequest); err != nil {
				log.Println("Error decoding request body:", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			projectID, err := conn.CreateProject(&createProjectRequest)
			if err != nil {
				log.Println("Error creating project:", err)
				http.Error(w, "Failed to create project", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Response{
				Data: projectID,
			})
		case http.MethodDelete:
			// Handle DELETE request
			projectID := r.URL.Query().Get("id")

			if projectID == "" {
				http.Error(w, "Missing project ID", http.StatusBadRequest)
				return
			}

			if err := conn.DeleteProject(projectID); err != nil {
				log.Println("Error deleting project:", err)
				http.Error(w, "Failed to delete project", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// environments
	http.HandleFunc("/api/v1/environments", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			// Handle GET request
			projectID := r.URL.Query().Get("project_id")
			if projectID == "" {
				http.Error(w, "Missing project ID", http.StatusBadRequest)
				return
			}

			environments, err := conn.GetEnvironments(projectID)
			if err != nil {
				http.Error(w, "Failed to get environments", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(models.Response{
				Data: environments,
			})
		case http.MethodPut:
			// Handle PUT request
			var updateEnvironmentRequest models.UpdateEnvironmentRequest
			if err := json.NewDecoder(r.Body).Decode(&updateEnvironmentRequest); err != nil {
				log.Println("Error decoding request body:", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			if err := conn.UpdateEnvironment(&updateEnvironmentRequest); err != nil {
				log.Println("Error updating environment:", err)
				http.Error(w, "Failed to update environment", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			// Handle POST request
			var createEnvironmentRequest models.CreateEnvironmentRequest
			if err := json.NewDecoder(r.Body).Decode(&createEnvironmentRequest); err != nil {
				log.Println("Error decoding request body:", err)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			environmentID, err := conn.CreateEnvironment(&createEnvironmentRequest)
			if err != nil {
				log.Println("Error creating environment:", err)
				http.Error(w, "Failed to create environment", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Response{
				Data: environmentID,
			})
		case http.MethodDelete:
			// Handle DELETE request
			environmentID := r.URL.Query().Get("id")

			if environmentID == "" {
				http.Error(w, "Missing environment ID", http.StatusBadRequest)
				return
			}

			if err := conn.DeleteEnvironment(environmentID); err != nil {
				log.Println("Error deleting environment:", err)
			}

			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":8080", nil)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
}
