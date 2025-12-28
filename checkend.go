// Package checkend provides error monitoring for Go applications.
//
// Checkend is a lightweight, zero-dependency error monitoring SDK that sends
// errors to the Checkend service for tracking and analysis.
//
// Basic usage:
//
//	import "github.com/Checkend/checkend-go"
//
//	func main() {
//	    checkend.Configure(checkend.Config{
//	        APIKey: "your-api-key",
//	    })
//	    defer checkend.Stop()
//
//	    // Your application code...
//	    if err != nil {
//	        checkend.Notify(err)
//	    }
//	}
package checkend

import (
	"context"
	"sync"
)

// Version is the SDK version.
const Version = "0.1.0"

var (
	config      *Configuration
	worker      *Worker
	initialized bool
	mu          sync.RWMutex
)

// Configure initializes the Checkend SDK with the given configuration.
func Configure(cfg Config) *Configuration {
	mu.Lock()
	defer mu.Unlock()

	config = NewConfiguration(cfg)

	if config.AsyncSend && config.Enabled {
		worker = NewWorker(config)
		worker.Start()
	}

	initialized = true
	return config
}

// GetConfiguration returns the current configuration.
func GetConfiguration() *Configuration {
	mu.RLock()
	defer mu.RUnlock()
	return config
}

// Notify sends an error to Checkend asynchronously.
func Notify(err error, opts ...NotifyOption) {
	NotifyWithContext(context.Background(), err, opts...)
}

// NotifyWithContext sends an error to Checkend asynchronously with context.
func NotifyWithContext(ctx context.Context, err error, opts ...NotifyOption) {
	mu.RLock()
	defer mu.RUnlock()

	if !initialized || config == nil || !config.Enabled {
		return
	}

	// Check if error should be ignored
	if shouldIgnore(err) {
		return
	}

	// Build notice
	notice := buildNotice(ctx, err, opts...)

	// Run before notify callbacks
	if !runBeforeNotify(notice) {
		return
	}

	// Handle testing mode
	if testingEnabled {
		testingMu.Lock()
		testingNotices = append(testingNotices, notice)
		testingMu.Unlock()
		return
	}

	// Send asynchronously or synchronously
	if config.AsyncSend && worker != nil {
		worker.Push(notice)
	} else {
		client := NewClient(config)
		client.Send(notice)
	}
}

// NotifySync sends an error to Checkend synchronously and returns the response.
func NotifySync(err error, opts ...NotifyOption) *APIResponse {
	return NotifySyncWithContext(context.Background(), err, opts...)
}

// NotifySyncWithContext sends an error to Checkend synchronously with context.
func NotifySyncWithContext(ctx context.Context, err error, opts ...NotifyOption) *APIResponse {
	mu.RLock()
	defer mu.RUnlock()

	if !initialized || config == nil || !config.Enabled {
		return nil
	}

	// Check if error should be ignored
	if shouldIgnore(err) {
		return nil
	}

	// Build notice
	notice := buildNotice(ctx, err, opts...)

	// Run before notify callbacks
	if !runBeforeNotify(notice) {
		return nil
	}

	// Handle testing mode
	if testingEnabled {
		testingMu.Lock()
		testingNotices = append(testingNotices, notice)
		testingMu.Unlock()
		return &APIResponse{ID: 0, ProblemID: 0}
	}

	client := NewClient(config)
	return client.Send(notice)
}

// Flush waits for all queued notices to be sent.
func Flush() {
	mu.RLock()
	w := worker
	mu.RUnlock()

	if w != nil {
		w.Flush()
	}
}

// Stop stops the worker and waits for pending notices.
func Stop() {
	mu.Lock()
	defer mu.Unlock()

	if worker != nil {
		worker.Stop()
		worker = nil
	}
}

// Reset resets all state (useful for testing).
func Reset() {
	Stop()

	mu.Lock()
	config = nil
	initialized = false
	mu.Unlock()

	ClearTesting()
}

func shouldIgnore(err error) bool {
	if config == nil {
		return false
	}

	filter := NewIgnoreFilter(config.IgnoredErrors)
	return filter.ShouldIgnore(err)
}

func buildNotice(ctx context.Context, err error, opts ...NotifyOption) *Notice {
	options := &notifyOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Get context data
	ctxData := GetContextData(ctx)

	// Merge context
	mergedContext := make(map[string]interface{})
	for k, v := range ctxData.Context {
		mergedContext[k] = v
	}
	for k, v := range options.Context {
		mergedContext[k] = v
	}

	// Merge user
	mergedUser := ctxData.User
	if options.User != nil {
		mergedUser = options.User
	}

	// Merge request
	mergedRequest := ctxData.Request
	if options.Request != nil {
		mergedRequest = options.Request
	}

	builder := NewNoticeBuilder(config)
	return builder.Build(
		err,
		mergedContext,
		mergedUser,
		mergedRequest,
		options.Fingerprint,
		options.Tags,
	)
}

func runBeforeNotify(notice *Notice) bool {
	if config == nil || len(config.BeforeNotify) == 0 {
		return true
	}

	for _, callback := range config.BeforeNotify {
		if !callback(notice) {
			return false
		}
	}

	return true
}

// NotifyOption is a functional option for Notify.
type NotifyOption func(*notifyOptions)

type notifyOptions struct {
	Context     map[string]interface{}
	User        map[string]interface{}
	Request     map[string]interface{}
	Fingerprint string
	Tags        []string
}

// WithContext sets additional context data.
func WithContext(ctx map[string]interface{}) NotifyOption {
	return func(o *notifyOptions) {
		o.Context = ctx
	}
}

// WithUser sets user information.
func WithUser(user map[string]interface{}) NotifyOption {
	return func(o *notifyOptions) {
		o.User = user
	}
}

// WithRequest sets request information.
func WithRequest(request map[string]interface{}) NotifyOption {
	return func(o *notifyOptions) {
		o.Request = request
	}
}

// WithFingerprint sets a custom fingerprint for error grouping.
func WithFingerprint(fingerprint string) NotifyOption {
	return func(o *notifyOptions) {
		o.Fingerprint = fingerprint
	}
}

// WithTags sets tags for the error.
func WithTags(tags ...string) NotifyOption {
	return func(o *notifyOptions) {
		o.Tags = tags
	}
}
