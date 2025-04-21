// internal/config/config.go
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	ServerPort           int
	LogLevel             string
	Environment          string
	BotUserId            string
	OauthToken           string
	ClientId             string
	ClientSecret         string
	ChatChannelUserId    string
	EventsubWebsocketUrl string
	TwitchSecretState    string
}

// Load returns app configuration from .env file and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found. Using environment variables only.")
	}

	// Server config
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, err
	}

	return &Config{
		// Server configs
		ServerPort:           port,
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		Environment:          getEnv("APP_ENV", "development"),
		BotUserId:            getEnv("BOT_USER_ID", "undefined"),
		OauthToken:           getEnv("OAUTH_TOKEN", "undefined"),
		ClientId:             getEnv("CLIENT_ID", "undefined"),
		ClientSecret:         getEnv("CLIENT_SECRET", "undefined"),
		ChatChannelUserId:    getEnv("CHAT_CHANNEL_USER_ID", "undefined"),
		EventsubWebsocketUrl: getEnv("EVENTSUB_WEBSOCKET_URL", "undefined"),
		TwitchSecretState:    getEnv("TWITCH_SECRET_STATE", "undefined"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
