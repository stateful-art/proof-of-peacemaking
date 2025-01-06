package config

import (
	"os"

	"github.com/joho/godotenv"
)

type LiveKitConfig struct {
	Host      string
	APIKey    string
	APISecret string
}

func NewLiveKitConfig() (*LiveKitConfig, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &LiveKitConfig{
		Host:      os.Getenv("LIVEKIT_HOST"),
		APIKey:    os.Getenv("LIVEKIT_API_KEY"),
		APISecret: os.Getenv("LIVEKIT_API_SECRET"),
	}

	return config, nil
}
