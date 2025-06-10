package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/OleksandrOleniuk/twitchong/internal/api/server"
	"github.com/OleksandrOleniuk/twitchong/internal/api/shared"
	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/OleksandrOleniuk/twitchong/internal/websocket"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"

	"go.uber.org/zap"
)

var (
	appConfig          *config.Config
	websocketSessionID = ""
	shutdownChan       = make(chan struct{})
	wg                 sync.WaitGroup
	logger             = utils.With(zap.String("component", "main"))
)

func main() {
	var appConfigErr error
	appConfig, appConfigErr = config.Load()

	if appConfigErr != nil {
		logger.Error("Failed to load app config")
	}

	// Initialize the server
	srv := server.New(appConfig)

	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(srv, appConfig); err != nil {
			logger.Error("server error", zap.Error(err))
			// If server fails to start, initiate graceful shutdown
			close(shutdownChan)
		}
	}()

	// Give the server a moment to start or fail
	time.Sleep(100 * time.Millisecond)

	// Wait for OAuth token from callback
	logger.Info("waiting for OAuth token from callback")
	accessToken := <-shared.OAuthTokenChan
	logger.Info("received OAuth token, proceeding with validation")

	// Update config with new access token
	appConfig.OauthToken = accessToken

	// Validate OAuth token
	validateOAuthToken()

	chat := websocket.NewTwitchChat(appConfig)

	// Start the ws in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := websocket.StartTwitchChat(chat); err != nil {
			logger.Error("server error", zap.Error(err))
			close(shutdownChan)
		}
	}()

	// Give the ws a moment to start or fail
	time.Sleep(100 * time.Millisecond)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either shutdown signal or server error
	select {
	case <-sigChan:
		fmt.Println("\nReceived shutdown signal...")
	case <-shutdownChan:
		fmt.Println("\nServer or WebSocket error detected, initiating shutdown...")
	}

	// Initiate graceful shutdown
	fmt.Println("Shutting down...")
	close(shutdownChan)

	// Wait for all goroutines to complete
	wg.Wait()
	fmt.Println("Shutdocomplete")
}

func validateOAuthToken() {
	config := utils.RequestConfig{
		Method: "GET",
		URL:    "https://id.twitch.tv/oauth2/validate",
		Headers: map[string]string{
			"Authorization": "OAuth " + appConfig.OauthToken,
		},
	}

	var validation struct {
		ClientID  string   `json:"client_id"`
		Login     string   `json:"login"`
		Scopes    []string `json:"scopes"`
		UserID    string   `json:"user_id"`
		ExpiresIn int      `json:"expires_in"`
	}

	err := utils.SendRequestAndParseResponse(config, &validation)
	if err != nil {
		logger.Error("token validation failed", zap.Error(err))
		return
	}

	logger.Info("token validated",
		zap.String("user", validation.Login),
		zap.Strings("scopes", validation.Scopes),
	)
}
