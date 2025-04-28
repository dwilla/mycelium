package handlers

import (
	"context"
	"net/http"

	"github.com/dwilla/mycelium/internal/auth"
	"github.com/dwilla/mycelium/internal/database"
	"github.com/dwilla/mycelium/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type contextKey string

const userContextKey contextKey = "user"

func (cfg Config) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			renderLogin(w, r)
			return
		}

		refreshCookie, err := r.Cookie("refresh-token")
		if err != nil {
			renderLogin(w, r)
			return
		}

		userID, err := auth.ValidateJWT(tokenCookie.Value, cfg.JwtSecret)
		if err != nil {
			if refreshCookie.Value == "" {
				renderLogin(w, r)
				return
			}
			renderLogin(w, r)
			return
		}

		user, err := cfg.DB.GetUserFromId(r.Context(), userID)
		if err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		sse := datastar.NewSSE(w, r)
		if err := sse.MergeSignals([]byte(`{"auth":true}`)); err != nil {
			http.Error(w, "issue merging auth signal", 500)
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetCurrentUser(r *http.Request) (database.User, bool) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	return user, ok
}

func renderLogin(w http.ResponseWriter, r *http.Request) {
	component := templates.Login()
	sse := datastar.NewSSE(w, r)
	if err := sse.MergeFragmentTempl(component); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
