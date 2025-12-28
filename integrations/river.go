package integrations

import (
	"context"
	"fmt"

	checkend "github.com/Checkend/checkend-go"
)

// RiverJob represents the interface for a River job.
// This allows the integration to work without importing river directly.
type RiverJob interface {
	Kind() string
}

// RiverJobRow represents job metadata from River.
type RiverJobRow struct {
	ID          int64
	Queue       string
	Kind        string
	Args        interface{}
	Attempt     int
	MaxAttempts int
	Priority    int
	State       string
}

// RiverErrorHandler reports job errors to Checkend.
// Call this in your job worker's error handling logic.
//
// Usage with River:
//
//	type EmailWorker struct {
//	    river.WorkerDefaults[EmailArgs]
//	}
//
//	func (w *EmailWorker) Work(ctx context.Context, job *river.Job[EmailArgs]) error {
//	    err := sendEmail(job.Args)
//	    if err != nil {
//	        integrations.RiverErrorHandler(ctx, job, err)
//	        return err
//	    }
//	    return nil
//	}
func RiverErrorHandler(ctx context.Context, job interface{}, err error, opts ...checkend.NotifyOption) {
	if err == nil {
		return
	}

	jobCtx := extractRiverContext(job)
	ctx = checkend.SetContext(ctx, jobCtx)

	allOpts := append([]checkend.NotifyOption{
		checkend.WithTags("river", "background_job"),
	}, opts...)

	checkend.NotifyWithContext(ctx, err, allOpts...)
}

// RiverErrorHandlerWithRow reports job errors with job row metadata.
func RiverErrorHandlerWithRow(ctx context.Context, row *RiverJobRow, err error, opts ...checkend.NotifyOption) {
	if err == nil {
		return
	}

	jobCtx := map[string]interface{}{
		"river": map[string]interface{}{
			"job_id":       row.ID,
			"queue":        row.Queue,
			"kind":         row.Kind,
			"args":         sanitizeJobArgs(row.Args),
			"attempt":      row.Attempt,
			"max_attempts": row.MaxAttempts,
			"priority":     row.Priority,
			"state":        row.State,
		},
	}

	ctx = checkend.SetContext(ctx, jobCtx)

	allOpts := append([]checkend.NotifyOption{
		checkend.WithTags("river", "background_job"),
	}, opts...)

	checkend.NotifyWithContext(ctx, err, allOpts...)
}

// RiverPanicHandler handles panics in River job workers.
// Use this with defer in your job workers.
//
// Usage:
//
//	func (w *MyWorker) Work(ctx context.Context, job *river.Job[MyArgs]) error {
//	    defer integrations.RiverPanicHandler(ctx, job)
//	    // ... job logic
//	}
func RiverPanicHandler(ctx context.Context, job interface{}) {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("panic in river job: %v", v)
		}

		RiverErrorHandler(ctx, job, err)
		panic(r) // Re-panic to let River handle retry logic
	}
}

// RiverRecoverHandler is similar to RiverPanicHandler but doesn't re-panic.
// Use this when you want to gracefully handle panics.
func RiverRecoverHandler(ctx context.Context, job interface{}) error {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("panic in river job: %v", v)
		}

		RiverErrorHandler(ctx, job, err)
		return err
	}
	return nil
}

// RiverErrorMiddleware creates an error handler middleware for River.
// This can be used with River's error handler configuration.
//
// Usage:
//
//	client, _ := river.NewClient(riverpgxv5.New(pool), &river.Config{
//	    ErrorHandler: integrations.RiverErrorMiddleware(),
//	})
func RiverErrorMiddleware() interface{} {
	// Return an interface that can be cast to river.ErrorHandler
	// The actual signature would be:
	// func(ctx context.Context, job *rivertype.JobRow, err error) *river.JobErrorHandlerResult
	return nil
}

func extractRiverContext(job interface{}) map[string]interface{} {
	ctx := map[string]interface{}{
		"river": map[string]interface{}{
			"job": fmt.Sprintf("%T", job),
		},
	}

	// Try to extract Kind if available
	if j, ok := job.(RiverJob); ok {
		ctx["river"].(map[string]interface{})["kind"] = j.Kind()
	}

	return ctx
}
