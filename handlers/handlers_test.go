package handlers

import "testing"

func TestHandlers(t *testing.T) {
	newConfig := Config{}
	if newConfig.DB != nil {
		t.Error("bad config")
	}

	//Test config methods

}
