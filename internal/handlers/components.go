package handlers

import (
	"fmt"
	"net/http"

	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleMain(w http.ResponseWriter, r *http.Request) {
	tokenCookie, _ := r.Cookie("token")
	refreshCookie, _ := r.Cookie("refresh-token")
	isAuthenticated := tokenCookie != nil || refreshCookie != nil

	component := templates.Main(true, isAuthenticated)
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
	viewSignals := channelSignals{
		ViewChannel: ViewChannel{},
	}
	if len(channels) != 0 {
		viewSignals.ViewChannel.ID = channels[0].ID.String()
		viewSignals.ViewChannel.Name = channels[0].Name
	}

	component := templates.Home()
	sse := datastar.NewSSE(w, r)

	if err := sse.MarshalAndMergeSignals(viewSignals); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := sse.MergeFragmentTempl(component); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
