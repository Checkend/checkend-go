package checkend

import (
	"os"
	"testing"
	"time"
)

func TestConfigurationDefaults(t *testing.T) {
	cfg := NewConfiguration(Config{APIKey: "test-key"})

	if cfg.Endpoint != DefaultEndpoint {
		t.Errorf("Expected endpoint '%s', got '%s'", DefaultEndpoint, cfg.Endpoint)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, cfg.Timeout)
	}

	if cfg.MaxQueueSize != DefaultMaxQueueSize {
		t.Errorf("Expected max queue size %d, got %d", DefaultMaxQueueSize, cfg.MaxQueueSize)
	}

	if !cfg.AsyncSend {
		t.Error("Expected AsyncSend to be true by default")
	}
}

func TestConfigurationAPIKeyFromEnv(t *testing.T) {
	os.Setenv("CHECKEND_API_KEY", "env-key")
	defer os.Unsetenv("CHECKEND_API_KEY")

	cfg := NewConfiguration(Config{})

	if cfg.APIKey != "env-key" {
		t.Errorf("Expected API key 'env-key', got '%s'", cfg.APIKey)
	}
}

func TestConfigurationParameterOverridesEnv(t *testing.T) {
	os.Setenv("CHECKEND_API_KEY", "env-key")
	defer os.Unsetenv("CHECKEND_API_KEY")

	cfg := NewConfiguration(Config{APIKey: "param-key"})

	if cfg.APIKey != "param-key" {
		t.Errorf("Expected API key 'param-key', got '%s'", cfg.APIKey)
	}
}

func TestConfigurationEndpointFromEnv(t *testing.T) {
	os.Setenv("CHECKEND_ENDPOINT", "https://custom.example.com")
	defer os.Unsetenv("CHECKEND_ENDPOINT")

	cfg := NewConfiguration(Config{APIKey: "test-key"})

	if cfg.Endpoint != "https://custom.example.com" {
		t.Errorf("Expected endpoint 'https://custom.example.com', got '%s'", cfg.Endpoint)
	}
}

func TestConfigurationEnvironmentDetection(t *testing.T) {
	// Test default
	cfg := NewConfiguration(Config{APIKey: "test-key"})
	if cfg.Environment != "development" {
		t.Errorf("Expected environment 'development', got '%s'", cfg.Environment)
	}

	// Test from env var
	os.Setenv("GO_ENV", "production")
	defer os.Unsetenv("GO_ENV")

	cfg = NewConfiguration(Config{APIKey: "test-key"})
	if cfg.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", cfg.Environment)
	}
}

func TestConfigurationEnabledInProduction(t *testing.T) {
	os.Setenv("GO_ENV", "production")
	defer os.Unsetenv("GO_ENV")

	cfg := NewConfiguration(Config{APIKey: "test-key"})

	if !cfg.Enabled {
		t.Error("Expected Enabled to be true in production")
	}
}

func TestConfigurationDisabledInDevelopment(t *testing.T) {
	cfg := NewConfiguration(Config{APIKey: "test-key"})

	if cfg.Enabled {
		t.Error("Expected Enabled to be false in development")
	}
}

func TestConfigurationExplicitEnabled(t *testing.T) {
	enabled := true
	cfg := NewConfiguration(Config{
		APIKey:  "test-key",
		Enabled: &enabled,
	})

	if !cfg.Enabled {
		t.Error("Expected Enabled to be true when explicitly set")
	}
}

func TestConfigurationFilterKeys(t *testing.T) {
	cfg := NewConfiguration(Config{
		APIKey:     "test-key",
		FilterKeys: []string{"custom_key"},
	})

	// Should include defaults
	hasPassword := false
	hasCustom := false
	for _, k := range cfg.FilterKeys {
		if k == "password" {
			hasPassword = true
		}
		if k == "custom_key" {
			hasCustom = true
		}
	}

	if !hasPassword {
		t.Error("Expected default filter key 'password'")
	}
	if !hasCustom {
		t.Error("Expected custom filter key 'custom_key'")
	}
}

func TestConfigurationCustomTimeout(t *testing.T) {
	cfg := NewConfiguration(Config{
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	})

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cfg.Timeout)
	}
}

func TestConfigurationCustomMaxQueueSize(t *testing.T) {
	cfg := NewConfiguration(Config{
		APIKey:       "test-key",
		MaxQueueSize: 500,
	})

	if cfg.MaxQueueSize != 500 {
		t.Errorf("Expected max queue size 500, got %d", cfg.MaxQueueSize)
	}
}

func TestConfigurationDebugFromEnv(t *testing.T) {
	os.Setenv("CHECKEND_DEBUG", "true")
	defer os.Unsetenv("CHECKEND_DEBUG")

	cfg := NewConfiguration(Config{APIKey: "test-key"})

	if !cfg.Debug {
		t.Error("Expected Debug to be true from env")
	}
}
