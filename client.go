package checkend

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// APIResponse represents the response from the Checkend API.
type APIResponse struct {
	ID        int `json:"id"`
	ProblemID int `json:"problem_id"`
}

// Client is the HTTP client for the Checkend API.
type Client struct {
	config     *Configuration
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(config *Configuration) *Client {
	return &Client{
		config:   config,
		endpoint: config.Endpoint + "/ingest/v1/errors",
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: buildTransport(config),
		},
	}
}

// buildTransport creates an HTTP transport with proxy, TLS, and timeout settings.
func buildTransport(config *Configuration) http.RoundTripper {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   config.ConnectTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Proxy configuration
	if config.Proxy != "" {
		proxyURL, err := url.Parse(config.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	// TLS configuration
	if !config.SSLVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // User explicitly disabled SSL verification
		}
	}

	return transport
}

// Send sends a notice to Checkend.
func (c *Client) Send(notice *Notice) *APIResponse {
	if c.config.APIKey == "" {
		c.log("error", "Cannot send notice: api_key not configured")
		return nil
	}

	payload := notice.ToPayload()
	data, err := json.Marshal(payload)
	if err != nil {
		c.log("error", fmt.Sprintf("Failed to marshal payload: %v", err))
		return nil
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewReader(data))
	if err != nil {
		c.log("error", fmt.Sprintf("Failed to create request: %v", err))
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Checkend-Ingestion-Key", c.config.APIKey)
	req.Header.Set("User-Agent", fmt.Sprintf("checkend-go/%s", Version))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log("error", fmt.Sprintf("Failed to send request: %v", err))
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log("error", fmt.Sprintf("Failed to read response: %v", err))
		return nil
	}

	if resp.StatusCode != http.StatusCreated {
		c.handleHTTPError(resp.StatusCode, body)
		return nil
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		c.log("error", fmt.Sprintf("Failed to parse response: %v", err))
		return nil
	}

	c.log("debug", fmt.Sprintf("Notice sent successfully: %+v", apiResp))
	return &apiResp
}

func (c *Client) handleHTTPError(statusCode int, body []byte) {
	switch statusCode {
	case http.StatusUnauthorized:
		c.log("error", "Authentication failed: invalid API key")
	case http.StatusUnprocessableEntity:
		c.log("error", fmt.Sprintf("Validation error: %s", string(body)))
	case http.StatusTooManyRequests:
		c.log("warning", "Rate limited by Checkend API")
	default:
		if statusCode >= 500 {
			c.log("error", fmt.Sprintf("Server error: %d", statusCode))
		} else {
			c.log("error", fmt.Sprintf("HTTP error: %d", statusCode))
		}
	}
}

func (c *Client) log(level, message string) {
	if !c.config.Debug && level == "debug" {
		return
	}
	fmt.Printf("[Checkend] [%s] %s\n", level, message)
}
