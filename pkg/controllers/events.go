package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (c *controller) EventsController(w http.ResponseWriter, r *http.Request) {
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
			features, err := c.conn.GetFeatures(projectId, sdkKey)
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
}
