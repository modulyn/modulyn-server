package main

import (
	"encoding/json"
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

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		for i := range 10 {
			fmt.Fprintf(w, "data: Event %s\n\n", fmt.Sprintf("Event %d", i))
			time.Sleep(2 * time.Second)
			w.(http.Flusher).Flush()
		}

		closeNotify := r.Context().Done()
		<-closeNotify
	})

	// features
	http.HandleFunc("/api/v1/features", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

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

	// create project
	http.HandleFunc("/api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

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
	})

	// create environment
	http.HandleFunc("/api/v1/environments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

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
	})

	http.ListenAndServe(":8080", nil)
}
