package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleGetChat(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("id")
	if channelID == "" {
		respondWithErrors(w, r, "no id in path", fmt.Errorf("no id"))
		return
	}

	channel, err := cfg.DB.GetChannelByID(r.Context(), uuid.MustParse(channelID))
	if err != nil {
		respondWithErrors(w, r, "error getting channel", err)
		return
	}

	messages, err := cfg.DB.GetMessagesForChannel(r.Context(), channel.ID)
	if err != nil {
		respondWithErrors(w, r, "error getting messages", err)
		return
	}

	sse := datastar.NewSSE(w, r)

	if err := sse.MergeFragments(
		`<ul id="messages"></ul>`,
		datastar.WithSelector("#messages"),
		datastar.WithMergeMode("outer"),
	); err != nil {
		respondWithErrors(w, r, "error clearing messages", err)
		return
	}

	for _, message := range messages {
		if err := sse.MergeFragments(
			fmt.Sprintf(`<li>%s:<br>%s</li>`, message.Username, message.Body),
			datastar.WithSelector("#messages"),
			datastar.WithMergeMode("append"),
		); err != nil {
			respondWithErrors(w, r, "error merging message fragment", err)
			return
		}
	}

	if err := sse.MergeSignals([]byte(`{"typingEvent": "", "msg": ""}`)); err != nil {
		respondWithErrors(w, r, "error clearing signals", err)
		return
	}

	sse.Context().Done()
}
