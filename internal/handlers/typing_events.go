package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

func (h *TypingHandler) HandleTypingEvents(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		http.Error(w, "Channel ID required", http.StatusBadRequest)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel to receive message events
	events := h.pubsub.Subscribe(channelID)
	defer h.pubsub.Unsubscribe(channelID, events)

	// Create a channel to detect client disconnection
	clientGone := r.Context().Done()

	for {
		select {
		case event := <-events:
			// Send the message event to the client
			data, _ := json.Marshal(event)
			_, err := w.Write([]byte("data: " + string(data) + "\n\n"))
			if err != nil {
				respondWithErrors(w, r, "writing error in typeing events", err)
			}
			w.(http.Flusher).Flush()
		case <-clientGone:
			return
		case <-time.After(30 * time.Second):
			// Send a keep-alive comment
			_, err := w.Write([]byte(": keepalive\n\n"))
			if err != nil {
				respondWithErrors(w, r, "writing error in typeing events", err)
			}
			w.(http.Flusher).Flush()
		}
	}
}
