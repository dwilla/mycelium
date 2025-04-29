package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func (cfg Config) SendPassReset(w http.ResponseWriter, r *http.Request) {
	signals := struct {
		Email string `json:"email"`
	}{}
	if err := datastar.ReadSignals(r, &signals); err != nil {
		http.Error(w, err.Error(), 500)
	}
	user, err := cfg.DB.GetUserByEmail(r.Context(), signals.Email)
	if err != nil {
		respondWithErrors(w, r, "Email not found in database", err)
		return
	}

	emailUrl := cfg.BaseURL + "/reset/" + user.ID.String()

	mg := mailgun.NewMailgun("sandboxa6b1230594ea46b68aa9b0fa14f4f859.mailgun.org", cfg.Mailgun)
	from := "Mailgun Sandbox <postmaster@sandboxa6b1230594ea46b68aa9b0fa14f4f859.mailgun.org>"
	sub := "Pass Link Test"
	body := fmt.Sprintf("Password reset link: \n%s", emailUrl)
	msg := mailgun.NewMessage(from, sub, body, signals.Email)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err = mg.Send(ctx, msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	sse := datastar.NewSSE(w, r)
	if err := sse.MergeFragments(`<div id="msg">Password reset email sent.</div>`); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
