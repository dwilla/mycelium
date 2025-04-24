package handlers

import (
	"net/http"

	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleMain(w http.ResponseWriter, r *http.Request) {
	component := templates.Main()
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (cfg Config) HandleHome(w http.ResponseWriter, r *http.Request) {
	component := templates.Home()
	sse := datastar.NewSSE(w, r)
	if err := sse.MergeFragmentTempl(component); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
