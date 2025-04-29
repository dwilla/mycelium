package handlers

import (
	"fmt"
	"net/http"

	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleMain(w http.ResponseWriter, r *http.Request) {
	component := templates.Main(true)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (cfg Config) HandleHome(w http.ResponseWriter, r *http.Request) {
	user, ok := GetCurrentUser(r)
	if !ok {
		respondWithErrors(w, r, "user not authenticated", fmt.Errorf("user not authenticated"))
		return
	}
	channels, err := cfg.DB.GetChannelsForUser(r.Context(), user.ID)
	if err != nil {
		respondWithErrors(w, r, "error with channels", err)
		return
	}
	viewChannel := database.Channel{}
	if len(channels) != 0 {
		viewChannel, err = cfg.DB.GetChannelByID(r.Context(), channels[0].ID)
		if err != nil {
			respondWithErrors(w, r, "database error", err)
			return
		}
	}

	component := templates.Home(viewChannel)
	sse := datastar.NewSSE(w, r)
	if err := sse.MergeFragmentTempl(component); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
