package config

import (
	"errors"
	"time"
)

// Config holds the configuration for XTS client
type Config struct {
	SecretKey string
	AppKey    string
	ClientID  string // Required for dealer accounts

	BaseURL           string
	MarketDataBaseURL string

	HostLookupURL  string
	AccessPassword string
	Version        string

	Timeout       time.Duration
	RetryAttempts int

	DisableSSL bool

	Source string

	BroadcastMode string

	Debug bool
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		BaseURL:           "https://developers.symphonyfintech.in",
		MarketDataBaseURL: "https://developers.symphonyfintech.in",
		HostLookupURL:     "https://developers.symphonyfintech.in",
		Source:            "WEBAPI",
		BroadcastMode:     "Full",
		Version:           "interactiveapi_1.0.1",
		Timeout:           30 * time.Second,
		RetryAttempts:     3,
		DisableSSL:        true,
		Debug:             false,
	}
}

// NewConfig creates a new configuration with the provided credentials
func NewConfig(secretKey, appKey, clientID string) *Config {
	config := DefaultConfig()
	config.SecretKey = secretKey
	config.AppKey = appKey
	config.ClientID = clientID
	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.SecretKey == "" {
		return errors.New("secret key is required")
	}
	if c.AppKey == "" {
		return errors.New("app key is required")
	}
	if c.BaseURL == "" {
		return errors.New("base URL is required")
	}
	if c.Source == "" {
		return errors.New("source is required")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	return nil
}

// SetEnvironment sets configuration for different environments
func (c *Config) SetEnvironment(env string) {
	switch env {
	case "production":
		c.BaseURL = "https://xts-api.trading"
		c.MarketDataBaseURL = "https://xts-api.trading"
		c.HostLookupURL = "https://xts-api.trading"
		c.DisableSSL = false
	case "development":
		c.BaseURL = "https://developers.symphonyfintech.in"
		c.MarketDataBaseURL = "https://developers.symphonyfintech.in"
		c.HostLookupURL = "https://developers.symphonyfintech.in"
		c.DisableSSL = true
	case "sandbox":
		c.BaseURL = "https://xts-api.trading"
		c.MarketDataBaseURL = "https://xts-api.trading"
		c.HostLookupURL = "https://xts-api.trading"
		c.DisableSSL = true
	}
}

// Clone creates a copy of the configuration
func (c *Config) Clone() *Config {
	return &Config{
		SecretKey:         c.SecretKey,
		AppKey:            c.AppKey,
		ClientID:          c.ClientID,
		BaseURL:           c.BaseURL,
		MarketDataBaseURL: c.MarketDataBaseURL,
		HostLookupURL:     c.HostLookupURL,
		AccessPassword:    c.AccessPassword,
		Version:           c.Version,
		Timeout:           c.Timeout,
		RetryAttempts:     c.RetryAttempts,
		DisableSSL:        c.DisableSSL,
		Source:            c.Source,
		BroadcastMode:     c.BroadcastMode,
		Debug:             c.Debug,
	}
}
