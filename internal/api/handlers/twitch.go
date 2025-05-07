package handlers

import (
	"fmt"
	"net/http"

	"github.com/OleksandrOleniuk/twitchong/internal/api/shared"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
	"github.com/OleksandrOleniuk/twitchong/views"
	"github.com/labstack/echo/v4"
)

// TwitchError represents the error returned from Twitch
type TwitchError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	State            string `json:"state"`
}

// Define a state store to prevent CSRF attacks
var stateStore = make(map[string]bool)

func SetState(key string, value bool) {
	stateStore[key] = value
}

func GetState() map[string]bool {
	return stateStore
}

// handleTwitchCallback handles the redirect from Twitch
func HandleTwitchCallback(c echo.Context) error {
	// Check for errors in query parameters
	errorCode := c.QueryParam("error")
	if errorCode != "" {
		errorDesc := c.QueryParam("error_description")
		state := c.QueryParam("state")

		// Validate state to prevent CSRF
		if _, validState := stateStore[state]; !validState {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid state parameter",
			})
		}

		// Clean up state
		delete(stateStore, state)

		// Handle the error
		twitchError := TwitchError{
			Error:            errorCode,
			ErrorDescription: errorDesc,
			State:            state,
		}

		c.Logger().Errorf("Twitch auth error: %v", twitchError)
		return c.Render(http.StatusBadRequest, "error.html", twitchError)
	}

	// For successful authentication, tokens are in the fragment
	// We need to render a page that extracts the fragment and sends it back
	return utils.TemplRender(c, http.StatusOK, views.FragmentsExtractionPage())
}

// TwitchTokens represents the tokens returned from Twitch
type TwitchTokens struct {
	AccessToken string `form:"access_token" json:"access_token"`
	IDToken     string `form:"id_token" json:"id_token"`
	Scope       string `form:"scope" json:"scope"`
	State       string `form:"state" json:"state"`
	TokenType   string `form:"token_type" json:"token_type"`
}

// processTokens handles the tokens sent from the client-side JavaScript
func ProcessTokens(c echo.Context) error {
	tokens := new(TwitchTokens)
	if err := c.Bind(tokens); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid token data",
		})
	}

	fmt.Printf("tokens: %v\n", tokens)

	// Validate state to prevent CSRF
	if _, validState := stateStore[tokens.State]; !validState {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid state parameter",
		})
	}

	// Clean up state
	delete(stateStore, tokens.State)

	// In a production app, you would:
	// 1. Decode and validate the ID token
	// 2. Extract the user information from the ID token
	// 3. Store the access token securely (e.g., encrypted in a cookie or database)

	// For this example, we'll simulate creating a session
	// Normally you would decode the JWT and extract user info
	// user := &User{
	// 	ID:       "12345", // This would come from the ID token
	// 	Username: "twitch_user",
	// 	Email:    "user@example.com",
	// }

	// Create a session cookie
	// cookie := new(http.Cookie)
	// cookie.Name = "session_id"
	// cookie.Value = "some-secure-session-id" // In real app, generate a secure session ID
	// cookie.HttpOnly = true
	// cookie.Path = "/"
	// c.SetCookie(cookie)

	// c.Logger().Infof("User authenticated: %s", user.Username)

	// Send access token to main goroutine
	shared.OAuthTokenChan <- tokens.AccessToken

	c.Response().Header().Set("HX-Redirect", "/")

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Authentication completed successfully",
	})
}
