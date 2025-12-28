package checkend

import (
	"sync"
)

var (
	testingEnabled bool
	testingNotices []*Notice
	testingMu      sync.Mutex
)

// SetupTesting enables testing mode.
func SetupTesting() {
	testingMu.Lock()
	defer testingMu.Unlock()
	testingEnabled = true
	testingNotices = nil
}

// TeardownTesting disables testing mode.
func TeardownTesting() {
	testingMu.Lock()
	defer testingMu.Unlock()
	testingEnabled = false
	testingNotices = nil
}

// ClearTesting clears testing state.
func ClearTesting() {
	testingMu.Lock()
	defer testingMu.Unlock()
	testingEnabled = false
	testingNotices = nil
}

// TestingNotices returns all captured notices.
func TestingNotices() []*Notice {
	testingMu.Lock()
	defer testingMu.Unlock()
	result := make([]*Notice, len(testingNotices))
	copy(result, testingNotices)
	return result
}

// TestingLastNotice returns the last captured notice.
func TestingLastNotice() *Notice {
	testingMu.Lock()
	defer testingMu.Unlock()
	if len(testingNotices) == 0 {
		return nil
	}
	return testingNotices[len(testingNotices)-1]
}

// TestingFirstNotice returns the first captured notice.
func TestingFirstNotice() *Notice {
	testingMu.Lock()
	defer testingMu.Unlock()
	if len(testingNotices) == 0 {
		return nil
	}
	return testingNotices[0]
}

// TestingNoticeCount returns the number of captured notices.
func TestingNoticeCount() int {
	testingMu.Lock()
	defer testingMu.Unlock()
	return len(testingNotices)
}

// TestingHasNotices returns true if any notices have been captured.
func TestingHasNotices() bool {
	testingMu.Lock()
	defer testingMu.Unlock()
	return len(testingNotices) > 0
}

// TestingClearNotices clears all captured notices.
func TestingClearNotices() {
	testingMu.Lock()
	defer testingMu.Unlock()
	testingNotices = nil
}
