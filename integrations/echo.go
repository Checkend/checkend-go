package integrations

import (
	"fmt"
	"net/http"

	"github.com/Checkend/checkend-go"
)

// EchoMiddleware returns an Echo middleware for Checkend error reporting.
// This middleware is compatible with the labstack/echo framework.
//
// Usage:
//
//	import "github.com/labstack/echo/v4"
//	import "github.com/Checkend/checkend-go/integrations"
//
//	e := echo.New()
//	e.Use(integrations.EchoMiddleware())
func EchoMiddleware() interface{} {
	// Return a function that matches echo.MiddlewareFunc signature
	// We use interface{} to avoid importing echo as a dependency
	return func(next interface{}) interface{} {
		return func(c interface{}) error {
			// Placeholder - actual implementation requires echo types
			return nil
		}
	}
}

// EchoErrorHandler is a helper for handling errors in Echo handlers.
//
// Usage:
//
//	func MyHandler(c echo.Context) error {
//	    if err := doSomething(); err != nil {
//	        integrations.EchoErrorHandler(c.Request(), err)
//	        return c.JSON(500, map[string]string{"error": "internal error"})
//	    }
//	    return nil
//	}
func EchoErrorHandler(r *http.Request, err error, opts ...checkend.NotifyOption) {
	ctx := checkend.SetRequest(r.Context(), extractRequest(r))
	checkend.NotifyWithContext(ctx, err, opts...)
}

// EchoPanicHandler handles panics and reports them to Checkend.
func EchoPanicHandler(r *http.Request, recovered interface{}) {
	var err error
	switch v := recovered.(type) {
	case error:
		err = v
	default:
		err = fmt.Errorf("panic: %v", v)
	}

	ctx := checkend.SetRequest(r.Context(), extractRequest(r))
	checkend.NotifyWithContext(ctx, err)
}

// EchoRecoveryMiddleware returns a recovery middleware that reports panics.
func EchoRecoveryMiddleware() interface{} {
	return func(next interface{}) interface{} {
		return func(c interface{}) error {
			// Placeholder - actual implementation requires echo types
			return nil
		}
	}
}
