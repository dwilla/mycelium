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

	// If we have a channel, load its messages
	if len(channels) != 0 {
		if err := sse.MergeFragments(
			`<ul id="messages"></ul>`,
			datastar.WithSelector("#messages"),
			datastar.WithMergeMode("outer"),
		); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		messages, err := cfg.DB.GetMessagesForChannel(r.Context(), channels[0].ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		for _, message := range messages {
			if err := sse.MergeFragments(
				fmt.Sprintf(`<li>%s:<br>%s</li>`, message.Username, message.Body),
				datastar.WithSelector("#messages"),
				datastar.WithMergeMode("append"),
			); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}
	}
}
