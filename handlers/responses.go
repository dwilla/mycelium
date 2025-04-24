package handlers

import (
	"fmt"
	"log"
	"net/http"

	datastar "github.com/starfederation/datastar/sdk/go"
)

func respondWithErrors(w http.ResponseWriter, r *http.Request, frontendMsg string, err error) error {
	sse := datastar.NewSSE(w, r)
	log.Println(frontendMsg, err)
	newFragment := fmt.Sprintf(`<div id="errors">%s</div>`, frontendMsg)
	if err := sse.MergeFragments(newFragment); err != nil {
		return err
	}
	return nil
}
