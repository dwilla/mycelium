package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dwilla/mycelium/internal/auth"
	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleValid(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	log.Println(token)
	_, err = auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	sse := datastar.NewSSE(w, r)
	sse.MergeFragments(`<div id="resp">Yes!</div>`)
}

func (cfg Config) HandleNewUser(w http.ResponseWriter, r *http.Request) {
	signals := struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err := datastar.ReadSignals(r, &signals); err != nil {
		http.Error(w, err.Error(), 500)
	}

	hashedPass, err := auth.HashPassword(signals.Password)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	newUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:        signals.Email,
		Username:     signals.Username,
		PasswordHash: sql.NullString{String: hashedPass},
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	token, err := auth.MakeJWT(newUser.ID, cfg.JwtSecret, time.Hour)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	json := fmt.Sprintf(`{"token": "%s"}`, token)
	sse := datastar.NewSSE(w, r)
	sse.MergeSignals([]byte(json))

	component := templates.Home()
	if err := sse.MergeFragmentTempl(component); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (cfg Config) CheckEmail(w http.ResponseWriter, r *http.Request) {
	signal := struct {
		Email string `json:"email"`
	}{}

	if err := datastar.ReadSignals(r, &signal); err != nil {
		http.Error(w, err.Error(), 500)
	}

	sse := datastar.NewSSE(w, r)

	if !strings.Contains(signal.Email, "@") || !strings.Contains(signal.Email, ".") {
		if err := sse.MergeSignals([]byte(`{valid: false}`)); err != nil {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	if err := sse.MergeSignals([]byte(`{valid: true}`)); err != nil {
		http.Error(w, err.Error(), 500)
	}

	if _, err := cfg.DB.GetUserByEmail(r.Context(), signal.Email); err != nil {
		return
	}

	if err := sse.MergeSignals([]byte(`{exists: true}`)); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (cfg Config) CheckUsername(w http.ResponseWriter, r *http.Request) {
	signal := struct {
		Username string `json:"username"`
	}{}

	if err := datastar.ReadSignals(r, &signal); err != nil {
		http.Error(w, err.Error(), 500)
	}

	sse := datastar.NewSSE(w, r)

	if strings.Contains(signal.Username, " ") || signal.Username == "" {
		if err := sse.MergeSignals([]byte(`{'user-valid': false}`)); err != nil {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	if _, err := cfg.DB.GetUserByUsername(r.Context(), signal.Username); err == nil {
		return
	}

	if err := sse.MergeSignals([]byte(`{'user-valid': true}`)); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (cfg Config) CheckPassword(w http.ResponseWriter, r *http.Request) {
	signal := struct {
		Password string `json:"password"`
	}{}

	if err := datastar.ReadSignals(r, &signal); err != nil {
		http.Error(w, err.Error(), 500)
	}

	sse := datastar.NewSSE(w, r)

	if len(signal.Password) < 12 {
		if err := sse.MergeSignals([]byte(`{'pass-valid': false}`)); err != nil {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	if err := sse.MergeSignals([]byte(`{'pass-valid': true}`)); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
