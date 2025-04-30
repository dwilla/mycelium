package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type ViewChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type channelSignals struct {
	ViewChannel ViewChannel `json:"viewChannel"`
}

func (cfg Config) HandleNewChannelComponent(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	component := templates.NewChannel()
	if err := sse.MergeFragmentTempl(component); err != nil {
		respondWithErrors(w, r, "error getting new component", err)
		return
	}
}

func (cfg Config) HandleNewChannel(w http.ResponseWriter, r *http.Request) {
	user, ok := GetCurrentUser(r)
	if !ok {
		respondWithErrors(w, r, "Authentication issue", fmt.Errorf("error finding authenticated user"))
		return
	}

	// Log the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
	} else {
		log.Printf("Request body: %s", string(body))
		// Restore the request body so it can be read again
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	signal := struct {
		NewChanName string `json:"newChanName"`
	}{}

	if err := datastar.ReadSignals(r, &signal); err != nil {
		log.Println("Value for name signal: ", signal.NewChanName)
		respondWithErrors(w, r, "Signal read issue", err)
		return
	}

	newChannel, err := cfg.DB.CreateChannel(r.Context(), database.CreateChannelParams{
		Name:    signal.NewChanName,
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

	viewSignals := channelSignals{
		ViewChannel: ViewChannel{
			ID:   newChannel.ID.String(),
			Name: newChannel.Name,
		},
	}

	component := templates.Home()

	sse := datastar.NewSSE(w, r)
	if err := sse.MergeFragmentTempl(component); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := sse.MarshalAndMergeSignals(viewSignals); err != nil {
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
