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

		projectId := r.URL.Query().Get("project_id")

		ticker := time.NewTicker(5 * time.Second)
		go func() {
			for range ticker.C {
				// Fetch the features from the database
				features, err := conn.GetFeatures(projectId, sdkKey)
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
	http.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}/features", func(w http.ResponseWriter, r *http.Request) {
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

			features, err := conn.GetFeatures(projectID, environmentID)
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

			featureID, err := conn.CreateFeature(projectID, environmentID, &createFeatureRequest)
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
	})

	http.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}/features/{featureId}", func(w http.ResponseWriter, r *http.Request) {
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

			feature, err := conn.GetFeature(projectID, environmentID, featureID)
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

			if err := conn.UpdateFeature(projectID, environmentID, featureID, &updateFeatureRequest); err != nil {
				log.Println("Error updating feature:", err)
				http.Error(w, "Failed to update feature", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			projectID := r.PathValue("projectId")
			environmentID := r.PathValue("environmentId")
			featureID := r.PathValue("featureId")

			if err := conn.DeleteFeature(projectID, environmentID, featureID); err != nil {
				log.Println("Error deleting feature:", err)
				http.Error(w, "Failed to delete feature", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
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
		case http.MethodPost:
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
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/v1/projects/{projectId}", func(w http.ResponseWriter, r *http.Request) {
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

			if err := conn.UpdateProject(projectID, &updateProjectRequest); err != nil {
				log.Println("Error updating project:", err)
				http.Error(w, "Failed to update project", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			projectID := r.PathValue("projectId")

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
	http.HandleFunc("/api/v1/projects/{projectId}/environments", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			projectID := r.PathValue("projectId")

			environments, err := conn.GetEnvironments(projectID)
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

			environmentID, err := conn.CreateEnvironment(projectID, &createEnvironmentRequest)
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
	})

	http.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
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

			if err := conn.UpdateEnvironment(projectID, environmentID, &updateEnvironmentRequest); err != nil {
				log.Println("Error updating environment:", err)
				http.Error(w, "Failed to update environment", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			// Handle DELETE request
			projectID := r.PathValue("projectId")
			environmentID := r.PathValue("environmentId")

			if err := conn.DeleteEnvironment(projectID, environmentID); err != nil {
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
