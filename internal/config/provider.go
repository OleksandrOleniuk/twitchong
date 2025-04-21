package config

import (
	"log"
)

// Provider holds the application configuration
type Provider struct {
	cfg *Config
}

// NewConfigProvider creates a new config provider
func NewConfigProvider() (*Provider, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	return &Provider{cfg: cfg}, nil
}

// MustNewConfigProvider creates a new config provider and panics if there's an error
func MustNewConfigProvider() *Provider {
	provider, err := NewConfigProvider()
	if err != nil {
		log.Fatalf("Failed to create config provider: %v", err)
	}
	return provider
}

// Get returns the configuration
func (p *Provider) Get() *Config {
	return p.cfg
} 