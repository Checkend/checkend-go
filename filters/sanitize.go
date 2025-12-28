package filters

import (
	"strings"
)

const (
	// FilteredValue is the replacement for sensitive values.
	FilteredValue = "[FILTERED]"
	maxDepth      = 10
	maxStringLen  = 10000
)

// SanitizeFilter removes sensitive data from payloads.
type SanitizeFilter struct {
	filterKeys []string
	seen       map[uintptr]bool
}

// NewSanitizeFilter creates a new SanitizeFilter.
func NewSanitizeFilter(filterKeys []string) *SanitizeFilter {
	lowerKeys := make([]string, len(filterKeys))
	for i, k := range filterKeys {
		lowerKeys[i] = strings.ToLower(k)
	}
	return &SanitizeFilter{
		filterKeys: lowerKeys,
	}
}

// Filter recursively filters sensitive data from an object.
func (f *SanitizeFilter) Filter(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}
	f.seen = make(map[uintptr]bool)
	return f.filterMap(data, 0)
}

func (f *SanitizeFilter) filterMap(data map[string]interface{}, depth int) map[string]interface{} {
	if depth > maxDepth {
		return map[string]interface{}{"_truncated": "[MAX DEPTH EXCEEDED]"}
	}

	result := make(map[string]interface{})
	for key, value := range data {
		if f.shouldFilter(key) {
			result[key] = FilteredValue
		} else {
			result[key] = f.filterValue(value, depth+1)
		}
	}
	return result
}

func (f *SanitizeFilter) filterValue(value interface{}, depth int) interface{} {
	if depth > maxDepth {
		return "[MAX DEPTH EXCEEDED]"
	}

	switch v := value.(type) {
	case map[string]interface{}:
		return f.filterMap(v, depth)
	case []interface{}:
		return f.filterSlice(v, depth)
	case string:
		return f.truncateString(v)
	case nil, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return v
	default:
		// Convert to string for unknown types
		return f.truncateString(valueToString(v))
	}
}

func (f *SanitizeFilter) filterSlice(data []interface{}, depth int) []interface{} {
	result := make([]interface{}, len(data))
	for i, item := range data {
		result[i] = f.filterValue(item, depth+1)
	}
	return result
}

func (f *SanitizeFilter) shouldFilter(key string) bool {
	keyLower := strings.ToLower(key)
	for _, filterKey := range f.filterKeys {
		if strings.Contains(keyLower, filterKey) {
			return true
		}
	}
	return false
}

func (f *SanitizeFilter) truncateString(s string) string {
	if len(s) > maxStringLen {
		return s[:maxStringLen] + "..."
	}
	return s
}

func valueToString(v interface{}) string {
	if stringer, ok := v.(interface{ String() string }); ok {
		return stringer.String()
	}
	return "[UNKNOWN TYPE]"
}
