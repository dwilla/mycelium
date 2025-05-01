package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
	"github.com/google/uuid"
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
			`<button id="channel-%v" data-on-click="@get('/channel/%v')">%v</button>`,
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

func (cfg Config) HandleCheckChannel(w http.ResponseWriter, r *http.Request) {
	signal := struct {
		NewChanName string `json:"newChanName"`
	}{}

	if err := datastar.ReadSignals(r, &signal); err != nil {
		log.Println("Value for name signal: ", signal.NewChanName)
		respondWithErrors(w, r, "Signal read issue", err)
		return
	}

	if len(signal.NewChanName) != 36 {
		sse := datastar.NewSSE(w, r)
		if err := sse.MergeSignals([]byte(`{"chanExisting": false}`)); err != nil {
			respondWithErrors(w, r, "error updating signals", err)
			return
		}
		return
	}

	_, err := cfg.DB.GetChannelByID(r.Context(), uuid.MustParse(signal.NewChanName))
	if err != nil {
		if err == sql.ErrNoRows {
			sse := datastar.NewSSE(w, r)
			if err := sse.MergeSignals([]byte(`{"chanExisting": false}`)); err != nil {
				respondWithErrors(w, r, "error updating signals", err)
				return
			}
			return
		}
		respondWithErrors(w, r, "error accessing db", err)
		return
	}

	sse := datastar.NewSSE(w, r)
	if err := sse.MergeSignals([]byte(`{"chanExisting": true}`)); err != nil {
		respondWithErrors(w, r, "error updating signals", err)
		return
	}
}

func (cfg Config) HandleNewSub(w http.ResponseWriter, r *http.Request) {
	user, ok := GetCurrentUser(r)
	if !ok {
		respondWithErrors(w, r, "authentication error", fmt.Errorf("auth error"))
	}

	signal := struct {
		NewChanName string `json:"newChanName"`
	}{}

	if err := datastar.ReadSignals(r, &signal); err != nil {
		log.Println("Value for name signal: ", signal.NewChanName)
		respondWithErrors(w, r, "Signal read issue", err)
		return
	}

	channel, err := cfg.DB.GetChannelByID(r.Context(), uuid.MustParse(signal.NewChanName))
	if err != nil {
		respondWithErrors(w, r, "error accessing db", err)
		return
	}

	_, err = cfg.DB.CreateSub(r.Context(), database.CreateSubParams{
		UserID:    user.ID,
		ChannelID: channel.ID,
	})
	if err != nil {
		respondWithErrors(w, r, "subs database error", err)
		return
	}

	viewSignals := channelSignals{
		ViewChannel: ViewChannel{
			ID:   channel.ID.String(),
			Name: channel.Name,
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

func (cfg Config) HandleGetChannel(w http.ResponseWriter, r *http.Request) {
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

	// First send a signal to close the old typing events connection
	closeSSE := datastar.NewSSE(w, r)
	if err := closeSSE.MergeSignals([]byte(`{"closeTypingEvents": true}`)); err != nil {
		respondWithErrors(w, r, "error sending close signal", err)
		return
	}
	closeSSE.Context().Done()

	// Now update the view with the new channel
	viewSignals := channelSignals{
		ViewChannel: ViewChannel{
			ID:   channel.ID.String(),
			Name: channel.Name,
		},
	}

	sse := datastar.NewSSE(w, r)
	if err := sse.MarshalAndMergeSignals(viewSignals); err != nil {
		respondWithErrors(w, r, "error merging signals", err)
		return
	}

	sse.Context().Done()

	chatHandler := cfg.HandleGetChat
	chatHandler(w, r)
}
