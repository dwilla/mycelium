package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dwilla/mycelium/internal/pubsub"
)

type TypingHandler struct {
	pubsub *pubsub.PubSub
}

func NewTypingHandler(ps *pubsub.PubSub) *TypingHandler {
	return &TypingHandler{
		pubsub: ps,
	}
}

func (h *TypingHandler) HandleTyping(w http.ResponseWriter, r *http.Request) {
	var event pubsub.TypingEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Publish the message event
	h.pubsub.Publish(event.ChannelID, event)

	w.WriteHeader(http.StatusOK)
}
