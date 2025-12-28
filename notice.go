package checkend

import (
	"time"
)

// Notice represents an error notice to be sent to Checkend.
type Notice struct {
	ErrorClass  string                 `json:"error_class"`
	Message     string                 `json:"message"`
	Backtrace   []string               `json:"backtrace"`
	Fingerprint string                 `json:"fingerprint,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Request     map[string]interface{} `json:"request,omitempty"`
	User        map[string]interface{} `json:"user,omitempty"`
	Environment string                 `json:"environment"`
	OccurredAt  time.Time              `json:"occurred_at"`
	Notifier    NotifierInfo           `json:"notifier"`
	AppName     string                 `json:"app_name,omitempty"`
	Revision    string                 `json:"revision,omitempty"`
	Hostname    string                 `json:"hostname,omitempty"`
}

// NotifierInfo contains SDK metadata.
type NotifierInfo struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Language        string `json:"language"`
	LanguageVersion string `json:"language_version"`
}

// ServerInfo contains server/application metadata.
type ServerInfo struct {
	AppName  string `json:"app_name,omitempty"`
	Revision string `json:"revision,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

// Payload represents the API request payload.
type Payload struct {
	Error    ErrorPayload           `json:"error"`
	Context  map[string]interface{} `json:"context"`
	Request  map[string]interface{} `json:"request,omitempty"`
	User     map[string]interface{} `json:"user,omitempty"`
	Notifier NotifierInfo           `json:"notifier"`
	Server   *ServerInfo            `json:"server,omitempty"`
}

// ErrorPayload represents the error portion of the payload.
type ErrorPayload struct {
	Class       string   `json:"class"`
	Message     string   `json:"message"`
	Backtrace   []string `json:"backtrace"`
	Fingerprint string   `json:"fingerprint,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	OccurredAt  string   `json:"occurred_at"`
}

// ToPayload converts the Notice to an API payload.
func (n *Notice) ToPayload() *Payload {
	ctx := make(map[string]interface{})
	ctx["environment"] = n.Environment
	for k, v := range n.Context {
		ctx[k] = v
	}

	payload := &Payload{
		Error: ErrorPayload{
			Class:       n.ErrorClass,
			Message:     n.Message,
			Backtrace:   n.Backtrace,
			Fingerprint: n.Fingerprint,
			Tags:        n.Tags,
			OccurredAt:  n.OccurredAt.UTC().Format(time.RFC3339),
		},
		Context:  ctx,
		Notifier: n.Notifier,
	}

	if len(n.Request) > 0 {
		payload.Request = n.Request
	}

	if len(n.User) > 0 {
		payload.User = n.User
	}

	// Include server info if any field is set
	if n.AppName != "" || n.Revision != "" || n.Hostname != "" {
		payload.Server = &ServerInfo{
			AppName:  n.AppName,
			Revision: n.Revision,
			Hostname: n.Hostname,
		}
	}

	return payload
}
