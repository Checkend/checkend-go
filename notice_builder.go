package checkend

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

const (
	maxBacktraceLines = 100
	maxMessageLength  = 10000
)

// NoticeBuilder builds Notice objects from errors.
type NoticeBuilder struct {
	config         *Configuration
	sanitizeFilter *SanitizeFilter
}

// NewNoticeBuilder creates a new NoticeBuilder.
func NewNoticeBuilder(config *Configuration) *NoticeBuilder {
	return &NoticeBuilder{
		config:         config,
		sanitizeFilter: NewSanitizeFilter(config.FilterKeys),
	}
}

// Build creates a Notice from an error.
func (b *NoticeBuilder) Build(
	err error,
	context map[string]interface{},
	user map[string]interface{},
	request map[string]interface{},
	fingerprint string,
	tags []string,
) *Notice {
	errorClass := b.extractClassName(err)
	message := b.extractMessage(err)
	backtrace := b.extractBacktrace()

	// Sanitize context (always included)
	sanitizedContext := b.sanitizeFilter.Filter(context)

	// Add environment variables if enabled
	if b.config.SendEnvironment {
		sanitizedContext["env"] = b.getEnvironmentVars()
	}

	// Conditionally include user data
	var sanitizedUser map[string]interface{}
	if b.config.SendUserData && len(user) > 0 {
		sanitizedUser = b.sanitizeFilter.Filter(user)
	}

	// Conditionally include request data
	var sanitizedRequest map[string]interface{}
	if b.config.SendRequestData && len(request) > 0 {
		sanitizedRequest = b.sanitizeFilter.Filter(request)
	}

	return &Notice{
		ErrorClass:  errorClass,
		Message:     message,
		Backtrace:   backtrace,
		Fingerprint: fingerprint,
		Tags:        tags,
		Context:     sanitizedContext,
		User:        sanitizedUser,
		Request:     sanitizedRequest,
		Environment: b.config.Environment,
		OccurredAt:  time.Now().UTC(),
		Notifier:    b.buildNotifier(),
		AppName:     b.config.AppName,
		Revision:    b.config.Revision,
		Hostname:    b.getHostname(),
	}
}

func (b *NoticeBuilder) extractClassName(err error) string {
	t := reflect.TypeOf(err)
	if t == nil {
		return "error"
	}

	// Get the type name
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	if name == "" {
		name = t.String()
	}

	// Include package path for custom types
	if pkg := t.PkgPath(); pkg != "" && !strings.HasPrefix(pkg, "errors") {
		return fmt.Sprintf("%s.%s", pkg, name)
	}

	return name
}

func (b *NoticeBuilder) extractMessage(err error) string {
	message := err.Error()
	if len(message) > maxMessageLength {
		message = message[:maxMessageLength] + "..."
	}
	return message
}

func (b *NoticeBuilder) extractBacktrace() []string {
	var backtrace []string

	// Skip frames from checkend package
	skip := 4 // Adjust based on call depth

	pcs := make([]uintptr, maxBacktraceLines)
	n := runtime.Callers(skip, pcs)
	pcs = pcs[:n]

	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()

		// Skip internal checkend frames
		if strings.Contains(frame.File, "checkend-go") {
			if !more {
				break
			}
			continue
		}

		// Clean file path using RootPath
		filePath := b.cleanFilePath(frame.File)
		line := fmt.Sprintf("%s:%d in %s", filePath, frame.Line, frame.Function)
		backtrace = append(backtrace, line)

		if !more {
			break
		}
	}

	return backtrace
}

// cleanFilePath removes RootPath prefix from file paths for cleaner backtraces.
func (b *NoticeBuilder) cleanFilePath(path string) string {
	if b.config.RootPath != "" && strings.HasPrefix(path, b.config.RootPath) {
		cleaned := strings.TrimPrefix(path, b.config.RootPath)
		// Remove leading slash if present
		cleaned = strings.TrimPrefix(cleaned, "/")
		return cleaned
	}
	return path
}

func (b *NoticeBuilder) buildNotifier() NotifierInfo {
	return NotifierInfo{
		Name:            "checkend-go",
		Version:         Version,
		Language:        "go",
		LanguageVersion: runtime.Version(),
	}
}

// getEnvironmentVars returns filtered environment variables.
func (b *NoticeBuilder) getEnvironmentVars() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			key := pair[0]
			if !b.isSensitiveEnvVar(key) {
				env[key] = pair[1]
			}
		}
	}
	return env
}

// isSensitiveEnvVar checks if an environment variable name contains sensitive patterns.
func (b *NoticeBuilder) isSensitiveEnvVar(key string) bool {
	sensitivePatterns := []string{
		"SECRET",
		"PASSWORD",
		"KEY",
		"TOKEN",
		"CREDENTIAL",
		"AUTH",
		"PRIVATE",
	}
	keyUpper := strings.ToUpper(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(keyUpper, pattern) {
			return true
		}
	}
	return false
}

// getHostname returns the current hostname.
func (b *NoticeBuilder) getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}
