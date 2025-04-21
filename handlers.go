package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
)

// Handlers ---------------------------------

func (cfg config) handleMain(w http.ResponseWriter, r *http.Request) {
	// Make auth function for here
	currentUser := database.User{}
	if currentUser.Username != "" {
		component := templates.Main(currentUser)
		if err := component.Render(r.Context(), w); err != nil {
			respondWithError(w, 500, "error rendering component", err)
		}
	} else {
		component := templates.Login()
		if err := component.Render(r.Context(), w); err != nil {
			respondWithError(w, 500, "error rendering component", err)
		}
	}
}

// Responses --------------------------------

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	resBody := errorResponse{
		Error: msg + ":" + err.Error(),
	}
	jsonRes, err := json.Marshal(resBody)
	if err != nil {
		log.Printf("error marshalling response: %v", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	log.Fatal(w.Write(jsonRes))
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	jsonRes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling payload: %v", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	log.Fatal(w.Write(jsonRes))
}
