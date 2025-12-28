package integrations

import (
	"context"
	"fmt"

	checkend "github.com/Checkend/checkend-go"
)

// MachineryTask represents the interface for a Machinery task.
// This allows the integration to work without importing machinery directly.
type MachineryTask interface {
	GetName() string
	GetUUID() string
}

// MachinerySignature represents task metadata from Machinery.
type MachinerySignature struct {
	UUID         string
	Name         string
	RoutingKey   string
	Args         []interface{}
	RetryCount   int
	RetryTimeout int
}

// MachineryErrorHandler reports task errors to Checkend.
// Call this in your task's error handling logic.
//
// Usage with Machinery:
//
//	func sendEmailTask(email string) error {
//	    err := sendEmail(email)
//	    if err != nil {
//	        integrations.MachineryErrorHandler(context.Background(), "send_email", err)
//	        return err
//	    }
//	    return nil
//	}
func MachineryErrorHandler(ctx context.Context, taskName string, err error, opts ...checkend.NotifyOption) {
	if err == nil {
		return
	}

	taskCtx := map[string]interface{}{
		"machinery": map[string]interface{}{
			"task_name": taskName,
		},
	}

	ctx = checkend.SetContext(ctx, taskCtx)

	allOpts := append([]checkend.NotifyOption{
		checkend.WithTags("machinery", "background_job"),
	}, opts...)

	checkend.NotifyWithContext(ctx, err, allOpts...)
}

// MachineryErrorHandlerWithSignature reports task errors with full signature metadata.
func MachineryErrorHandlerWithSignature(ctx context.Context, sig *MachinerySignature, err error, opts ...checkend.NotifyOption) {
	if err == nil {
		return
	}

	taskCtx := map[string]interface{}{
		"machinery": map[string]interface{}{
			"uuid":          sig.UUID,
			"task_name":     sig.Name,
			"routing_key":   sig.RoutingKey,
			"args":          sanitizeJobArgs(sig.Args),
			"retry_count":   sig.RetryCount,
			"retry_timeout": sig.RetryTimeout,
		},
	}

	ctx = checkend.SetContext(ctx, taskCtx)

	allOpts := append([]checkend.NotifyOption{
		checkend.WithTags("machinery", "background_job"),
	}, opts...)

	checkend.NotifyWithContext(ctx, err, allOpts...)
}

// MachineryOnTaskFailure creates a callback for Machinery's OnTaskFailure hook.
// Use this when configuring Machinery server callbacks.
//
// Usage:
//
//	server.SetErrorHandler(integrations.MachineryOnTaskFailure())
func MachineryOnTaskFailure() func(signature interface{}, err error) {
	return func(signature interface{}, err error) {
		if err == nil {
			return
		}

		ctx := context.Background()
		taskCtx := extractMachineryContext(signature)
		ctx = checkend.SetContext(ctx, taskCtx)

		checkend.NotifyWithContext(ctx, err,
			checkend.WithTags("machinery", "background_job", "task_failure"),
		)
	}
}

// MachineryOnTaskSuccessWithError creates a callback that reports errors
// even when the task "succeeds" but returns an error value.
func MachineryOnTaskSuccessWithError() func(signature interface{}, results []interface{}) {
	return func(signature interface{}, results []interface{}) {
		// Check if any result is an error
		for _, result := range results {
			if err, ok := result.(error); ok && err != nil {
				ctx := context.Background()
				taskCtx := extractMachineryContext(signature)
				ctx = checkend.SetContext(ctx, taskCtx)

				checkend.NotifyWithContext(ctx, err,
					checkend.WithTags("machinery", "background_job", "task_error_result"),
				)
			}
		}
	}
}

// MachineryPanicHandler handles panics in Machinery tasks.
// Use this with defer in your task functions.
//
// Usage:
//
//	func myTask(args ...interface{}) error {
//	    defer integrations.MachineryPanicHandler("my_task")
//	    // ... task logic
//	}
func MachineryPanicHandler(taskName string) {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("panic in machinery task: %v", v)
		}

		MachineryErrorHandler(context.Background(), taskName, err)
		panic(r) // Re-panic to let Machinery handle retry logic
	}
}

// MachineryRecoverHandler is similar to MachineryPanicHandler but doesn't re-panic.
func MachineryRecoverHandler(taskName string) error {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("panic in machinery task: %v", v)
		}

		MachineryErrorHandler(context.Background(), taskName, err)
		return err
	}
	return nil
}

func extractMachineryContext(signature interface{}) map[string]interface{} {
	ctx := map[string]interface{}{
		"machinery": map[string]interface{}{
			"signature": fmt.Sprintf("%T", signature),
		},
	}

	// Try to extract task info if available
	if task, ok := signature.(MachineryTask); ok {
		ctx["machinery"].(map[string]interface{})["task_name"] = task.GetName()
		ctx["machinery"].(map[string]interface{})["uuid"] = task.GetUUID()
	}

	return ctx
}
