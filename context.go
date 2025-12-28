package checkend

import (
	"context"
)

// Context keys for storing Checkend data.
type contextKey string

const (
	contextDataKey contextKey = "checkend_context_data"
)

// ContextData holds request-scoped Checkend data.
type ContextData struct {
	Context map[string]interface{}
	User    map[string]interface{}
	Request map[string]interface{}
}

// WithContextData returns a new context with Checkend data.
func WithContextData(ctx context.Context, data *ContextData) context.Context {
	return context.WithValue(ctx, contextDataKey, data)
}

// GetContextData retrieves Checkend data from the context.
func GetContextData(ctx context.Context) *ContextData {
	if data, ok := ctx.Value(contextDataKey).(*ContextData); ok {
		return data
	}
	return &ContextData{
		Context: make(map[string]interface{}),
		User:    make(map[string]interface{}),
		Request: make(map[string]interface{}),
	}
}

// SetContext adds context data to the given context.
func SetContext(ctx context.Context, data map[string]interface{}) context.Context {
	ctxData := GetContextData(ctx)
	newData := &ContextData{
		Context: make(map[string]interface{}),
		User:    ctxData.User,
		Request: ctxData.Request,
	}
	for k, v := range ctxData.Context {
		newData.Context[k] = v
	}
	for k, v := range data {
		newData.Context[k] = v
	}
	return WithContextData(ctx, newData)
}

// SetUser sets user information in the given context.
func SetUser(ctx context.Context, user map[string]interface{}) context.Context {
	ctxData := GetContextData(ctx)
	newData := &ContextData{
		Context: ctxData.Context,
		User:    user,
		Request: ctxData.Request,
	}
	return WithContextData(ctx, newData)
}

// SetRequest sets request information in the given context.
func SetRequest(ctx context.Context, request map[string]interface{}) context.Context {
	ctxData := GetContextData(ctx)
	newData := &ContextData{
		Context: ctxData.Context,
		User:    ctxData.User,
		Request: request,
	}
	return WithContextData(ctx, newData)
}
