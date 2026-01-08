package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ksick/gcp-api-mock/internal/sqladmin"
	"github.com/ksick/gcp-api-mock/internal/store"
)

func setupTestSQLAdmin() (*SQLAdmin, *store.Store) {
	s := store.New()
	return NewSQLAdmin(s), s
}

// =============================================================================
// Instance Handler Tests
// =============================================================================

func TestSQLAdmin_ListInstances_Empty(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/instances", nil)
	rr := httptest.NewRecorder()

	h.ListInstances(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp sqladmin.InstancesListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Kind != "sql#instancesList" {
		t.Errorf("expected kind 'sql#instancesList', got '%s'", resp.Kind)
	}

	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

func TestSQLAdmin_CreateInstance(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	body := `{"name": "test-instance", "databaseVersion": "MYSQL_8_0", "region": "us-central1"}`
	req := httptest.NewRequest(http.MethodPost, "/sql/v1/projects/test-project/instances", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateInstance(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var op sqladmin.Operation
	if err := json.NewDecoder(rr.Body).Decode(&op); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if op.Kind != "sql#operation" {
		t.Errorf("expected kind 'sql#operation', got '%s'", op.Kind)
	}

	if op.OperationType != "CREATE" {
		t.Errorf("expected operationType 'CREATE', got '%s'", op.OperationType)
	}
}

func TestSQLAdmin_CreateInstance_InvalidJSON(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	req := httptest.NewRequest(http.MethodPost, "/sql/v1/projects/test-project/instances", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateInstance(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestSQLAdmin_CreateInstance_MissingName(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	body := `{"databaseVersion": "MYSQL_8_0"}`
	req := httptest.NewRequest(http.MethodPost, "/sql/v1/projects/test-project/instances", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateInstance(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestSQLAdmin_CreateInstance_Duplicate(t *testing.T) {
	h, s := setupTestSQLAdmin()

	// Create first instance
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	// Try to create duplicate
	body := `{"name": "test-instance"}`
	req := httptest.NewRequest(http.MethodPost, "/sql/v1/projects/test-project/instances", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateInstance(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

func TestSQLAdmin_GetInstance(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/instances/test-instance", nil)
	rr := httptest.NewRecorder()

	h.GetInstance(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var instance sqladmin.DatabaseInstance
	if err := json.NewDecoder(rr.Body).Decode(&instance); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if instance.Name != "test-instance" {
		t.Errorf("expected name 'test-instance', got '%s'", instance.Name)
	}
}

func TestSQLAdmin_GetInstance_NotFound(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/instances/non-existent", nil)
	rr := httptest.NewRecorder()

	h.GetInstance(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestSQLAdmin_UpdateInstance(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	body := `{"settings": {"tier": "db-n1-standard-2", "userLabels": {"env": "test"}}}`
	req := httptest.NewRequest(http.MethodPatch, "/sql/v1/projects/test-project/instances/test-instance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateInstance(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var op sqladmin.Operation
	if err := json.NewDecoder(rr.Body).Decode(&op); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if op.OperationType != "UPDATE" {
		t.Errorf("expected operationType 'UPDATE', got '%s'", op.OperationType)
	}

	// Verify instance was updated
	instance := s.GetSQLInstance("test-instance")
	if instance.Settings.Tier != "db-n1-standard-2" {
		t.Errorf("expected tier 'db-n1-standard-2', got '%s'", instance.Settings.Tier)
	}
}

func TestSQLAdmin_UpdateInstance_NotFound(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	body := `{"settings": {"tier": "db-n1-standard-2"}}`
	req := httptest.NewRequest(http.MethodPatch, "/sql/v1/projects/test-project/instances/non-existent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateInstance(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestSQLAdmin_DeleteInstance(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	req := httptest.NewRequest(http.MethodDelete, "/sql/v1/projects/test-project/instances/test-instance", nil)
	rr := httptest.NewRecorder()

	h.DeleteInstance(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var op sqladmin.Operation
	if err := json.NewDecoder(rr.Body).Decode(&op); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if op.OperationType != "DELETE" {
		t.Errorf("expected operationType 'DELETE', got '%s'", op.OperationType)
	}

	// Verify instance was deleted
	if s.GetSQLInstance("test-instance") != nil {
		t.Error("expected instance to be deleted")
	}
}

func TestSQLAdmin_DeleteInstance_NotFound(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	req := httptest.NewRequest(http.MethodDelete, "/sql/v1/projects/test-project/instances/non-existent", nil)
	rr := httptest.NewRecorder()

	h.DeleteInstance(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// =============================================================================
// Database Handler Tests
// =============================================================================

func TestSQLAdmin_ListDatabases(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/instances/test-instance/databases", nil)
	rr := httptest.NewRecorder()

	h.ListDatabases(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp sqladmin.DatabasesListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Kind != "sql#databasesList" {
		t.Errorf("expected kind 'sql#databasesList', got '%s'", resp.Kind)
	}

	// Should have the default 'mysql' database
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
}

func TestSQLAdmin_CreateDatabase(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	body := `{"name": "mydb", "charset": "utf8mb4", "collation": "utf8mb4_general_ci"}`
	req := httptest.NewRequest(http.MethodPost, "/sql/v1/projects/test-project/instances/test-instance/databases", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateDatabase(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var op sqladmin.Operation
	if err := json.NewDecoder(rr.Body).Decode(&op); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if op.OperationType != "CREATE_DATABASE" {
		t.Errorf("expected operationType 'CREATE_DATABASE', got '%s'", op.OperationType)
	}
}

func TestSQLAdmin_CreateDatabase_InstanceNotFound(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	body := `{"name": "mydb"}`
	req := httptest.NewRequest(http.MethodPost, "/sql/v1/projects/test-project/instances/non-existent/databases", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateDatabase(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestSQLAdmin_GetDatabase(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "mydb"})

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/instances/test-instance/databases/mydb", nil)
	rr := httptest.NewRecorder()

	h.GetDatabase(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var db sqladmin.Database
	if err := json.NewDecoder(rr.Body).Decode(&db); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if db.Name != "mydb" {
		t.Errorf("expected name 'mydb', got '%s'", db.Name)
	}
}

func TestSQLAdmin_GetDatabase_NotFound(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/instances/test-instance/databases/non-existent", nil)
	rr := httptest.NewRecorder()

	h.GetDatabase(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestSQLAdmin_DeleteDatabase(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "mydb"})

	req := httptest.NewRequest(http.MethodDelete, "/sql/v1/projects/test-project/instances/test-instance/databases/mydb", nil)
	rr := httptest.NewRecorder()

	h.DeleteDatabase(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify database was deleted
	if s.GetSQLDatabase("test-instance", "mydb") != nil {
		t.Error("expected database to be deleted")
	}
}

// =============================================================================
// User Handler Tests
// =============================================================================

func TestSQLAdmin_ListUsers(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/instances/test-instance/users", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp sqladmin.UsersListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Kind != "sql#usersList" {
		t.Errorf("expected kind 'sql#usersList', got '%s'", resp.Kind)
	}

	// Should have the default 'root' user
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
}

func TestSQLAdmin_CreateUser(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	body := `{"name": "testuser", "password": "secret123", "host": "%"}`
	req := httptest.NewRequest(http.MethodPost, "/sql/v1/projects/test-project/instances/test-instance/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var op sqladmin.Operation
	if err := json.NewDecoder(rr.Body).Decode(&op); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if op.OperationType != "CREATE_USER" {
		t.Errorf("expected operationType 'CREATE_USER', got '%s'", op.OperationType)
	}
}

func TestSQLAdmin_CreateUser_InstanceNotFound(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	body := `{"name": "testuser"}`
	req := httptest.NewRequest(http.MethodPost, "/sql/v1/projects/test-project/instances/non-existent/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestSQLAdmin_UpdateUser(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "testuser", Host: "%"})

	body := `{"password": "newpassword"}`
	req := httptest.NewRequest(http.MethodPut, "/sql/v1/projects/test-project/instances/test-instance/users?name=testuser&host=%25", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var op sqladmin.Operation
	if err := json.NewDecoder(rr.Body).Decode(&op); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if op.OperationType != "UPDATE_USER" {
		t.Errorf("expected operationType 'UPDATE_USER', got '%s'", op.OperationType)
	}
}

func TestSQLAdmin_DeleteUser(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "testuser", Host: "%"})

	req := httptest.NewRequest(http.MethodDelete, "/sql/v1/projects/test-project/instances/test-instance/users?name=testuser&host=%25", nil)
	rr := httptest.NewRecorder()

	h.DeleteUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify user was deleted
	if s.GetSQLUser("test-instance", "testuser", "%") != nil {
		t.Error("expected user to be deleted")
	}
}

// =============================================================================
// Operation Handler Tests
// =============================================================================

func TestSQLAdmin_ListOperations(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/operations", nil)
	rr := httptest.NewRecorder()

	h.ListOperations(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp sqladmin.OperationsListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Kind != "sql#operationsList" {
		t.Errorf("expected kind 'sql#operationsList', got '%s'", resp.Kind)
	}

	// Should have at least the CREATE operation
	if len(resp.Items) < 1 {
		t.Errorf("expected at least 1 operation, got %d", len(resp.Items))
	}
}

func TestSQLAdmin_ListOperations_FilterByInstance(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance-1"})
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance-2"})

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/operations?instance=test-instance-1", nil)
	rr := httptest.NewRecorder()

	h.ListOperations(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp sqladmin.OperationsListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// All operations should be for test-instance-1
	for _, op := range resp.Items {
		if op.TargetId != "test-instance-1" {
			t.Errorf("expected all operations to be for test-instance-1, got %s", op.TargetId)
		}
	}
}

func TestSQLAdmin_GetOperation(t *testing.T) {
	h, s := setupTestSQLAdmin()
	_, op, _ := s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/operations/"+op.Name, nil)
	rr := httptest.NewRecorder()

	h.GetOperation(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var result sqladmin.Operation
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Name != op.Name {
		t.Errorf("expected name '%s', got '%s'", op.Name, result.Name)
	}
}

func TestSQLAdmin_GetOperation_NotFound(t *testing.T) {
	h, _ := setupTestSQLAdmin()

	req := httptest.NewRequest(http.MethodGet, "/sql/v1/projects/test-project/operations/non-existent", nil)
	rr := httptest.NewRecorder()

	h.GetOperation(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestExtractSQLInstanceName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "standard path",
			path:     "/sql/v1/projects/test-project/instances/my-instance",
			expected: "my-instance",
		},
		{
			name:     "path with trailing slash",
			path:     "/sql/v1/projects/test-project/instances/my-instance/",
			expected: "my-instance",
		},
		{
			name:     "path with databases",
			path:     "/sql/v1/projects/test-project/instances/my-instance/databases",
			expected: "my-instance",
		},
		{
			name:     "no instances in path",
			path:     "/sql/v1/projects/test-project/operations/my-op",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSQLInstanceName(tt.path)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExtractInstanceAndDatabase(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		expectedInstance string
		expectedDB       string
	}{
		{
			name:             "standard path",
			path:             "/sql/v1/projects/test-project/instances/my-instance/databases/mydb",
			expectedInstance: "my-instance",
			expectedDB:       "mydb",
		},
		{
			name:             "no database",
			path:             "/sql/v1/projects/test-project/instances/my-instance/databases",
			expectedInstance: "",
			expectedDB:       "",
		},
		{
			name:             "no instances marker",
			path:             "/sql/v1/projects/test-project/databases/mydb",
			expectedInstance: "",
			expectedDB:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, db := extractInstanceAndDatabase(tt.path)
			if instance != tt.expectedInstance || db != tt.expectedDB {
				t.Errorf("expected ('%s', '%s'), got ('%s', '%s')", tt.expectedInstance, tt.expectedDB, instance, db)
			}
		})
	}
}

func TestExtractOperationName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "standard path",
			path:     "/sql/v1/projects/test-project/operations/op-12345",
			expected: "op-12345",
		},
		{
			name:     "path with trailing content",
			path:     "/sql/v1/projects/test-project/operations/op-12345/extra",
			expected: "op-12345",
		},
		{
			name:     "no operations in path",
			path:     "/sql/v1/projects/test-project/instances/my-instance",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOperationName(tt.path)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
