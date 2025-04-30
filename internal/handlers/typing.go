package handlers

import (
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
	user, ok := GetCurrentUser(r)
	if !ok {
		http.Error(w, "authentication error", http.StatusInternalServerError)
		return
	}

	channelID := r.URL.Query().Get("channel")
	message := r.URL.Query().Get("message")

	if channelID == "" {
		http.Error(w, "channel id required", http.StatusBadRequest)
		return
	}

	event := pubsub.TypingEvent{
		UserID:   user.ID.String(),
		Username: user.Username,
		Channel:  channelID,
		Message:  message,
	}

	h.pubsub.Publish(event.Channel, event)

	w.WriteHeader(http.StatusOK)
}
