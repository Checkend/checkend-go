package checkend

import (
	"github.com/Checkend/checkend-go/filters"
)

// SanitizeFilter wraps the filters package SanitizeFilter for internal use.
type SanitizeFilter = filters.SanitizeFilter

// NewSanitizeFilter creates a new SanitizeFilter.
func NewSanitizeFilter(filterKeys []string) *SanitizeFilter {
	return filters.NewSanitizeFilter(filterKeys)
}

// IgnoreFilter wraps the filters package IgnoreFilter for internal use.
type IgnoreFilter = filters.IgnoreFilter

// NewIgnoreFilter creates a new IgnoreFilter.
func NewIgnoreFilter(patterns []interface{}) *IgnoreFilter {
	return filters.NewIgnoreFilter(patterns)
}
