# Checkend Go SDK

[![CI](https://github.com/Checkend/checkend-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Checkend/checkend-go/actions/workflows/ci.yml)

Go SDK for [Checkend](https://checkend.io) error monitoring. Zero dependencies, async by default.

## Features

- **Zero dependencies** - Uses only Go standard library
- **Async by default** - Non-blocking error sending via goroutine worker
- **Framework integrations** - net/http, Gin, Echo middleware helpers
- **Job queue integrations** - Asynq, River, Machinery support
- **Context-based** - Request-scoped context using `context.Context`
- **Sensitive data filtering** - Automatic scrubbing of passwords, tokens, etc.
- **Testing utilities** - Capture errors in tests without sending

## Installation

```bash
go get github.com/Checkend/checkend-go
```

## Quick Start

```go
package main

import (
    "errors"
    "github.com/Checkend/checkend-go"
)

func main() {
    // Configure the SDK
    checkend.Configure(checkend.Config{
        APIKey: "your-api-key",
    })
    defer checkend.Stop()

    // Report an error
    if err := doSomething(); err != nil {
        checkend.Notify(err)
    }
}
```

## Configuration

```go
import "github.com/Checkend/checkend-go"

enabled := true
sendEnv := false
checkend.Configure(checkend.Config{
    // Required
    APIKey: "your-api-key",

    // API Settings
    Endpoint:    "https://app.checkend.io",   // Custom endpoint
    Environment: "production",                 // Auto-detected from GO_ENV, etc.
    Enabled:     &enabled,                     // Enable/disable reporting

    // Application Metadata
    AppName:  "my-app",                        // Application identifier
    Revision: "abc123",                        // Git commit/revision
    RootPath: "/app",                          // Root path for backtrace cleaning

    // HTTP Settings
    Timeout:        15 * time.Second,          // Request timeout (default: 15s)
    ConnectTimeout: 5 * time.Second,           // Connection timeout (default: 5s)
    Proxy:          "http://proxy:8080",       // HTTP proxy URL
    SSLVerify:      &enabled,                  // TLS verification (default: true)

    // Async Settings
    AsyncSend:       true,                     // Async sending (default: true)
    MaxQueueSize:    1000,                     // Max queue size (default: 1000)
    ShutdownTimeout: 5 * time.Second,          // Graceful shutdown timeout (default: 5s)

    // Data Control
    SendRequestData: &enabled,                 // Include request data (default: true)
    SendUserData:    &enabled,                 // Include user data (default: true)
    SendEnvironment: &sendEnv,                 // Include env vars (default: false)
    SendSessionData: &enabled,                 // Include session data (default: true)

    // Filtering
    FilterKeys:    []string{"custom_secret"},  // Additional keys to filter
    IgnoredErrors: []interface{}{MyError{}},   // Errors to ignore

    // Callbacks
    BeforeNotify: []func(*checkend.Notice) bool{...},

    // Debug
    Debug: false,                              // Enable debug logging
})
```

### Environment Variables

```bash
# Core Settings
CHECKEND_API_KEY=your-api-key
CHECKEND_ENDPOINT=https://your-server.com
CHECKEND_ENVIRONMENT=production
CHECKEND_DEBUG=true

# Application Metadata
CHECKEND_APP_NAME=my-app
CHECKEND_REVISION=abc123      # Also reads GIT_COMMIT
CHECKEND_ROOT_PATH=/app

# HTTP Settings
HTTPS_PROXY=http://proxy:8080
HTTP_PROXY=http://proxy:8080
CHECKEND_SSL_VERIFY=false
```

## Manual Error Reporting

```go
import "github.com/Checkend/checkend-go"

// Basic error reporting
if err := riskyOperation(); err != nil {
    checkend.Notify(err)
}

// With additional context
checkend.Notify(err,
    checkend.WithContext(map[string]interface{}{
        "order_id": orderID,
    }),
    checkend.WithUser(map[string]interface{}{
        "id":    user.ID,
        "email": user.Email,
    }),
    checkend.WithTags("orders", "critical"),
    checkend.WithFingerprint("order-processing-error"),
)

// Synchronous sending (blocks until sent)
response := checkend.NotifySync(err)
fmt.Printf("Notice ID: %d\n", response.ID)
```

## Context & User Tracking

```go
import (
    "context"
    "github.com/Checkend/checkend-go"
)

// Create context with Checkend data
ctx := context.Background()
ctx = checkend.SetContext(ctx, map[string]interface{}{
    "order_id":     12345,
    "feature_flag": "new-checkout",
})

ctx = checkend.SetUser(ctx, map[string]interface{}{
    "id":    user.ID,
    "email": user.Email,
    "name":  user.Name,
})

ctx = checkend.SetRequest(ctx, map[string]interface{}{
    "url":    request.URL.String(),
    "method": request.Method,
})

// Report error with context
checkend.NotifyWithContext(ctx, err)
```

## Framework Integrations

### net/http

```go
import (
    "net/http"
    "github.com/Checkend/checkend-go"
    "github.com/Checkend/checkend-go/integrations"
)

func main() {
    checkend.Configure(checkend.Config{APIKey: "your-api-key"})
    defer checkend.Stop()

    handler := http.HandlerFunc(myHandler)
    http.Handle("/", integrations.HTTPMiddleware(handler))
    http.ListenAndServe(":8080", nil)
}
```

### Gin

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/Checkend/checkend-go"
    "github.com/Checkend/checkend-go/integrations"
)

func main() {
    checkend.Configure(checkend.Config{APIKey: "your-api-key"})
    defer checkend.Stop()

    r := gin.New()

    // Use recovery middleware that reports to Checkend
    r.Use(func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                integrations.GinPanicHandler(c.Request, err)
                panic(err) // Re-panic for Gin's default handler
            }
        }()
        c.Next()
    })

    // Or manually report errors in handlers
    r.GET("/api/users", func(c *gin.Context) {
        if err := doSomething(); err != nil {
            integrations.GinErrorHandler(c.Request, err)
            c.JSON(500, gin.H{"error": "internal error"})
            return
        }
    })
}
```

### Echo

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/Checkend/checkend-go"
    "github.com/Checkend/checkend-go/integrations"
)

func main() {
    checkend.Configure(checkend.Config{APIKey: "your-api-key"})
    defer checkend.Stop()

    e := echo.New()

    // Report errors in handlers
    e.GET("/api/users", func(c echo.Context) error {
        if err := doSomething(); err != nil {
            integrations.EchoErrorHandler(c.Request(), err)
            return c.JSON(500, map[string]string{"error": "internal error"})
        }
        return nil
    })
}
```

## Job Queue Integrations

### Asynq (Redis-based)

```go
import (
    "context"
    "github.com/hibiken/asynq"
    "github.com/Checkend/checkend-go"
    "github.com/Checkend/checkend-go/integrations"
)

func handleEmailTask(ctx context.Context, task *asynq.Task) error {
    // Use panic handler for unexpected errors
    defer integrations.AsynqPanicHandler(ctx, task)

    err := sendEmail(task.Payload())
    if err != nil {
        // Report the error to Checkend
        integrations.AsynqErrorHandler(ctx, task, err)
        return err
    }
    return nil
}

// With full task info
func handleWithInfo(ctx context.Context, task *asynq.Task, info *asynq.TaskInfo) error {
    err := doWork()
    if err != nil {
        integrations.AsynqErrorHandlerWithInfo(ctx, &integrations.AsynqTaskInfo{
            ID:       info.ID,
            Queue:    info.Queue,
            Type:     info.Type,
            Retried:  info.Retried,
            MaxRetry: info.MaxRetry,
        }, err)
        return err
    }
    return nil
}
```

### River (Postgres-based)

```go
import (
    "context"
    "github.com/riverqueue/river"
    "github.com/Checkend/checkend-go"
    "github.com/Checkend/checkend-go/integrations"
)

type EmailWorker struct {
    river.WorkerDefaults[EmailArgs]
}

func (w *EmailWorker) Work(ctx context.Context, job *river.Job[EmailArgs]) error {
    // Use panic handler for unexpected errors
    defer integrations.RiverPanicHandler(ctx, job)

    err := sendEmail(job.Args.To, job.Args.Subject)
    if err != nil {
        // Report the error to Checkend
        integrations.RiverErrorHandler(ctx, job, err)
        return err
    }
    return nil
}
```

### Machinery (Distributed)

```go
import (
    "github.com/RichardKnop/machinery/v2"
    "github.com/Checkend/checkend-go"
    "github.com/Checkend/checkend-go/integrations"
)

func sendEmailTask(to, subject string) error {
    // Use panic handler for unexpected errors
    defer integrations.MachineryPanicHandler("send_email")

    err := sendEmail(to, subject)
    if err != nil {
        // Report the error to Checkend
        integrations.MachineryErrorHandler(context.Background(), "send_email", err)
        return err
    }
    return nil
}

// Configure error handler on server
server.SetErrorHandler(integrations.MachineryOnTaskFailure())
```

## Testing

Use the testing functions to capture errors without sending them:

```go
import (
    "testing"
    "errors"
    "github.com/Checkend/checkend-go"
)

func TestErrorReporting(t *testing.T) {
    // Enable testing mode
    checkend.SetupTesting()
    defer checkend.Reset()

    enabled := true
    checkend.Configure(checkend.Config{
        APIKey:  "test-key",
        Enabled: &enabled,
    })

    // Trigger an error
    checkend.Notify(errors.New("test error"))

    // Assert on captured notices
    if !checkend.TestingHasNotices() {
        t.Error("Expected notices to be captured")
    }

    if checkend.TestingNoticeCount() != 1 {
        t.Errorf("Expected 1 notice, got %d", checkend.TestingNoticeCount())
    }

    notice := checkend.TestingLastNotice()
    if notice.Message != "test error" {
        t.Errorf("Expected message 'test error', got '%s'", notice.Message)
    }
}
```

## Filtering Sensitive Data

By default, these keys are filtered: `password`, `secret`, `token`, `api_key`, `authorization`, `credit_card`, `cvv`, `ssn`, etc.

Add custom keys:

```go
checkend.Configure(checkend.Config{
    APIKey:     "your-api-key",
    FilterKeys: []string{"custom_secret", "internal_token"},
})
```

Filtered values appear as `[FILTERED]` in the dashboard.

## Ignoring Errors

```go
checkend.Configure(checkend.Config{
    APIKey: "your-api-key",
    IgnoredErrors: []interface{}{
        &MyCustomError{},     // By error instance type
        "context.Canceled",   // By string pattern
        ".*timeout.*",        // By regex
    },
})
```

## Before Notify Callbacks

```go
checkend.Configure(checkend.Config{
    APIKey: "your-api-key",
    BeforeNotify: []func(*checkend.Notice) bool{
        func(notice *checkend.Notice) bool {
            notice.Context["server"] = "web-1"
            return true // Continue sending
        },
        func(notice *checkend.Notice) bool {
            if strings.Contains(notice.Message, "ignore-me") {
                return false // Skip sending
            }
            return true
        },
    },
})
```

## Graceful Shutdown

The SDK automatically flushes pending notices when `Stop()` is called with a configurable timeout. Always defer `Stop()`:

```go
func main() {
    checkend.Configure(checkend.Config{
        APIKey:          "your-api-key",
        ShutdownTimeout: 10 * time.Second, // Wait up to 10s for pending notices
    })
    defer checkend.Stop()

    // Your application code...
}
```

For manual control:

```go
// Wait for pending notices to send
checkend.Flush()

// Stop the worker
checkend.Stop()
```

## Requirements

- Go 1.21+
- No external dependencies

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Build
make build

# Install git hooks
make install-hooks
```

Or using Go directly:

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestNotify
```

## License

MIT License - see [LICENSE](LICENSE) for details.
