package controllers

import (
	"encoding/json"
	"fmt"
	"modulyn/pkg/models"
	"net/http"
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

	appId := r.URL.Query().Get("appid")
	if appId == "" {
		http.Error(w, "Missing appid parameter", http.StatusBadRequest)
		return
	}

	client := models.Client{
		SDKKey:   sdkKey,
		AppID:    appId,
		Messages: make(chan models.Event),
	}
	c.store.Subscribe(client)
	defer c.store.Unsubscribe(client)

	go func() {
		for event := range client.Messages {
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		}
	}()

	// send all features to the client when they connect
	features, err := c.conn.GetFeaturesByEnvironmentID(sdkKey)
	if err != nil {
		http.Error(w, "Failed to get features", http.StatusInternalServerError)
		return
	}

	// send the features as an initial event
	featuresData, _ := json.Marshal(features)
	initialEvent := models.Event{
		Type: "all_features",
		Data: featuresData,
	}

	client.Messages <- initialEvent

	closeNotify := r.Context().Done()
	<-closeNotify
}
