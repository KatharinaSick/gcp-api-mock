package requestid

import (
	"context"
	"testing"
)

func TestGenerate(t *testing.T) {
	id1 := Generate()
	id2 := Generate()

	if id1 == "" {
		t.Error("Generate() returned empty string")
	}

	if len(id1) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id1))
	}

	if id1 == id2 {
		t.Error("Generate() returned same ID twice")
	}
}

func TestFromContext(t *testing.T) {
	t.Run("with request ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ContextKey, "test-id-123")
		id := FromContext(ctx)

		if id != "test-id-123" {
			t.Errorf("expected 'test-id-123', got '%s'", id)
		}
	})

	t.Run("without request ID", func(t *testing.T) {
		ctx := context.Background()
		id := FromContext(ctx)

		if id != "" {
			t.Errorf("expected empty string, got '%s'", id)
		}
	})
}
