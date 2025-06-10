package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
	"go.uber.org/zap"
)

func NewTwitchChat(appConfig *config.Config) *Client {
	client := New(appConfig,
		WithReconnectDelay(time.Second*3),
		WithOnConnect(func() {
			logger.Info("Connected to WebSocket server!")
		}),
		WithOnDisconnect(func(err error) {
			if err != nil {
				logger.Error("Disconnected with error")
			} else {
				logger.Info("Disconnected from WebSocket server")
			}
		}),
	)

	client.HandleWelcome()

	client.HandleMessage(func(text string) {
		if strings.Contains(text, "HeyGuys") {
			sendChatMessage(client.appConfig, "VoHiYo")
		}
	})

	// client.HandleMessage(func(text string) {
	// 	if strings.HasPrefix(text, "@WayongBotJr") {
	// 		sendChatMessage(client.appConfig, "???")
	// 	}
	// })

	client.HandleMessage(func(text string) {
		if strings.HasPrefix(text, "@WayongBotJr") {
			prompt := strings.Split(text, "@WayongBotJr ")[1]
			fmt.Printf("prompt: %v\n", prompt)
			config := utils.RequestConfig{
				Method: "POST",
				URL:    "http://localhost:11434/api/generate",
				Body: map[string]any{
					"model":  "gemma3:1b",
					"prompt": "My question to you is: \"" + prompt + "\". Respond to it with maximum 50 words in Ukrainian language. Do not ask questions. Do not apologize. Do not quote yourself.",
					"stream": false,
				},
			}

			var res struct {
				Response string `json:"response"`
			}

			err := utils.SendRequestAndParseResponse(config, &res)
			if err != nil {
				logger.Error("token validation failed", zap.Error(err))
				return
			}

			fmt.Printf("res.Response: %v\n", res.Response)

			sendChatMessage(client.appConfig, res.Response)

		}
	})

	// Default handler for unmatched message types
	client.HandleDefault(func(ctx context.Context, data map[string]any) error {
		// advanced logger
		// utils.PrettyObject("Unhandled ws message", "data", data)
		return nil
	})

// 	ticker:=time.NewTicker(60000)
// 	go func(c *Client)() {
// defer ticker.Stop()
//
// 	}
//
// 	return client
}

func StartTwitchChat(client *Client) error {
	if err := client.Start(); err != nil {
		logger.Error("Failed to establish WebSocket connection")
	}
	return nil
}

func sendChatMessage(appConfig *config.Config, chatMessage string) error {
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
