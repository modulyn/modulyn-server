package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
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

	http.ListenAndServe(":8080", nil)
}
