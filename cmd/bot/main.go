package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/OleksandrOleniuk/twitchong/internal/api/server"
	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/gorilla/websocket"
)

var (
	websocketSessionID = ""
	shutdownChan       = make(chan struct{})
	wg                 sync.WaitGroup
)

func main() {
	// Load configuration
	cfgProvider := config.MustNewConfigProvider()

	// Initialize the server
	srv := server.New(cfgProvider)

	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
			// If server fails to start, initiate graceful shutdown
			close(shutdownChan)
		}
	}()

	// Give the server a moment to start or fail
	time.Sleep(100 * time.Millisecond)

	// Validate OAuth token in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		validateOAuthToken(cfgProvider.Get())
	}()

	// Start WebSocket client in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn := startWebSocketClient(cfgProvider.Get().EventsubWebsocketUrl)
		if conn == nil {
			log.Println("Failed to establish WebSocket connection")
			close(shutdownChan)
			return
		}
		defer conn.Close()

		// Wait for shutdown signal
		<-shutdownChan
	}()

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
	fmt.Println("Shutdown complete")
}

func validateOAuthToken(cfg *config.Config) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
	if err != nil {
		log.Printf("error creating request: %v", err)
		return
	}

	req.Header.Add("Authorization", "OAuth "+cfg.OauthToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Response status: %d %s", resp.StatusCode, resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	log.Printf("Response body: %s", string(body))
}

func startWebSocketClient(url string) *websocket.Conn {
	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Error connecting to WebSocket: %v", err)
		return nil
	}

	fmt.Printf("WebSocket connection opened to %s\n", url)

	// Start a goroutine to handle incoming messages
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				return
			}

			var data interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}

			// Log the parsed data
			prettyJSON, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				log.Printf("Error formatting JSON for logging: %v", err)
			} else {
				log.Printf("Parsed message data:\n%s", string(prettyJSON))
			}

			handleWebSocketMessage(data)
		}
	}()

	return conn
}

// handleWebSocketMessage should be defined elsewhere in your code
func handleWebSocketMessage(data interface{}) {
	// Convert the generic interface{} to a map to access its fields
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		log.Println("Error: data is not a map")
		return
	}

	// Extract metadata
	metadata, ok := dataMap["metadata"].(map[string]interface{})
	if !ok {
		log.Println("Error: metadata is not a map")
		return
	}

	// Get message type
	messageType, ok := metadata["message_type"].(string)
	if !ok {
		log.Println("Error: message_type is not a string")
		return
	}

	switch messageType {
	case "session_welcome":
		// First message you get from the WebSocket server when connecting
		payload, ok := dataMap["payload"].(map[string]interface{})
		if !ok {
			log.Println("Error: payload is not a map")
			return
		}

		session, ok := payload["session"].(map[string]interface{})
		if !ok {
			log.Println("Error: session is not a map")
			return
		}

		sessionID, ok := session["id"].(string)
		if !ok {
			log.Println("Error: session ID is not a string")
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
			log.Println("Error: subscription_type is not a string")
			return
		}

		switch subscriptionType {
		case "channel.chat.message":
			payload, ok := dataMap["payload"].(map[string]interface{})
			if !ok {
				log.Println("Error: payload is not a map")
				return
			}

			event, ok := payload["event"].(map[string]interface{})
			if !ok {
				log.Println("Error: event is not a map")
				return
			}

			broadcasterUserLogin, _ := event["broadcaster_user_login"].(string)
			chatterUserLogin, _ := event["chatter_user_login"].(string)

			message, ok := event["message"].(map[string]interface{})
			if !ok {
				log.Println("Error: message is not a map")
				return
			}

			text, ok := message["text"].(string)
			if !ok {
				log.Println("Error: text is not a string")
				return
			}

			// First, print the message to the program's console
			log.Printf("MSG #%s <%s> %s", broadcasterUserLogin, chatterUserLogin, text)

			// Then check to see if that message was "HeyGuys"
			if strings.TrimSpace(text) == "HeyGuys" {
				// If so, send back "VoHiYo" to the chatroom
				sendChatMessage("VoHiYo")
			}
		}
	}
}

func sendChatMessage(chatMessage string) error {
	// Load configuration
	cfgProvider := config.MustNewConfigProvider()
	cfg := cfgProvider.Get()

	// Prepare the request body
	requestBody := map[string]string{
		"broadcaster_id": cfg.ChatChannelUserId,
		"sender_id":      cfg.BotUserId,
		"message":        chatMessage,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request body: %v", err)
		return err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/chat/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+cfg.OauthToken)
	req.Header.Set("Client-Id", cfg.ClientId)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to send chat message (Status: %d)", resp.StatusCode)
		log.Printf("Response: %s", string(body))
		return fmt.Errorf("failed to send chat message: status code %d", resp.StatusCode)
	}

	log.Printf("Sent chat message: %s", chatMessage)
	return nil
}

func registerEventSubListeners() error {
	// Load configuration
	cfgProvider := config.MustNewConfigProvider()
	cfg := cfgProvider.Get()

	fmt.Printf("%#v\n", cfg)
	// Create the request body
	requestBody := map[string]interface{}{
		"type":    "channel.chat.message",
		"version": "1",
		"condition": map[string]string{
			"broadcaster_user_id": cfg.ChatChannelUserId,
			"user_id":             cfg.BotUserId,
		},
		"transport": map[string]string{
			"method":     "websocket",
			"session_id": websocketSessionID,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request body: %v", err)
		return err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+cfg.OauthToken)
	req.Header.Set("Client-Id", cfg.ClientId)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return err
	}

	// Check response status (202 Accepted is the expected response)
	if resp.StatusCode != 202 {
		log.Printf("Failed to subscribe to channel.chat.message. API call returned status code %d", resp.StatusCode)
		log.Printf("Response: %s", string(body))
		return fmt.Errorf("failed to subscribe to channel.chat.message: status code %d", resp.StatusCode)
	}

	// Parse the response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		log.Printf("Error parsing response JSON: %v", err)
		return err
	}

	// Extract subscription ID
	dataArray, ok := responseData["data"].([]interface{})
	if !ok || len(dataArray) == 0 {
		log.Printf("Unexpected response format")
		return fmt.Errorf("unexpected response format")
	}

	firstItem, ok := dataArray[0].(map[string]interface{})
	if !ok {
		log.Printf("Unexpected data item format")
		return fmt.Errorf("unexpected data item format")
	}

	subscriptionID, ok := firstItem["id"].(string)
	if !ok {
		log.Printf("Could not find subscription ID")
		return fmt.Errorf("could not find subscription ID")
	}

	log.Printf("Subscribed to channel.chat.message [%s]", subscriptionID)
	return nil
}
