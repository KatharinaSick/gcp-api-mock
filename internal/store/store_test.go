package store

import (
	"testing"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Error("New() returned nil")
	}
}

func TestStore_Reset(t *testing.T) {
	s := New()
	// Should not panic when called on empty store
	s.Reset()
}
