package handlers

import (
	"net/http"

	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/internal/pubsub"
	"github.com/google/uuid"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type TypingHandler struct {
	pubsub *pubsub.PubSub
	cfg    *Config
}

func NewTypingHandler(ps *pubsub.PubSub, cfg *Config) *TypingHandler {
	return &TypingHandler{
		pubsub: ps,
		cfg:    cfg,
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
	sent := r.URL.Query().Get("sent")

	if channelID == "" {
		http.Error(w, "channel id required", http.StatusBadRequest)
		return
	}

	if sent == "true" {
		channelUUID, err := uuid.Parse(channelID)
		if err != nil {
			http.Error(w, "invalid channel id", http.StatusBadRequest)
			return
		}

		_, err = h.cfg.DB.AddMessage(r.Context(), database.AddMessageParams{
			Author:  user.ID,
			Channel: channelUUID,
			Body:    message,
		})
		if err != nil {
			http.Error(w, "failed to save message", http.StatusInternalServerError)
			return
		}

		sse := datastar.NewSSE(w, r)
		if err := sse.MergeSignals([]byte(`{"msg": ""}`)); err != nil {
			respondWithErrors(w, r, "error clearing message field", err)
			return
		}
	}

	event := pubsub.TypingEvent{
		UserID:   user.ID.String(),
		Username: user.Username,
		Channel:  channelID,
		Message:  message,
		Sent:     sent == "true",
	}

	h.pubsub.Publish(event.Channel, event)

	w.WriteHeader(http.StatusOK)
}
