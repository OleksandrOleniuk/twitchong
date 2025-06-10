package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var logger = utils.With(zap.String("component", "websocket"))

// MessageType represents different types of messages that can be handled
type MessageType string

// HandlerFunc defines the function signature for message handlers
type HandlerFunc func(ctx context.Context, data map[string]any) error

// Client represents a WebSocket client connection
type Client struct {
	appConfig        *config.Config
	conn             *websocket.Conn
	wsSessionId      string
	wsSubscriptionId string
	handlers         map[MessageType]HandlerFunc
	defaultHandler   HandlerFunc
	welcomeHandler   HandlerFunc
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	reconnectDelay   time.Duration
	connected        bool
	autoReconnect    bool
	onConnectFunc    func()
	onDisconnectFunc func(error)
}

// ClientOption defines functional options for configuring the Client
type ClientOption func(*Client)

// New creates a new WebSocket client with the provided URL
func New(config *config.Config, options ...ClientOption) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		handlers:       make(map[MessageType]HandlerFunc),
		ctx:            ctx,
		cancel:         cancel,
		reconnectDelay: time.Second * 5,
		autoReconnect:  true,
		appConfig:      config,
	}

	// Apply all options
	for _, option := range options {
		option(client)
	}

	return client
}

// WithReconnectDelay sets the delay between reconnection attempts
func WithReconnectDelay(delay time.Duration) ClientOption {
	return func(c *Client) {
		c.reconnectDelay = delay
	}
}

// WithAutoReconnect enables or disables automatic reconnection
func WithAutoReconnect(enable bool) ClientOption {
	return func(c *Client) {
		c.autoReconnect = enable
	}
}

// WithOnConnect sets a function to be called when a connection is established
func WithOnConnect(fn func()) ClientOption {
	return func(c *Client) {
		c.onConnectFunc = fn
	}
}

// WithOnDisconnect sets a function to be called when a connection is closed
func WithOnDisconnect(fn func(error)) ClientOption {
	return func(c *Client) {
		c.onDisconnectFunc = fn
	}
}

// Handle registers a handler for a specific message type
func (c *Client) Handle(msgType MessageType, handler HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[msgType] = handler
}

// Handle registers a handler for a specific message type
func (c *Client) HandleMessage(handler func(text string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers["notification"] = func(ctx context.Context, data map[string]any) error {
		// Extract metadata
		metadata, ok := data["metadata"].(map[string]any)
		if !ok {
			logger.Error("metadata is not a map")
			return nil
		}

		// An EventSub notification has occurred, such as channel.chat.message
		subscriptionType, ok := metadata["subscription_type"].(string)
		if !ok {
			logger.Error("subscription_type is not a string")
			return nil
		}

		switch subscriptionType {
		case "channel.chat.message":
			payload, ok := data["payload"].(map[string]any)
			if !ok {
				logger.Error("payload is not a map")
				return nil
			}

			event, ok := payload["event"].(map[string]any)
			if !ok {
				logger.Error("event is not a map")
				return nil
			}

			broadcasterUserLogin, _ := event["broadcaster_user_login"].(string)
			chatterUserLogin, _ := event["chatter_user_login"].(string)

			message, ok := event["message"].(map[string]any)
			if !ok {
				logger.Error("message is not a map")
				return nil
			}

			text, ok := message["text"].(string)
			if !ok {
				logger.Error("text is not a string")
				return nil
			}

			// First, print the message to the program's console
			logger.Info("chat message received",
				zap.String("channel", broadcasterUserLogin),
				zap.String("user", chatterUserLogin),
				zap.String("message", text),
			)

			handler(text)
		}

		return nil
	}
}

// // Handle registers a handler for a specific message type
// func (c *Client) HandleReply(msgType MessageType, handler HandlerFunc) {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()
// 	c.handlers[msgType] = handler
// }

// Handle welcome message
func (c *Client) HandleWelcome() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.welcomeHandler = func(ctx context.Context, data map[string]any) error {
		// First message you get from the WebSocket server when connecting
		payload, ok := data["payload"].(map[string]interface{})
		if !ok {
			logger.Error("payload is not a map")
			return nil
		}

		fmt.Printf("payload: %v\n", payload)

		session, ok := payload["session"].(map[string]interface{})
		if !ok {
			logger.Error("session is not a map")
			return nil
		}

		sessionID, ok := session["id"].(string)
		if !ok {
			logger.Error("session ID is not a string")
			return nil
		}

		// Register the Session ID it gives us
		c.wsSessionId = sessionID

		// Listen to EventSub, which joins the chatroom from your bot's account
		registerEventSubListeners(c)

		return nil
	}
}

// HandleDefault sets a default handler for messages that don't match any specific type
func (c *Client) HandleDefault(handler HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultHandler = handler
}

// Start initiates the WebSocket connection and begins processing messages
func (c *Client) Start() error {
	err := c.connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Start goroutines for reading and writing
	go c.readPump()

	return nil
}

// IsConnected returns the current connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Stop closes the WebSocket connection and stops all goroutines
func (c *Client) Stop() {
	c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}
}

// connect establishes a WebSocket connection
func (c *Client) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.appConfig.EventsubWebsocketUrl, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.mu.Unlock()

	if c.onConnectFunc != nil {
		c.onConnectFunc()
	}

	return nil
}

// reconnect attempts to reconnect to the WebSocket server
func (c *Client) reconnect() {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			log.Printf("Attempting to reconnect to %s in %v", c.appConfig.EventsubWebsocketUrl, c.reconnectDelay)
			time.Sleep(c.reconnectDelay)

			err := c.connect()
			if err == nil {
				log.Printf("Successfully reconnected to %s", c.appConfig.EventsubWebsocketUrl)
				go c.readPump()
				return
			}

			log.Printf("Failed to reconnect: %v", err)
		}
	}
}

// readPump handles reading messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.conn.Close()
		c.mu.Unlock()

		if c.onDisconnectFunc != nil {
			c.onDisconnectFunc(nil)
		}

		if c.autoReconnect {
			go c.reconnect()
		}
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			}
			break
		}

		var data any
		if err := json.Unmarshal(message, &data); err != nil {
			logger.Error("error parsing message", zap.Error(err))
			continue
		}

		go c.handleMessage(data)
	}
}

// handleMessage processes incoming messages and routes them to the appropriate handler
func (c *Client) handleMessage(data any) {
	// Convert the generic interface{} to a map to access its fields
	dataMap, ok := data.(map[string]any)
	if !ok {
		logger.Error("data is not a map")
		return
	}

	// Extract metadata
	metadata, ok := dataMap["metadata"].(map[string]any)
	if !ok {
		logger.Error("metadata is not a map")
		return
	}

	// Get message type
	metadataMsgType, ok := metadata["message_type"].(string)
	if !ok {
		logger.Error("message_type is not a string")
		return
	}

	msgType := MessageType(metadataMsgType)

	// Find the appropriate handler
	c.mu.RLock()
	handler, ok := c.handlers[msgType]
	c.mu.RUnlock()

	if ok {
		// We have a specific handler for this message type
		err := handler(c.ctx, dataMap)
		if err != nil {
			log.Printf("Error handling message of type %s: %v", msgType, err)
		}
	} else if c.welcomeHandler != nil && c.wsSessionId == "" {
		c.welcomeHandler(c.ctx, dataMap)
	} else if c.defaultHandler != nil {
		// Use the default handler
		c.defaultHandler(c.ctx, dataMap)
	} else {
		log.Printf("No handler found for message type: %s", msgType)
	}
}

func registerEventSubListeners(c *Client) (string, error) {
	// Create the request body
	requestBody := map[string]interface{}{
		"type":    "channel.chat.message",
		"version": "1",
		"condition": map[string]string{
			"broadcaster_user_id": c.appConfig.ChatChannelUserId,
			"user_id":             c.appConfig.BotUserId,
		},
		"transport": map[string]string{
			"method":     "websocket",
			"session_id": c.wsSessionId,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("error marshaling request body", zap.Error(err))
		return "", err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error("error creating request", zap.Error(err))
		return "", err
	}

	fmt.Printf("requestBody: %v\n", requestBody)

	// Add headers
	req.Header.Set("Authorization", "Bearer "+c.appConfig.OauthToken)
	req.Header.Set("Client-Id", c.appConfig.ClientId)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("error sending request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("error reading response body", zap.Error(err))
		return "", err
	}

	// Check response status (202 Accepted is the expected response)
	if resp.StatusCode != 202 {
		logger.Error("failed to subscribe to channel.chat.message",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return "", fmt.Errorf("failed to subscribe to channel.chat.message: status code %d", resp.StatusCode)
	}

	// Parse the response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		logger.Error("error parsing response JSON", zap.Error(err))
		return "", err
	}

	// Extract subscription ID
	dataArray, ok := responseData["data"].([]interface{})
	if !ok || len(dataArray) == 0 {
		logger.Error("unexpected response format")
		return "", fmt.Errorf("unexpected response format")
	}

	firstItem, ok := dataArray[0].(map[string]interface{})
	if !ok {
		logger.Error("unexpected data item format")
		return "", fmt.Errorf("unexpected data item format")
	}

	subscriptionID, ok := firstItem["id"].(string)
	if !ok {
		logger.Error("could not find subscription ID")
		return "", fmt.Errorf("could not find subscription ID")
	}

	return subscriptionID, nil
}
