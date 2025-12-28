package filters

import (
	"reflect"
	"regexp"
	"strings"
)

// IgnoreFilter determines if an error should be ignored.
type IgnoreFilter struct {
	patterns []interface{}
}

// NewIgnoreFilter creates a new IgnoreFilter.
func NewIgnoreFilter(patterns []interface{}) *IgnoreFilter {
	return &IgnoreFilter{patterns: patterns}
}

// ShouldIgnore returns true if the error should be ignored.
func (f *IgnoreFilter) ShouldIgnore(err error) bool {
	if err == nil {
		return true
	}

	errType := reflect.TypeOf(err)
	errName := errType.String()

	// Remove pointer prefix
	errName = strings.TrimPrefix(errName, "*")

	for _, pattern := range f.patterns {
		switch p := pattern.(type) {
		case string:
			// String matching
			if f.matchesString(errName, p) {
				return true
			}
		case reflect.Type:
			// Type matching
			if errType == p || errType.AssignableTo(p) {
				return true
			}
		case error:
			// Error instance matching (compare types)
			if errType == reflect.TypeOf(p) {
				return true
			}
		}
	}

	return false
}

func (f *IgnoreFilter) matchesString(errName, pattern string) bool {
	// Exact match
	if errName == pattern {
		return true
	}

	// Suffix match (for package.Type patterns)
	if strings.HasSuffix(errName, "."+pattern) || strings.HasSuffix(errName, pattern) {
		return true
	}

	// Regex match
	if re, err := regexp.Compile(pattern); err == nil {
		if re.MatchString(errName) {
			return true
		}
	}

	return false
}
