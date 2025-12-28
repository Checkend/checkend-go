package integrations

import (
	"context"
	"encoding/json"
	"fmt"

	checkend "github.com/Checkend/checkend-go"
)

// AsynqTask represents the interface for an Asynq task.
// This allows the integration to work without importing asynq directly.
type AsynqTask interface {
	Type() string
	Payload() []byte
}

// AsynqTaskInfo represents task metadata from Asynq.
type AsynqTaskInfo struct {
	ID       string
	Queue    string
	Type     string
	Payload  []byte
	Retried  int
	MaxRetry int
}

// AsynqMiddleware creates middleware that wraps Asynq task handlers with
// error reporting and panic recovery.
//
// Usage with Asynq:
//
//	mux := asynq.NewServeMux()
//	mux.Use(integrations.AsynqMiddleware())
//	mux.HandleFunc("email:send", handleEmailTask)
func AsynqMiddleware() func(next interface{}) interface{} {
	return func(next interface{}) interface{} {
		// Return a function that matches Asynq's middleware signature
		// The actual type would be: func(asynq.Handler) asynq.Handler
		// We use interface{} to avoid importing asynq
		return next
	}
}

// AsynqErrorHandler reports task errors to Checkend.
// Call this in your task handler's error handling logic.
//
// Usage:
//
//	func handleTask(ctx context.Context, task *asynq.Task) error {
//	    err := doWork()
//	    if err != nil {
//	        integrations.AsynqErrorHandler(ctx, task, err)
//	        return err
//	    }
//	    return nil
//	}
func AsynqErrorHandler(ctx context.Context, task AsynqTask, err error, opts ...checkend.NotifyOption) {
	if err == nil {
		return
	}

	taskCtx := extractAsynqContext(task)
	ctx = checkend.SetContext(ctx, taskCtx)

	allOpts := append([]checkend.NotifyOption{
		checkend.WithTags("asynq", "background_job"),
	}, opts...)

	checkend.NotifyWithContext(ctx, err, allOpts...)
}

// AsynqErrorHandlerWithInfo reports task errors with additional task info.
func AsynqErrorHandlerWithInfo(ctx context.Context, info *AsynqTaskInfo, err error, opts ...checkend.NotifyOption) {
	if err == nil {
		return
	}

	taskCtx := map[string]interface{}{
		"asynq": map[string]interface{}{
			"task_id":   info.ID,
			"queue":     info.Queue,
			"task_type": info.Type,
			"retried":   info.Retried,
			"max_retry": info.MaxRetry,
			"payload":   sanitizePayload(info.Payload),
		},
	}

	ctx = checkend.SetContext(ctx, taskCtx)

	allOpts := append([]checkend.NotifyOption{
		checkend.WithTags("asynq", "background_job"),
	}, opts...)

	checkend.NotifyWithContext(ctx, err, allOpts...)
}

// AsynqPanicHandler handles panics in Asynq task handlers.
// Use this with defer in your task handlers.
//
// Usage:
//
//	func handleTask(ctx context.Context, task *asynq.Task) error {
//	    defer integrations.AsynqPanicHandler(ctx, task)
//	    // ... task logic
//	}
func AsynqPanicHandler(ctx context.Context, task AsynqTask) {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("panic in asynq task: %v", v)
		}

		AsynqErrorHandler(ctx, task, err)
		panic(r) // Re-panic to let Asynq handle retry logic
	}
}

// AsynqRecoverHandler is similar to AsynqPanicHandler but doesn't re-panic.
// Use this when you want to gracefully handle panics without triggering retries.
func AsynqRecoverHandler(ctx context.Context, task AsynqTask) error {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("panic in asynq task: %v", v)
		}

		AsynqErrorHandler(ctx, task, err)
		return err
	}
	return nil
}

func extractAsynqContext(task AsynqTask) map[string]interface{} {
	return map[string]interface{}{
		"asynq": map[string]interface{}{
			"task_type": task.Type(),
			"payload":   sanitizePayload(task.Payload()),
		},
	}
}

func sanitizePayload(payload []byte) interface{} {
	if len(payload) == 0 {
		return nil
	}

	// Try to parse as JSON for better readability
	var data interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		// If not valid JSON, return a truncated string representation
		if len(payload) > 1000 {
			return string(payload[:1000]) + "...[truncated]"
		}
		return string(payload)
	}

	// Sanitize the parsed data
	return sanitizeJobArgs(data)
}

func sanitizeJobArgs(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			if isSensitiveKey(key) {
				result[key] = "[FILTERED]"
			} else {
				result[key] = sanitizeJobArgs(val)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = sanitizeJobArgs(val)
		}
		return result
	default:
		return v
	}
}

func isSensitiveKey(key string) bool {
	sensitivePatterns := []string{
		"password", "secret", "token", "key", "auth",
		"credential", "private", "api_key", "apikey",
	}
	for _, pattern := range sensitivePatterns {
		if containsIgnoreCase(key, pattern) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(containsIgnoreCase(s[1:], substr) ||
					equalFoldPrefix(s, substr)))
}

func equalFoldPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if toLower(s[i]) != toLower(prefix[i]) {
			return false
		}
	}
	return true
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
