package api

import (
	"sync"
	"time"
)

// RequestLogEntry represents a logged API request
type RequestLogEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Duration  string    `json:"duration"`
}

// RequestLogger stores the last N API requests for the dashboard
type RequestLogger struct {
	mu        sync.RWMutex
	entries   []RequestLogEntry
	maxSize   int
	idCounter int64
}

// NewRequestLogger creates a new RequestLogger with the specified max size
func NewRequestLogger(maxSize int) *RequestLogger {
	return &RequestLogger{
		entries: make([]RequestLogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Log adds a new request entry to the log
func (rl *RequestLogger) Log(method, path string, status int, duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.idCounter++
	entry := RequestLogEntry{
		ID:        generateRequestID(rl.idCounter),
		Timestamp: time.Now().UTC(),
		Method:    method,
		Path:      path,
		Status:    status,
		Duration:  formatDuration(duration),
	}

	// Prepend new entry (most recent first)
	rl.entries = append([]RequestLogEntry{entry}, rl.entries...)

	// Trim to max size
	if len(rl.entries) > rl.maxSize {
		rl.entries = rl.entries[:rl.maxSize]
	}
}

// GetAll returns all logged requests (most recent first)
func (rl *RequestLogger) GetAll() []RequestLogEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]RequestLogEntry, len(rl.entries))
	copy(result, rl.entries)
	return result
}

// GetRecent returns the most recent N requests
func (rl *RequestLogger) GetRecent(n int) []RequestLogEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if n > len(rl.entries) {
		n = len(rl.entries)
	}

	result := make([]RequestLogEntry, n)
	copy(result, rl.entries[:n])
	return result
}

// Clear removes all entries from the log
func (rl *RequestLogger) Clear() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.entries = make([]RequestLogEntry, 0, rl.maxSize)
}

// Count returns the number of logged requests
func (rl *RequestLogger) Count() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.entries)
}

// generateRequestID creates a simple incrementing ID
func generateRequestID(counter int64) string {
	return time.Now().Format("20060102150405") + "-" + formatCounter(counter)
}

// formatCounter formats the counter with leading zeros
func formatCounter(n int64) string {
	return padLeft(itoa(n), 4, '0')
}

// itoa converts int64 to string without importing strconv
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	return string(buf[i+1:])
}

// padLeft pads a string with a character on the left
func padLeft(s string, n int, c byte) string {
	if len(s) >= n {
		return s
	}
	pad := make([]byte, n-len(s))
	for i := range pad {
		pad[i] = c
	}
	return string(pad) + s
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.String()
	}
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(time.Millisecond).String()
}
