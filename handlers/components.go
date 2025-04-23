package handlers

import (
	"net/http"

	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleMain(w http.ResponseWriter, r *http.Request) {
	component := templates.Main()
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (cfg Config) HandleApp(w http.ResponseWriter, r *http.Request) {
	// Make auth function for here
	currentUser := database.User{}
	if currentUser.Username == "" {
		component := templates.Login()

		sse := datastar.NewSSE(w, r)
		sse.MergeFragmentTempl(component)
	}
}
