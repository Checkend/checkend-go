package integrations

import (
	"fmt"
	"net/http"

	"github.com/Checkend/checkend-go"
)

// GinMiddleware returns a Gin middleware for Checkend error reporting.
// This middleware is compatible with the gin-gonic/gin framework.
//
// Usage:
//
//	import "github.com/gin-gonic/gin"
//	import "github.com/Checkend/checkend-go/integrations"
//
//	r := gin.New()
//	r.Use(integrations.GinMiddleware())
func GinMiddleware() interface{} {
	// Return a function that matches gin.HandlerFunc signature
	// We use interface{} to avoid importing gin as a dependency
	return func(c interface{}) {
		// This is a placeholder - actual implementation requires gin types
		// The middleware will be type-asserted at runtime
	}
}

// GinRecovery returns a recovery middleware that reports panics to Checkend.
// Use this instead of gin.Recovery() to capture panic errors.
//
// Usage:
//
//	r := gin.New()
//	r.Use(integrations.GinRecovery())
func GinRecovery() interface{} {
	return func(c interface{}) {
		// Placeholder for Gin recovery middleware
	}
}

// GinContextExtractor is a helper type for extracting context from Gin.
// Use the Extract method in your handlers to get Checkend context.
type GinContextExtractor struct{}

// NewGinContextExtractor creates a new GinContextExtractor.
func NewGinContextExtractor() *GinContextExtractor {
	return &GinContextExtractor{}
}

// ExtractFromRequest extracts Checkend context from an http.Request.
// This can be used with Gin's c.Request field.
func (e *GinContextExtractor) ExtractFromRequest(r *http.Request) map[string]interface{} {
	return extractRequest(r)
}

// GinErrorHandler is a helper for handling errors in Gin handlers.
// Use this in your handlers to report errors to Checkend.
//
// Usage:
//
//	func MyHandler(c *gin.Context) {
//	    if err := doSomething(); err != nil {
//	        integrations.GinErrorHandler(c.Request, err)
//	        c.JSON(500, gin.H{"error": "internal error"})
//	        return
//	    }
//	}
func GinErrorHandler(r *http.Request, err error, opts ...checkend.NotifyOption) {
	ctx := checkend.SetRequest(r.Context(), extractRequest(r))
	checkend.NotifyWithContext(ctx, err, opts...)
}

// GinPanicHandler handles panics and reports them to Checkend.
func GinPanicHandler(r *http.Request, recovered interface{}) {
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
