package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for LoggingMiddleware

func TestLoggingMiddleware_LogsRequestDetails(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := LoggingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	logOutput := logBuf.String()

	// Verify log contains method, path, status code
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test/path")
	assert.Contains(t, logOutput, "200")
}

func TestLoggingMiddleware_LogsCorrectStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"BadRequest", http.StatusBadRequest},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			log.SetOutput(&logBuf)
			defer log.SetOutput(os.Stderr)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			wrapped := LoggingMiddleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			logOutput := logBuf.String()
			assert.Contains(t, logOutput, string(rune(tt.statusCode/100+'0')))
		})
	}
}

func TestLoggingMiddleware_DefaultStatusCodeIs200(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	// Handler that doesn't explicitly set status code
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	wrapped := LoggingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "200")
}

func TestLoggingMiddleware_LogsDuration(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := LoggingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	logOutput := logBuf.String()
	// Duration should contain time units (µs, ms, s, etc.)
	assert.True(t, strings.Contains(logOutput, "µs") || 
		strings.Contains(logOutput, "ms") || 
		strings.Contains(logOutput, "ns") ||
		strings.Contains(logOutput, "s"), 
		"Log should contain duration: %s", logOutput)
}

func TestLoggingMiddleware_PassesRequestToNextHandler(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Suppress log output for this test
	log.SetOutput(&bytes.Buffer{})
	defer log.SetOutput(os.Stderr)

	wrapped := LoggingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "Handler should have been called")
}

// Tests for CORSMiddleware

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORSMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "PUT")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "DELETE")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "PATCH")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestCORSMiddleware_HandlesOptionsRequest(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORSMiddleware(handler)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, handlerCalled, "Handler should not be called for OPTIONS requests")
}

func TestCORSMiddleware_PassesNonOptionsToNextHandler(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handlerCalled := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			wrapped := CORSMiddleware(handler)

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			assert.True(t, handlerCalled, "Handler should be called for %s requests", method)
		})
	}
}

// Tests for WriteGCPError

func TestWriteGCPError_FormatsCorrectly(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteGCPError(rec, http.StatusNotFound, "The specified bucket does not exist.", "notFound")

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json; charset=UTF-8", rec.Header().Get("Content-Type"))

	var gcpErr GCPError
	err := json.NewDecoder(rec.Body).Decode(&gcpErr)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, gcpErr.Error.Code)
	assert.Equal(t, "The specified bucket does not exist.", gcpErr.Error.Message)
	require.Len(t, gcpErr.Error.Errors, 1)
	assert.Equal(t, "The specified bucket does not exist.", gcpErr.Error.Errors[0].Message)
	assert.Equal(t, "global", gcpErr.Error.Errors[0].Domain)
	assert.Equal(t, "notFound", gcpErr.Error.Errors[0].Reason)
}

func TestWriteGCPError_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		reason  string
	}{
		{
			name:    "BadRequest",
			code:    http.StatusBadRequest,
			message: "Invalid request",
			reason:  "badRequest",
		},
		{
			name:    "Unauthorized",
			code:    http.StatusUnauthorized,
			message: "Authentication required",
			reason:  "unauthorized",
		},
		{
			name:    "Forbidden",
			code:    http.StatusForbidden,
			message: "Access denied",
			reason:  "forbidden",
		},
		{
			name:    "NotFound",
			code:    http.StatusNotFound,
			message: "Resource not found",
			reason:  "notFound",
		},
		{
			name:    "InternalServerError",
			code:    http.StatusInternalServerError,
			message: "Internal server error",
			reason:  "internalError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			WriteGCPError(rec, tt.code, tt.message, tt.reason)

			assert.Equal(t, tt.code, rec.Code)

			var gcpErr GCPError
			err := json.NewDecoder(rec.Body).Decode(&gcpErr)
			require.NoError(t, err)

			assert.Equal(t, tt.code, gcpErr.Error.Code)
			assert.Equal(t, tt.message, gcpErr.Error.Message)
			assert.Equal(t, tt.reason, gcpErr.Error.Errors[0].Reason)
		})
	}
}

func TestWriteGCPError_JSONStructure(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteGCPError(rec, http.StatusNotFound, "Test message", "testReason")

	// Parse as raw JSON to verify structure
	var rawJSON map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&rawJSON)
	require.NoError(t, err)

	// Verify top-level "error" key exists
	errorObj, ok := rawJSON["error"].(map[string]interface{})
	require.True(t, ok, "Response should have 'error' object at top level")

	// Verify error object structure
	assert.Contains(t, errorObj, "code")
	assert.Contains(t, errorObj, "message")
	assert.Contains(t, errorObj, "errors")

	// Verify errors array
	errorsArray, ok := errorObj["errors"].([]interface{})
	require.True(t, ok, "errors should be an array")
	require.Len(t, errorsArray, 1)

	// Verify error item structure
	errorItem, ok := errorsArray[0].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, errorItem, "message")
	assert.Contains(t, errorItem, "domain")
	assert.Contains(t, errorItem, "reason")
}

// Tests for responseWriter

func TestResponseWriter_CapturesStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	wrapped.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, wrapped.statusCode)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestResponseWriter_DefaultStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	// Don't call WriteHeader, status should remain default
	assert.Equal(t, http.StatusOK, wrapped.statusCode)
}

func TestResponseWriter_WriteDelegates(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	n, err := wrapped.Write([]byte("test content"))

	assert.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, "test content", rec.Body.String())
}

// Tests for GCPError types

func TestGCPError_Serialization(t *testing.T) {
	gcpErr := GCPError{
		Error: GCPErrorDetail{
			Code:    404,
			Message: "Not found",
			Errors: []GCPErrorItem{
				{
					Message: "Not found",
					Domain:  "global",
					Reason:  "notFound",
				},
			},
		},
	}

	data, err := json.Marshal(gcpErr)
	require.NoError(t, err)

	var decoded GCPError
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, gcpErr.Error.Code, decoded.Error.Code)
	assert.Equal(t, gcpErr.Error.Message, decoded.Error.Message)
	assert.Equal(t, gcpErr.Error.Errors[0].Reason, decoded.Error.Errors[0].Reason)
}

// Test middleware chaining

func TestMiddleware_Chaining(t *testing.T) {
	// Suppress log output
	log.SetOutput(&bytes.Buffer{})
	defer log.SetOutput(os.Stderr)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Chain middleware: CORS -> Logging -> Handler
	wrapped := CORSMiddleware(LoggingMiddleware(handler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())

	// Verify CORS headers are set
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestMiddleware_ChainingWithOptions(t *testing.T) {
	// Suppress log output
	log.SetOutput(&bytes.Buffer{})
	defer log.SetOutput(os.Stderr)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Chain middleware: CORS -> Logging -> Handler
	wrapped := CORSMiddleware(LoggingMiddleware(handler))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// OPTIONS should be handled by CORS middleware
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, handlerCalled)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

// Tests for WriteGCPJSON

func TestWriteGCPJSON_FormatsCorrectly(t *testing.T) {
	rec := httptest.NewRecorder()

	data := map[string]interface{}{
		"kind":  "storage#bucket",
		"name":  "test-bucket",
		"count": 42,
	}

	WriteGCPJSON(rec, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=UTF-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, max-age=0, must-revalidate", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "Origin, X-Origin", rec.Header().Get("Vary"))
	assert.NotEmpty(t, rec.Header().Get("X-GUploader-UploadID"))

	var result map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "storage#bucket", result["kind"])
	assert.Equal(t, "test-bucket", result["name"])
	assert.Equal(t, float64(42), result["count"])
}

func TestWriteGCPJSON_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"Accepted", http.StatusAccepted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			data := map[string]string{"status": "ok"}

			WriteGCPJSON(rec, tt.code, data)

			assert.Equal(t, tt.code, rec.Code)
		})
	}
}

// Tests for WriteGCPNoContent

func TestWriteGCPNoContent_ReturnsNoContent(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteGCPNoContent(rec)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
}
