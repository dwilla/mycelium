package handlers

import (
	"fmt"
	"net/http"
	"time"

	datastar "github.com/starfederation/datastar/sdk/go"
)

func (h *TypingHandler) HandleTypingEvents(w http.ResponseWriter, r *http.Request) {
	_, ok := GetCurrentUser(r)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		http.Error(w, "Channel ID required", http.StatusBadRequest)
		return
	}

	events := h.pubsub.Subscribe(channelID)
	defer h.pubsub.Unsubscribe(channelID, events)

	sse := datastar.NewSSE(w, r)
	clientGone := r.Context().Done()

	for {
		select {
		case event := <-events:
			if event.Message == "" {
				if err := sse.MergeSignals([]byte(`{"typingEvent": ""}`)); err != nil {
					return
				}
			} else {
				if err := sse.MergeSignals(fmt.Appendf(nil, `{"typingEvent": "%s: %s"}`, event.Username, event.Message)); err != nil {
					return
				}
			}
		case <-clientGone:
			return
		case <-time.After(30 * time.Second):
			if err := sse.MergeSignals([]byte(`{"keepalive": true}`)); err != nil {
				return
			}
		}
	}
}
