package api

import (
	"testing"
	"time"
)

func TestNewRequestLogger(t *testing.T) {
	logger := NewRequestLogger(50)

	if logger == nil {
		t.Fatal("Expected logger to be created, got nil")
	}

	if logger.maxSize != 50 {
		t.Errorf("Expected maxSize 50, got %d", logger.maxSize)
	}

	if len(logger.entries) != 0 {
		t.Errorf("Expected empty entries, got %d", len(logger.entries))
	}
}

func TestRequestLogger_Log(t *testing.T) {
	logger := NewRequestLogger(50)

	// Log a request
	logger.Log("GET", "/storage/v1/b", 200, 15*time.Millisecond)

	entries := logger.GetAll()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Method != "GET" {
		t.Errorf("Expected method GET, got %s", entry.Method)
	}
	if entry.Path != "/storage/v1/b" {
		t.Errorf("Expected path /storage/v1/b, got %s", entry.Path)
	}
	if entry.Status != 200 {
		t.Errorf("Expected status 200, got %d", entry.Status)
	}
	if entry.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestRequestLogger_MaxSize(t *testing.T) {
	logger := NewRequestLogger(5)

	// Log more requests than max size
	for i := 0; i < 10; i++ {
		logger.Log("GET", "/test", 200, time.Millisecond)
	}

	entries := logger.GetAll()
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries (max size), got %d", len(entries))
	}
}

func TestRequestLogger_OrderMostRecentFirst(t *testing.T) {
	logger := NewRequestLogger(50)

	logger.Log("GET", "/first", 200, time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	logger.Log("POST", "/second", 201, time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	logger.Log("DELETE", "/third", 204, time.Millisecond)

	entries := logger.GetAll()
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Most recent should be first
	if entries[0].Path != "/third" {
		t.Errorf("Expected /third first, got %s", entries[0].Path)
	}
	if entries[1].Path != "/second" {
		t.Errorf("Expected /second second, got %s", entries[1].Path)
	}
	if entries[2].Path != "/first" {
		t.Errorf("Expected /first third, got %s", entries[2].Path)
	}
}

func TestRequestLogger_GetRecent(t *testing.T) {
	logger := NewRequestLogger(50)

	for i := 0; i < 10; i++ {
		logger.Log("GET", "/test", 200, time.Millisecond)
	}

	recent := logger.GetRecent(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(recent))
	}

	// Request more than available
	all := logger.GetRecent(100)
	if len(all) != 10 {
		t.Errorf("Expected 10 entries, got %d", len(all))
	}
}

func TestRequestLogger_Clear(t *testing.T) {
	logger := NewRequestLogger(50)

	logger.Log("GET", "/test", 200, time.Millisecond)
	logger.Log("POST", "/test", 201, time.Millisecond)

	if logger.Count() != 2 {
		t.Errorf("Expected 2 entries before clear, got %d", logger.Count())
	}

	logger.Clear()

	if logger.Count() != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", logger.Count())
	}
}

func TestRequestLogger_Count(t *testing.T) {
	logger := NewRequestLogger(50)

	if logger.Count() != 0 {
		t.Errorf("Expected 0, got %d", logger.Count())
	}

	logger.Log("GET", "/test", 200, time.Millisecond)
	if logger.Count() != 1 {
		t.Errorf("Expected 1, got %d", logger.Count())
	}

	logger.Log("GET", "/test", 200, time.Millisecond)
	if logger.Count() != 2 {
		t.Errorf("Expected 2, got %d", logger.Count())
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{100 * time.Microsecond, "100µs"},
		{15 * time.Millisecond, "15ms"},
		{1500 * time.Millisecond, "1.5s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, want %s", tt.duration, result, tt.expected)
		}
	}
}
