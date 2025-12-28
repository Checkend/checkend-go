package checkend

import (
	"context"
	"errors"
	"testing"
)

func TestConfigure(t *testing.T) {
	defer Reset()

	cfg := Configure(Config{
		APIKey:  "test-key",
		Enabled: boolPtr(true),
	})

	if cfg.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got '%s'", cfg.APIKey)
	}

	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}
}

func TestNotify(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
	})

	err := errors.New("test error")
	Notify(err)

	if !TestingHasNotices() {
		t.Error("Expected notices to be captured")
	}

	if TestingNoticeCount() != 1 {
		t.Errorf("Expected 1 notice, got %d", TestingNoticeCount())
	}

	notice := TestingLastNotice()
	if notice.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", notice.Message)
	}
}

func TestNotifyWithContext(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
	})

	err := errors.New("test error")
	Notify(err, WithContext(map[string]interface{}{
		"order_id": 123,
	}))

	notice := TestingLastNotice()
	if notice.Context["order_id"] != 123 {
		t.Errorf("Expected order_id 123, got %v", notice.Context["order_id"])
	}
}

func TestNotifyWithUser(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
	})

	err := errors.New("test error")
	Notify(err, WithUser(map[string]interface{}{
		"id":    "user-1",
		"email": "test@example.com",
	}))

	notice := TestingLastNotice()
	if notice.User["id"] != "user-1" {
		t.Errorf("Expected user id 'user-1', got %v", notice.User["id"])
	}
}

func TestNotifyWithTags(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
	})

	err := errors.New("test error")
	Notify(err, WithTags("critical", "backend"))

	notice := TestingLastNotice()
	if len(notice.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(notice.Tags))
	}
}

func TestNotifyWithFingerprint(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
	})

	err := errors.New("test error")
	Notify(err, WithFingerprint("custom-fingerprint"))

	notice := TestingLastNotice()
	if notice.Fingerprint != "custom-fingerprint" {
		t.Errorf("Expected fingerprint 'custom-fingerprint', got '%s'", notice.Fingerprint)
	}
}

func TestNotifyDisabled(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:  "test-key",
		Enabled: boolPtr(false),
	})

	err := errors.New("test error")
	Notify(err)

	if TestingHasNotices() {
		t.Error("Expected no notices when disabled")
	}
}

func TestNotifySync(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
	})

	err := errors.New("test error")
	response := NotifySync(err)

	if response == nil {
		t.Error("Expected response, got nil")
	}

	if !TestingHasNotices() {
		t.Error("Expected notices to be captured")
	}
}

func TestSetContext(t *testing.T) {
	ctx := context.Background()
	ctx = SetContext(ctx, map[string]interface{}{
		"key1": "value1",
	})
	ctx = SetContext(ctx, map[string]interface{}{
		"key2": "value2",
	})

	data := GetContextData(ctx)
	if data.Context["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got %v", data.Context["key1"])
	}
	if data.Context["key2"] != "value2" {
		t.Errorf("Expected key2 'value2', got %v", data.Context["key2"])
	}
}

func TestSetUser(t *testing.T) {
	ctx := context.Background()
	ctx = SetUser(ctx, map[string]interface{}{
		"id":    "user-1",
		"email": "test@example.com",
	})

	data := GetContextData(ctx)
	if data.User["id"] != "user-1" {
		t.Errorf("Expected user id 'user-1', got %v", data.User["id"])
	}
}

func TestSetRequest(t *testing.T) {
	ctx := context.Background()
	ctx = SetRequest(ctx, map[string]interface{}{
		"url":    "https://example.com",
		"method": "POST",
	})

	data := GetContextData(ctx)
	if data.Request["url"] != "https://example.com" {
		t.Errorf("Expected url 'https://example.com', got %v", data.Request["url"])
	}
}

func TestBeforeNotifyCallback(t *testing.T) {
	defer Reset()

	SetupTesting()

	called := false
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
		BeforeNotify: []func(*Notice) bool{
			func(n *Notice) bool {
				called = true
				return true
			},
		},
	})

	err := errors.New("test error")
	Notify(err)

	if !called {
		t.Error("Expected before_notify callback to be called")
	}
}

func TestBeforeNotifyCanSkip(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
		BeforeNotify: []func(*Notice) bool{
			func(n *Notice) bool {
				return false // Skip sending
			},
		},
	})

	err := errors.New("test error")
	Notify(err)

	if TestingHasNotices() {
		t.Error("Expected no notices when before_notify returns false")
	}
}

func TestContextMergedIntoNotice(t *testing.T) {
	defer Reset()

	SetupTesting()
	Configure(Config{
		APIKey:    "test-key",
		Enabled:   boolPtr(true),
		AsyncSend: false,
	})

	ctx := context.Background()
	ctx = SetContext(ctx, map[string]interface{}{
		"global_key": "global_value",
	})

	err := errors.New("test error")
	NotifyWithContext(ctx, err, WithContext(map[string]interface{}{
		"local_key": "local_value",
	}))

	notice := TestingLastNotice()
	if notice.Context["global_key"] != "global_value" {
		t.Errorf("Expected global_key 'global_value', got %v", notice.Context["global_key"])
	}
	if notice.Context["local_key"] != "local_value" {
		t.Errorf("Expected local_key 'local_value', got %v", notice.Context["local_key"])
	}
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
