package filters

import (
	"testing"
)

func TestSanitizeFilterSimple(t *testing.T) {
	filter := NewSanitizeFilter([]string{"password", "secret"})

	data := map[string]interface{}{
		"username": "john",
		"password": "secret123",
	}

	result := filter.Filter(data)

	if result["username"] != "john" {
		t.Errorf("Expected username 'john', got '%v'", result["username"])
	}

	if result["password"] != FilteredValue {
		t.Errorf("Expected password to be filtered, got '%v'", result["password"])
	}
}

func TestSanitizeFilterNested(t *testing.T) {
	filter := NewSanitizeFilter([]string{"password", "token"})

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"credentials": map[string]interface{}{
				"password":  "secret123",
				"api_token": "abc123",
			},
		},
	}

	result := filter.Filter(data)

	user := result["user"].(map[string]interface{})
	credentials := user["credentials"].(map[string]interface{})

	if user["name"] != "John" {
		t.Errorf("Expected name 'John', got '%v'", user["name"])
	}

	if credentials["password"] != FilteredValue {
		t.Errorf("Expected password to be filtered, got '%v'", credentials["password"])
	}

	if credentials["api_token"] != FilteredValue {
		t.Errorf("Expected api_token to be filtered, got '%v'", credentials["api_token"])
	}
}

func TestSanitizeFilterSlice(t *testing.T) {
	filter := NewSanitizeFilter([]string{"password"})

	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "password": "pass1"},
			map[string]interface{}{"name": "Bob", "password": "pass2"},
		},
	}

	result := filter.Filter(data)

	users := result["users"].([]interface{})
	user1 := users[0].(map[string]interface{})
	user2 := users[1].(map[string]interface{})

	if user1["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got '%v'", user1["name"])
	}

	if user1["password"] != FilteredValue {
		t.Errorf("Expected password to be filtered, got '%v'", user1["password"])
	}

	if user2["password"] != FilteredValue {
		t.Errorf("Expected password to be filtered, got '%v'", user2["password"])
	}
}

func TestSanitizeFilterCaseInsensitive(t *testing.T) {
	filter := NewSanitizeFilter([]string{"password"})

	data := map[string]interface{}{
		"PASSWORD": "value1",
		"Password": "value2",
		"password": "value3",
	}

	result := filter.Filter(data)

	for key, value := range result {
		if value != FilteredValue {
			t.Errorf("Expected %s to be filtered, got '%v'", key, value)
		}
	}
}

func TestSanitizeFilterPartialMatch(t *testing.T) {
	filter := NewSanitizeFilter([]string{"password", "secret"})

	data := map[string]interface{}{
		"user_password": "secret",
		"password_hash": "hash",
		"secret_key":    "key",
	}

	result := filter.Filter(data)

	for key, value := range result {
		if value != FilteredValue {
			t.Errorf("Expected %s to be filtered, got '%v'", key, value)
		}
	}
}

func TestSanitizeFilterPreservesNonSensitive(t *testing.T) {
	filter := NewSanitizeFilter([]string{"password"})

	data := map[string]interface{}{
		"id":     123,
		"name":   "Test",
		"active": true,
		"value":  3.14,
	}

	result := filter.Filter(data)

	if result["id"] != 123 {
		t.Errorf("Expected id 123, got %v", result["id"])
	}
	if result["name"] != "Test" {
		t.Errorf("Expected name 'Test', got %v", result["name"])
	}
	if result["active"] != true {
		t.Errorf("Expected active true, got %v", result["active"])
	}
	if result["value"] != 3.14 {
		t.Errorf("Expected value 3.14, got %v", result["value"])
	}
}

func TestSanitizeFilterHandlesNil(t *testing.T) {
	filter := NewSanitizeFilter([]string{"password"})

	result := filter.Filter(nil)

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestSanitizeFilterTruncatesLongStrings(t *testing.T) {
	filter := NewSanitizeFilter([]string{})

	longString := make([]byte, 15000)
	for i := range longString {
		longString[i] = 'x'
	}

	data := map[string]interface{}{
		"message": string(longString),
	}

	result := filter.Filter(data)

	message := result["message"].(string)
	if len(message) != maxStringLen+3 { // +3 for "..."
		t.Errorf("Expected length %d, got %d", maxStringLen+3, len(message))
	}
}

func TestSanitizeFilterHandlesDeepNesting(t *testing.T) {
	filter := NewSanitizeFilter([]string{})

	// Create deeply nested structure
	data := map[string]interface{}{"level": 0}
	current := data
	for i := 0; i < 15; i++ {
		nested := map[string]interface{}{"level": i + 1}
		current["nested"] = nested
		current = nested
	}

	result := filter.Filter(data)

	// Should not panic, should handle max depth
	if result["level"] != 0 {
		t.Errorf("Expected level 0, got %v", result["level"])
	}
}
