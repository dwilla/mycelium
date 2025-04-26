package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dwilla/mycelium/internal/auth"
	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) HandleSignOut(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (cfg Config) HandleLogin(w http.ResponseWriter, r *http.Request) {
	signals := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := datastar.ReadSignals(r, &signals); err != nil {
		http.Error(w, err.Error(), 500)
	}
	user, err := cfg.DB.GetUserByEmail(r.Context(), signals.Email)
	if err != nil {
		respondWithErrors(w, r, "Email not found in database", err)
		return
	}
	if err := auth.CheckPasswordHash(user.PasswordHash.String, signals.Password); err != nil {
		respondWithErrors(w, r, "Incorrect Password", err)
		return
	}

	newToken, err := auth.MakeRefreshToken()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	refreshToken, err := cfg.DB.MakeToken(r.Context(), database.MakeTokenParams{
		Token:     newToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(1440 * time.Hour),
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.JwtSecret, time.Hour)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Log the cookie settings
	log.Printf("Setting cookies for domain: %s", r.Host)
	log.Printf("Token length: %d", len(token))
	log.Printf("Refresh token length: %d", len(refreshToken.Token))

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    refreshToken.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(1440 * time.Hour),
	})

	sse := datastar.NewSSE(w, r)
	if err := sse.MergeSignals([]byte(`{"auth":true}`)); err != nil {
		http.Error(w, "issue merging auth signal", 500)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (cfg Config) HandleNewUser(w http.ResponseWriter, r *http.Request) {
	signals := struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err := datastar.ReadSignals(r, &signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hashedPass, err := auth.HashPassword(signals.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:        signals.Email,
		Username:     signals.Username,
		PasswordHash: sql.NullString{String: hashedPass},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newToken, err := auth.MakeRefreshToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, err := cfg.DB.MakeToken(r.Context(), database.MakeTokenParams{
		Token:     newToken,
		UserID:    newUser.ID,
		ExpiresAt: time.Now().Add(1440 * time.Hour),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := auth.MakeJWT(newUser.ID, cfg.JwtSecret, time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    refreshToken.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(1440 * time.Hour),
	})
	sse := datastar.NewSSE(w, r)

	if err := sse.MergeSignals([]byte(`{"auth":true}`)); err != nil {
		http.Error(w, "can't update signals", http.StatusInternalServerError)
		return
	}

	component := templates.Home()
	if err := sse.MergeFragmentTempl(component); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
