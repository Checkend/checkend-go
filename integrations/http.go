package integrations

import (
	"fmt"
	"net/http"

	"github.com/Checkend/checkend-go"
)

// HTTPMiddleware wraps an http.Handler with Checkend error reporting.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set request context
		ctx := checkend.SetRequest(r.Context(), extractRequest(r))

		// Create a response wrapper to catch panics
		defer func() {
			if err := recover(); err != nil {
				// Convert panic to error
				var e error
				switch v := err.(type) {
				case error:
					e = v
				default:
					e = fmt.Errorf("panic: %v", v)
				}

				checkend.NotifyWithContext(ctx, e)

				// Re-panic to let the default panic handler respond
				panic(err)
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HTTPMiddlewareFunc wraps an http.HandlerFunc with Checkend error reporting.
func HTTPMiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return HTTPMiddleware(next).ServeHTTP
}

func extractRequest(r *http.Request) map[string]interface{} {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	headers := make(map[string]interface{})
	for key, values := range r.Header {
		if len(values) == 1 {
			headers[key] = values[0]
		} else {
			headers[key] = values
		}
	}

	request := map[string]interface{}{
		"url":     url,
		"method":  r.Method,
		"headers": headers,
	}

	if r.URL.RawQuery != "" {
		params := make(map[string]interface{})
		for key, values := range r.URL.Query() {
			if len(values) == 1 {
				params[key] = values[0]
			} else {
				params[key] = values
			}
		}
		request["params"] = params
	}

	return request
}
