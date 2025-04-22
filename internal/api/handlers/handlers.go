package handlers

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/OleksandrOleniuk/twitchong/internal/api/shared"
	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
)

// HealthCheck provides an endpoint to check API health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// TwitchCallbackHandler handles the OAuth callback from Twitch
type TwitchCallbackHandler struct {
	cfg *config.Provider
}

// NewTwitchCallback creates a new TwitchCallbackHandler
func NewTwitchCallback(cfg *config.Provider) http.HandlerFunc {
	handler := &TwitchCallbackHandler{cfg: cfg}
	return handler.Handle
}

// Handle processes the Twitch OAuth callback
func (h *TwitchCallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// For GET requests, serve the callback page
		http.ServeFile(w, r, "web/static/callback.html")
		return
	case "POST":
		// For POST requests, read the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			w.WriteHeader(http.StatusOK)
			return
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			log.Printf("Error parsing data: %v", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Extract parameters from the parsed values
		accessToken := values.Get("access_token")
		state := values.Get("state")

		// Verify state parameter
		if state != h.cfg.Get().TwitchSecretState {
			log.Printf("Invalid state parameter. Expected: %s, Got: %s", h.cfg.Get().TwitchSecretState, state)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Send access token to main goroutine
		shared.OAuthTokenChan <- accessToken
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Return 200 OK
	w.WriteHeader(http.StatusOK)
}
