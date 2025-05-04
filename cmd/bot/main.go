package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/OleksandrOleniuk/twitchong/internal/api/server"
	"github.com/OleksandrOleniuk/twitchong/internal/api/shared"
	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
	"github.com/gorilla/websocket"
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
		if err := srv.Start(); err != nil {
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

	// Start WebSocket client
	conn := startWebSocketClient(appConfig.EventsubWebsocketUrl)
	if conn == nil {
		logger.Error("failed to establish WebSocket connection")
		close(shutdownChan)
		return
	}
	defer conn.Close()

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

func startWebSocketClient(url string) *websocket.Conn {
	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logger.Error("error connecting to WebSocket", zap.Error(err))
		return nil
	}

	logger.Info("WebSocket connection opened", zap.String("url", url))

	// Start a goroutine to handle incoming messages
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				logger.Error("error reading message", zap.Error(err))
				return
			}

			var data interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				logger.Error("error parsing message", zap.Error(err))
				continue
			}

			// Log the parsed data using the new PrettyObject function
			// utils.PrettyObject("received WebSocket message", "data", data)

			handleWebSocketMessage(data)
		}
	}()

	return conn
}

func handleWebSocketMessage(data interface{}) {
	// Convert the generic interface{} to a map to access its fields
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		logger.Error("data is not a map")
		return
	}

	// Extract metadata
	metadata, ok := dataMap["metadata"].(map[string]interface{})
	if !ok {
		logger.Error("metadata is not a map")
		return
	}

	// Get message type
	messageType, ok := metadata["message_type"].(string)
	if !ok {
		logger.Error("message_type is not a string")
		return
	}

	switch messageType {
	case "session_welcome":
		// First message you get from the WebSocket server when connecting
		payload, ok := dataMap["payload"].(map[string]interface{})
		if !ok {
			logger.Error("payload is not a map")
			return
		}

		session, ok := payload["session"].(map[string]interface{})
		if !ok {
			logger.Error("session is not a map")
			return
		}

		sessionID, ok := session["id"].(string)
		if !ok {
			logger.Error("session ID is not a string")
			return
		}

		// Register the Session ID it gives us
		websocketSessionID = sessionID

		// Listen to EventSub, which joins the chatroom from your bot's account
		registerEventSubListeners()

	case "notification":
		// An EventSub notification has occurred, such as channel.chat.message
		subscriptionType, ok := metadata["subscription_type"].(string)
		if !ok {
			logger.Error("subscription_type is not a string")
			return
		}

		switch subscriptionType {
		case "channel.chat.message":
			payload, ok := dataMap["payload"].(map[string]interface{})
			if !ok {
				logger.Error("payload is not a map")
				return
			}

			event, ok := payload["event"].(map[string]interface{})
			if !ok {
				logger.Error("event is not a map")
				return
			}

			broadcasterUserLogin, _ := event["broadcaster_user_login"].(string)
			chatterUserLogin, _ := event["chatter_user_login"].(string)

			message, ok := event["message"].(map[string]interface{})
			if !ok {
				logger.Error("message is not a map")
				return
			}

			text, ok := message["text"].(string)
			if !ok {
				logger.Error("text is not a string")
				return
			}

			// First, print the message to the program's console
			logger.Info("chat message received",
				zap.String("channel", broadcasterUserLogin),
				zap.String("user", chatterUserLogin),
				zap.String("message", text),
			)

			// Then check to see if that message was "HeyGuys"
			if strings.Contains(text, "HeyGuys") {
				// If so, send back "VoHiYo" to the chatroom
				sendChatMessage("VoHiYo")
			}
		}
	}
}

func sendChatMessage(chatMessage string) error {
	// Prepare the request body
	requestBody := map[string]string{
		"broadcaster_id": appConfig.ChatChannelUserId,
		"sender_id":      appConfig.BotUserId,
		"message":        chatMessage,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("error marshaling request body", zap.Error(err))
		return err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/chat/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error("error creating request", zap.Error(err))
		return err
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+appConfig.OauthToken)
	req.Header.Set("Client-Id", appConfig.ClientId)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("error sending request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("failed to send chat message",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return fmt.Errorf("failed to send chat message: status code %d", resp.StatusCode)
	}

	logger.Info("chat message sent", zap.String("message", chatMessage))
	return nil
}

func registerEventSubListeners() error {
	// Create the request body
	requestBody := map[string]interface{}{
		"type":    "channel.chat.message",
		"version": "1",
		"condition": map[string]string{
			"broadcaster_user_id": appConfig.ChatChannelUserId,
			"user_id":             appConfig.BotUserId,
		},
		"transport": map[string]string{
			"method":     "websocket",
			"session_id": websocketSessionID,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("error marshaling request body", zap.Error(err))
		return err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error("error creating request", zap.Error(err))
		return err
	}

	fmt.Printf("requestBody: %v\n", requestBody)

	// Add headers
	req.Header.Set("Authorization", "Bearer "+appConfig.OauthToken)
	req.Header.Set("Client-Id", appConfig.ClientId)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("error sending request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("error reading response body", zap.Error(err))
		return err
	}

	// Check response status (202 Accepted is the expected response)
	if resp.StatusCode != 202 {
		logger.Error("failed to subscribe to channel.chat.message",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return fmt.Errorf("failed to subscribe to channel.chat.message: status code %d", resp.StatusCode)
	}

	// Parse the response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		logger.Error("error parsing response JSON", zap.Error(err))
		return err
	}

	// Extract subscription ID
	dataArray, ok := responseData["data"].([]interface{})
	if !ok || len(dataArray) == 0 {
		logger.Error("unexpected response format")
		return fmt.Errorf("unexpected response format")
	}

	firstItem, ok := dataArray[0].(map[string]interface{})
	if !ok {
		logger.Error("unexpected data item format")
		return fmt.Errorf("unexpected data item format")
	}

	subscriptionID, ok := firstItem["id"].(string)
	if !ok {
		logger.Error("could not find subscription ID")
		return fmt.Errorf("could not find subscription ID")
	}

	logger.Info("subscribed to channel.chat.message", zap.String("subscription_id", subscriptionID))
	return nil
}
