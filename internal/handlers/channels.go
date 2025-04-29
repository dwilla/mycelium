package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleNewChannelComponent(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	component := templates.NewChannel()
	if err := sse.MergeFragmentTempl(component); err != nil {
		respondWithErrors(w, r, "error getting new component", err)
	}
}

func (cfg Config) HandleNewChannel(w http.ResponseWriter, r *http.Request) {
	user, ok := GetCurrentUser(r)
	if !ok {
		respondWithErrors(w, r, "Authentication issue", fmt.Errorf("error finding authenticated user"))
		return
	}
	signal := struct {
		NewName string `json:"name"`
	}{}

	if err := datastar.ReadSignals(r, &signal); err != nil {
		respondWithErrors(w, r, "Signal read issue", err)
		return
	}

	newChannel, err := cfg.DB.CreateChannel(r.Context(), database.CreateChannelParams{
		Name:    signal.NewName,
		Creator: user.ID,
	})
	if err != nil {
		respondWithErrors(w, r, "Name must be unique.", err)
		return
	}

	_, err = cfg.DB.CreateSub(r.Context(), database.CreateSubParams{
		UserID:    user.ID,
		ChannelID: newChannel.ID,
	})
	if err != nil {
		respondWithErrors(w, r, "subs database error", err)
		return
	}

	component := templates.Home(newChannel)

	sse := datastar.NewSSE(w, r)
	if err := sse.MergeFragmentTempl(component); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cfg Config) GetUserChannels(w http.ResponseWriter, r *http.Request) {
	user, ok := GetCurrentUser(r)
	if !ok {
		respondWithErrors(w, r, "couldn't authenticate user", fmt.Errorf("authentication error"))
		return
	}
	channels, err := cfg.DB.GetChannelsForUser(r.Context(), user.ID)
	if err != nil {
		respondWithErrors(w, r, "error getting channels from db", err)
	}

	sse := datastar.NewSSE(w, r)

	var fragments strings.Builder
	fragments.WriteString(`<div id="user-channels">`)
	for i, channel := range channels {
		fragments.WriteString(fmt.Sprintf(
			`<button id="channel-%v" data-on-click="@get('/chat/%v')">%v</button>`,
			i,
			channel.ID,
			channel.Name,
		))
	}
	fragments.WriteString("</div>")

	if err := sse.MergeFragments(fragments.String()); err != nil {
		respondWithErrors(w, r, "error merging fragments", err)
	}
}
