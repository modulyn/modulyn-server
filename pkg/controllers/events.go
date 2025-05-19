package controllers

import (
	"fmt"
	"modulyn/pkg/models"
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

	appId := r.URL.Query().Get("appid")
	if appId == "" {
		http.Error(w, "Missing appid parameter", http.StatusBadRequest)
		return
	}

	client := models.Client{
		SDKKey: sdkKey,
		AppID:  appId,
	}
	c.store.Subscribe(client)
	defer c.store.Unsubscribe(client)

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			events := "events"
			fmt.Fprintf(w, "data: %+v\n\n", events)
			w.(http.Flusher).Flush()
		}
	}()

	closeNotify := r.Context().Done()
	<-closeNotify
	ticker.Stop()
}
