package handlers

import (
	"fmt"
	"net/http"

	"github.com/dwilla/mycelium/templates"
	"github.com/google/uuid"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleGetChat(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("id")
	if channelID == "" {
		respondWithErrors(w, r, "no id in path", fmt.Errorf("no id"))
		return
	}

	channel, err := cfg.DB.GetChannelByID(r.Context(), uuid.MustParse(channelID)) //use this channel
	if err != nil {
		respondWithErrors(w, r, "error getting channel", err)
		return
	}

	viewSignals := channelSignals{
		ViewChannel: ViewChannel{
			ID:   channel.ID.String(),
			Name: channel.Name,
		},
	}

	component := templates.Chat()

	sse := datastar.NewSSE(w, r)
	if err := sse.MarshalAndMergeSignals(viewSignals); err != nil {
		respondWithErrors(w, r, "error merging signals", err)
		return
	}
	if err := sse.MergeFragmentTempl(component); err != nil {
		respondWithErrors(w, r, "error merging component", err)
		return
	}
}
