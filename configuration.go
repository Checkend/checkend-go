package checkend

import (
	"os"
	"strings"
	"time"
)

// DefaultEndpoint is the default Checkend API endpoint.
const DefaultEndpoint = "https://app.checkend.io"

// DefaultTimeout is the default HTTP request timeout.
const DefaultTimeout = 15 * time.Second

// DefaultConnectTimeout is the default connection establishment timeout.
const DefaultConnectTimeout = 5 * time.Second

// DefaultShutdownTimeout is the default graceful shutdown timeout.
const DefaultShutdownTimeout = 5 * time.Second

// DefaultMaxQueueSize is the default maximum queue size for async sending.
const DefaultMaxQueueSize = 1000

// DefaultFilterKeys are the default keys to filter from payloads.
var DefaultFilterKeys = []string{
	"password",
	"password_confirmation",
	"secret",
	"secret_key",
	"api_key",
	"apikey",
	"access_token",
	"auth_token",
	"authorization",
	"token",
	"credit_card",
	"card_number",
	"cvv",
	"cvc",
	"ssn",
	"social_security",
}

// Config holds the configuration options for Checkend.
type Config struct {
	// APIKey is your Checkend ingestion API key (required).
	APIKey string

	// Endpoint is the API endpoint URL.
	Endpoint string

	// Environment is the environment name (e.g., "production", "staging").
	Environment string

	// Enabled controls whether error reporting is active.
	Enabled *bool

	// AsyncSend controls whether errors are sent asynchronously.
	AsyncSend bool

	// MaxQueueSize is the maximum queue size for async sending.
	MaxQueueSize int

	// Timeout is the HTTP request timeout.
	Timeout time.Duration

	// ConnectTimeout is the connection establishment timeout.
	ConnectTimeout time.Duration

	// ShutdownTimeout is the graceful shutdown timeout.
	ShutdownTimeout time.Duration

	// FilterKeys are additional keys to filter from payloads.
	FilterKeys []string

	// IgnoredErrors are error types or patterns to ignore.
	IgnoredErrors []interface{}

	// BeforeNotify are callbacks to run before sending a notice.
	// Return false to skip sending.
	BeforeNotify []func(*Notice) bool

	// Debug enables debug logging.
	Debug bool

	// AppName is the application identifier.
	AppName string

	// Revision is the code revision or commit hash.
	Revision string

	// RootPath is the application root path for cleaning backtraces.
	RootPath string

	// SendRequestData controls whether request data is included in notices.
	SendRequestData *bool

	// SendSessionData controls whether session data is included in notices.
	SendSessionData *bool

	// SendEnvironment controls whether environment variables are included in notices.
	SendEnvironment *bool

	// SendUserData controls whether user data is included in notices.
	SendUserData *bool

	// Proxy is the HTTP proxy URL.
	Proxy string

	// SSLVerify controls TLS certificate verification.
	SSLVerify *bool
}

// Configuration is the resolved configuration for the SDK.
type Configuration struct {
	APIKey          string
	Endpoint        string
	Environment     string
	Enabled         bool
	AsyncSend       bool
	MaxQueueSize    int
	Timeout         time.Duration
	ConnectTimeout  time.Duration
	ShutdownTimeout time.Duration
	FilterKeys      []string
	IgnoredErrors   []interface{}
	BeforeNotify    []func(*Notice) bool
	Debug           bool
	AppName         string
	Revision        string
	RootPath        string
	SendRequestData bool
	SendSessionData bool
	SendEnvironment bool
	SendUserData    bool
	Proxy           string
	SSLVerify       bool
}

// NewConfiguration creates a new Configuration from Config.
func NewConfiguration(cfg Config) *Configuration {
	c := &Configuration{
		APIKey:          cfg.APIKey,
		AsyncSend:       true,
		MaxQueueSize:    DefaultMaxQueueSize,
		Timeout:         DefaultTimeout,
		ConnectTimeout:  DefaultConnectTimeout,
		ShutdownTimeout: DefaultShutdownTimeout,
		FilterKeys:      append([]string{}, DefaultFilterKeys...),
		IgnoredErrors:   cfg.IgnoredErrors,
		BeforeNotify:    cfg.BeforeNotify,
		Debug:           cfg.Debug,
		SendRequestData: true,
		SendSessionData: true,
		SendEnvironment: false,
		SendUserData:    true,
		SSLVerify:       true,
	}

	// API key from environment
	if c.APIKey == "" {
		c.APIKey = os.Getenv("CHECKEND_API_KEY")
	}

	// Endpoint
	c.Endpoint = cfg.Endpoint
	if c.Endpoint == "" {
		c.Endpoint = os.Getenv("CHECKEND_ENDPOINT")
	}
	if c.Endpoint == "" {
		c.Endpoint = DefaultEndpoint
	}

	// Environment
	c.Environment = cfg.Environment
	if c.Environment == "" {
		c.Environment = os.Getenv("CHECKEND_ENVIRONMENT")
	}
	if c.Environment == "" {
		c.Environment = detectEnvironment()
	}

	// Enabled
	if cfg.Enabled != nil {
		c.Enabled = *cfg.Enabled
	} else {
		c.Enabled = c.Environment == "production" || c.Environment == "staging"
	}

	// AsyncSend - use explicit value only if set, otherwise keep default (true)
	// Note: Due to Go's zero values, we use a simple heuristic:
	// if APIKey is provided and AsyncSend is explicitly false, respect it
	// For guaranteed sync behavior, use NotifySync() instead
	if cfg.AsyncSend {
		c.AsyncSend = true
	}

	// MaxQueueSize
	if cfg.MaxQueueSize > 0 {
		c.MaxQueueSize = cfg.MaxQueueSize
	}

	// Timeout
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}

	// ConnectTimeout
	if cfg.ConnectTimeout > 0 {
		c.ConnectTimeout = cfg.ConnectTimeout
	}

	// ShutdownTimeout
	if cfg.ShutdownTimeout > 0 {
		c.ShutdownTimeout = cfg.ShutdownTimeout
	}

	// FilterKeys
	c.FilterKeys = append(c.FilterKeys, cfg.FilterKeys...)

	// Debug from environment
	if !c.Debug {
		debugEnv := strings.ToLower(os.Getenv("CHECKEND_DEBUG"))
		c.Debug = debugEnv == "true" || debugEnv == "1" || debugEnv == "yes"
	}

	// AppName
	c.AppName = cfg.AppName
	if c.AppName == "" {
		c.AppName = os.Getenv("CHECKEND_APP_NAME")
	}

	// Revision
	c.Revision = cfg.Revision
	if c.Revision == "" {
		c.Revision = os.Getenv("CHECKEND_REVISION")
	}
	if c.Revision == "" {
		c.Revision = os.Getenv("GIT_COMMIT")
	}

	// RootPath
	c.RootPath = cfg.RootPath
	if c.RootPath == "" {
		c.RootPath = os.Getenv("CHECKEND_ROOT_PATH")
	}

	// SendRequestData (default true, explicit false overrides)
	if cfg.SendRequestData != nil {
		c.SendRequestData = *cfg.SendRequestData
	}

	// SendSessionData (default true, explicit false overrides)
	if cfg.SendSessionData != nil {
		c.SendSessionData = *cfg.SendSessionData
	}

	// SendEnvironment (default false, explicit true overrides)
	if cfg.SendEnvironment != nil {
		c.SendEnvironment = *cfg.SendEnvironment
	}

	// SendUserData (default true, explicit false overrides)
	if cfg.SendUserData != nil {
		c.SendUserData = *cfg.SendUserData
	}

	// Proxy
	c.Proxy = cfg.Proxy
	if c.Proxy == "" {
		c.Proxy = os.Getenv("HTTPS_PROXY")
	}
	if c.Proxy == "" {
		c.Proxy = os.Getenv("HTTP_PROXY")
	}

	// SSLVerify (default true, explicit false overrides)
	if cfg.SSLVerify != nil {
		c.SSLVerify = *cfg.SSLVerify
	} else {
		sslVerifyEnv := strings.ToLower(os.Getenv("CHECKEND_SSL_VERIFY"))
		if sslVerifyEnv == "false" || sslVerifyEnv == "0" || sslVerifyEnv == "no" {
			c.SSLVerify = false
		}
	}

	return c
}

func detectEnvironment() string {
	envVars := []string{
		"GO_ENV",
		"ENVIRONMENT",
		"ENV",
		"APP_ENV",
	}

	for _, v := range envVars {
		if val := os.Getenv(v); val != "" {
			return val
		}
	}

	return "development"
}
