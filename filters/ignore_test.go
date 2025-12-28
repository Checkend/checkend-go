package filters

import (
	"errors"
	"testing"
)

type customError struct {
	message string
}

func (e *customError) Error() string {
	return e.message
}

type anotherError struct {
	message string
}

func (e *anotherError) Error() string {
	return e.message
}

func TestIgnoreFilterByString(t *testing.T) {
	filter := NewIgnoreFilter([]interface{}{"customError"})

	err := &customError{message: "test"}
	if !filter.ShouldIgnore(err) {
		t.Error("Expected error to be ignored")
	}

	err2 := &anotherError{message: "test"}
	if filter.ShouldIgnore(err2) {
		t.Error("Expected error not to be ignored")
	}
}

func TestIgnoreFilterByErrorInstance(t *testing.T) {
	customErr := &customError{message: "test"}
	filter := NewIgnoreFilter([]interface{}{customErr})

	err := &customError{message: "another"}
	if !filter.ShouldIgnore(err) {
		t.Error("Expected error to be ignored")
	}
}

func TestIgnoreFilterByRegex(t *testing.T) {
	filter := NewIgnoreFilter([]interface{}{".*Error"})

	err := errors.New("test")
	// errors.errorString doesn't match .*Error
	if filter.ShouldIgnore(err) {
		t.Error("Expected error not to be ignored")
	}

	customErr := &customError{message: "test"}
	if !filter.ShouldIgnore(customErr) {
		t.Error("Expected customError to be ignored")
	}
}

func TestIgnoreFilterMultiplePatterns(t *testing.T) {
	filter := NewIgnoreFilter([]interface{}{"customError", "anotherError"})

	err1 := &customError{message: "test"}
	err2 := &anotherError{message: "test"}
	err3 := errors.New("test")

	if !filter.ShouldIgnore(err1) {
		t.Error("Expected customError to be ignored")
	}

	if !filter.ShouldIgnore(err2) {
		t.Error("Expected anotherError to be ignored")
	}

	if filter.ShouldIgnore(err3) {
		t.Error("Expected generic error not to be ignored")
	}
}

func TestIgnoreFilterEmptyList(t *testing.T) {
	filter := NewIgnoreFilter([]interface{}{})

	err := &customError{message: "test"}
	if filter.ShouldIgnore(err) {
		t.Error("Expected no errors to be ignored with empty list")
	}
}

func TestIgnoreFilterNilError(t *testing.T) {
	filter := NewIgnoreFilter([]interface{}{"customError"})

	if !filter.ShouldIgnore(nil) {
		t.Error("Expected nil error to be ignored")
	}
}
