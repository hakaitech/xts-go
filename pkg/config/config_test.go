package config

import (
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	secretKey := "test-secret"
	appKey := "test-app-key"
	clientID := "test-client-id"

	cfg := NewConfig(secretKey, appKey, clientID)

	if cfg.SecretKey != secretKey {
		t.Errorf("Expected SecretKey %s, got %s", secretKey, cfg.SecretKey)
	}
	if cfg.AppKey != appKey {
		t.Errorf("Expected AppKey %s, got %s", appKey, cfg.AppKey)
	}
	if cfg.ClientID != clientID {
		t.Errorf("Expected ClientID %s, got %s", clientID, cfg.ClientID)
	}
	if cfg.Source != "WEBAPI" {
		t.Errorf("Expected default Source WEBAPI, got %s", cfg.Source)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.BaseURL == "" {
		t.Error("Expected non-empty BaseURL")
	}
	if cfg.Source != "WEBAPI" {
		t.Errorf("Expected Source WEBAPI, got %s", cfg.Source)
	}
	if cfg.BroadcastMode != "Full" {
		t.Errorf("Expected BroadcastMode Full, got %s", cfg.BroadcastMode)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %s", cfg.Timeout)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		expectErr bool
	}{
		{
			name:      "Valid config",
			cfg:       NewConfig("secret", "appkey", "client"),
			expectErr: false,
		},
		{
			name:      "Missing secret key",
			cfg:       NewConfig("", "appkey", "client"),
			expectErr: true,
		},
		{
			name:      "Missing app key",
			cfg:       NewConfig("secret", "", "client"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSetEnvironment(t *testing.T) {
	cfg := DefaultConfig()

	cfg.SetEnvironment("production")
	if cfg.BaseURL != "https://xts-api.trading" {
		t.Errorf("Expected production URL, got %s", cfg.BaseURL)
	}
	if cfg.DisableSSL {
		t.Error("Expected SSL to be enabled in production")
	}

	cfg.SetEnvironment("development")
	if cfg.BaseURL != "https://developers.symphonyfintech.in" {
		t.Errorf("Expected development URL, got %s", cfg.BaseURL)
	}
	if !cfg.DisableSSL {
		t.Error("Expected SSL to be disabled in development")
	}
}

func TestConfigClone(t *testing.T) {
	original := NewConfig("secret", "appkey", "client")
	original.Debug = true
	original.Timeout = 60 * time.Second

	cloned := original.Clone()

	if cloned.SecretKey != original.SecretKey {
		t.Error("Clone should have same SecretKey")
	}
	if cloned.Debug != original.Debug {
		t.Error("Clone should have same Debug setting")
	}
	if cloned.Timeout != original.Timeout {
		t.Error("Clone should have same Timeout")
	}

	// Verify it's a real clone, not the same reference
	cloned.Debug = false
	if original.Debug == cloned.Debug {
		t.Error("Modifying clone should not affect original")
	}
}