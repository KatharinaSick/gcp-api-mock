// Package handler provides HTTP handlers for the GCP API Mock.
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/katharinasick/gcp-api-mock/internal/sqladmin"
	"github.com/katharinasick/gcp-api-mock/internal/store"
)

// SQLAdmin handles Cloud SQL Admin API endpoints.
type SQLAdmin struct {
	store *store.Store
}

// NewSQLAdmin creates a new SQLAdmin handler.
func NewSQLAdmin(s *store.Store) *SQLAdmin {
	return &SQLAdmin{store: s}
}

// =============================================================================
// Instance Handlers
// =============================================================================

// ListInstances handles GET /sql/v1beta4/projects/{project}/instances - List instances.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/instances/list
func (h *SQLAdmin) ListInstances(w http.ResponseWriter, r *http.Request) {
	instances := h.store.ListSQLInstances()

	response := &sqladmin.InstancesListResponse{
		Kind:  "sql#instancesList",
		Items: instances,
	}

	respondSQLJSON(w, http.StatusOK, response)
}

// CreateInstance handles POST /sql/v1beta4/projects/{project}/instances - Create an instance.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/instances/insert
func (h *SQLAdmin) CreateInstance(w http.ResponseWriter, r *http.Request) {
	var req sqladmin.InstanceInsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondSQLError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_ARGUMENT", "invalid")
		return
	}

	if req.Name == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	_, op, err := h.store.CreateSQLInstance(&req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			respondSQLError(w, http.StatusConflict, err.Error(), "ALREADY_EXISTS", "conflict")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// GetInstance handles GET /sql/v1beta4/projects/{project}/instances/{instance} - Get instance.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/instances/get
func (h *SQLAdmin) GetInstance(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceName(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	instance := h.store.GetSQLInstance(instanceName)
	if instance == nil {
		respondSQLError(w, http.StatusNotFound, "Instance not found", "NOT_FOUND", "notFound")
		return
	}

	respondSQLJSON(w, http.StatusOK, instance)
}

// UpdateInstance handles PATCH /sql/v1beta4/projects/{project}/instances/{instance} - Update instance.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/instances/patch
func (h *SQLAdmin) UpdateInstance(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceName(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	var req sqladmin.InstancePatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondSQLError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_ARGUMENT", "invalid")
		return
	}

	_, op, err := h.store.UpdateSQLInstance(instanceName, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// DeleteInstance handles DELETE /sql/v1beta4/projects/{project}/instances/{instance} - Delete instance.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/instances/delete
func (h *SQLAdmin) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceName(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	op, err := h.store.DeleteSQLInstance(instanceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		if strings.Contains(err.Error(), "deletion protection") {
			respondSQLError(w, http.StatusBadRequest, err.Error(), "FAILED_PRECONDITION", "failedPrecondition")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// =============================================================================
// Database Handlers
// =============================================================================

// ListDatabases handles GET /sql/v1beta4/projects/{project}/instances/{instance}/databases - List databases.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/databases/list
func (h *SQLAdmin) ListDatabases(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceNameFromDatabasePath(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	databases, err := h.store.ListSQLDatabases(instanceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	response := &sqladmin.DatabasesListResponse{
		Kind:  "sql#databasesList",
		Items: databases,
	}

	respondSQLJSON(w, http.StatusOK, response)
}

// CreateDatabase handles POST /sql/v1beta4/projects/{project}/instances/{instance}/databases - Create database.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/databases/insert
func (h *SQLAdmin) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceNameFromDatabasePath(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	var req sqladmin.DatabaseInsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondSQLError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_ARGUMENT", "invalid")
		return
	}

	if req.Name == "" {
		respondSQLError(w, http.StatusBadRequest, "Database name is required", "INVALID_ARGUMENT", "required")
		return
	}

	_, op, err := h.store.CreateSQLDatabase(instanceName, &req)
	if err != nil {
		if strings.Contains(err.Error(), "instance") && strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			respondSQLError(w, http.StatusConflict, err.Error(), "ALREADY_EXISTS", "conflict")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// GetDatabase handles GET /sql/v1beta4/projects/{project}/instances/{instance}/databases/{database} - Get database.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/databases/get
func (h *SQLAdmin) GetDatabase(w http.ResponseWriter, r *http.Request) {
	instanceName, dbName := extractInstanceAndDatabase(r.URL.Path)

	if instanceName == "" || dbName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance and database names are required", "INVALID_ARGUMENT", "required")
		return
	}

	db := h.store.GetSQLDatabase(instanceName, dbName)
	if db == nil {
		respondSQLError(w, http.StatusNotFound, "Database not found", "NOT_FOUND", "notFound")
		return
	}

	respondSQLJSON(w, http.StatusOK, db)
}

// UpdateDatabase handles PATCH /sql/v1beta4/projects/{project}/instances/{instance}/databases/{database} - Update database.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/databases/patch
func (h *SQLAdmin) UpdateDatabase(w http.ResponseWriter, r *http.Request) {
	instanceName, dbName := extractInstanceAndDatabase(r.URL.Path)

	if instanceName == "" || dbName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance and database names are required", "INVALID_ARGUMENT", "required")
		return
	}

	var req sqladmin.DatabasePatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondSQLError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_ARGUMENT", "invalid")
		return
	}

	_, op, err := h.store.UpdateSQLDatabase(instanceName, dbName, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// DeleteDatabase handles DELETE /sql/v1beta4/projects/{project}/instances/{instance}/databases/{database} - Delete database.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/databases/delete
func (h *SQLAdmin) DeleteDatabase(w http.ResponseWriter, r *http.Request) {
	instanceName, dbName := extractInstanceAndDatabase(r.URL.Path)

	if instanceName == "" || dbName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance and database names are required", "INVALID_ARGUMENT", "required")
		return
	}

	op, err := h.store.DeleteSQLDatabase(instanceName, dbName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// =============================================================================
// User Handlers
// =============================================================================

// ListUsers handles GET /sql/v1beta4/projects/{project}/instances/{instance}/users - List users.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/users/list
func (h *SQLAdmin) ListUsers(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceNameFromUsersPath(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	users, err := h.store.ListSQLUsers(instanceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	response := &sqladmin.UsersListResponse{
		Kind:  "sql#usersList",
		Items: users,
	}

	respondSQLJSON(w, http.StatusOK, response)
}

// CreateUser handles POST /sql/v1beta4/projects/{project}/instances/{instance}/users - Create user.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/users/insert
func (h *SQLAdmin) CreateUser(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceNameFromUsersPath(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	var req sqladmin.UserInsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondSQLError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_ARGUMENT", "invalid")
		return
	}

	if req.Name == "" {
		respondSQLError(w, http.StatusBadRequest, "User name is required", "INVALID_ARGUMENT", "required")
		return
	}

	_, op, err := h.store.CreateSQLUser(instanceName, &req)
	if err != nil {
		if strings.Contains(err.Error(), "instance") && strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			respondSQLError(w, http.StatusConflict, err.Error(), "ALREADY_EXISTS", "conflict")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// UpdateUser handles PUT /sql/v1beta4/projects/{project}/instances/{instance}/users - Update user.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/users/update
func (h *SQLAdmin) UpdateUser(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceNameFromUsersPath(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	// Get name and host from query parameters
	userName := r.URL.Query().Get("name")
	host := r.URL.Query().Get("host")

	if userName == "" {
		respondSQLError(w, http.StatusBadRequest, "User name is required", "INVALID_ARGUMENT", "required")
		return
	}

	var req sqladmin.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondSQLError(w, http.StatusBadRequest, "Invalid JSON body", "INVALID_ARGUMENT", "invalid")
		return
	}

	_, op, err := h.store.UpdateSQLUser(instanceName, userName, host, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// DeleteUser handles DELETE /sql/v1beta4/projects/{project}/instances/{instance}/users - Delete user.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/users/delete
func (h *SQLAdmin) DeleteUser(w http.ResponseWriter, r *http.Request) {
	instanceName := extractSQLInstanceNameFromUsersPath(r.URL.Path)

	if instanceName == "" {
		respondSQLError(w, http.StatusBadRequest, "Instance name is required", "INVALID_ARGUMENT", "required")
		return
	}

	// Get name and host from query parameters
	userName := r.URL.Query().Get("name")
	host := r.URL.Query().Get("host")

	if userName == "" {
		respondSQLError(w, http.StatusBadRequest, "User name is required", "INVALID_ARGUMENT", "required")
		return
	}

	op, err := h.store.DeleteSQLUser(instanceName, userName, host)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondSQLError(w, http.StatusNotFound, err.Error(), "NOT_FOUND", "notFound")
			return
		}
		respondSQLError(w, http.StatusInternalServerError, err.Error(), "INTERNAL", "internalError")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// =============================================================================
// Operation Handlers
// =============================================================================

// ListOperations handles GET /sql/v1beta4/projects/{project}/operations - List operations.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/operations/list
func (h *SQLAdmin) ListOperations(w http.ResponseWriter, r *http.Request) {
	instanceName := r.URL.Query().Get("instance")

	operations := h.store.ListSQLOperations(instanceName)

	response := &sqladmin.OperationsListResponse{
		Kind:  "sql#operationsList",
		Items: operations,
	}

	respondSQLJSON(w, http.StatusOK, response)
}

// GetOperation handles GET /sql/v1beta4/projects/{project}/operations/{operation} - Get operation.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/operations/get
func (h *SQLAdmin) GetOperation(w http.ResponseWriter, r *http.Request) {
	opName := extractOperationName(r.URL.Path)

	if opName == "" {
		respondSQLError(w, http.StatusBadRequest, "Operation name is required", "INVALID_ARGUMENT", "required")
		return
	}

	op := h.store.GetSQLOperation(opName)
	if op == nil {
		respondSQLError(w, http.StatusNotFound, "Operation not found", "NOT_FOUND", "notFound")
		return
	}

	respondSQLJSON(w, http.StatusOK, op)
}

// =============================================================================
// Helper Functions
// =============================================================================

// respondSQLJSON writes a JSON response.
func respondSQLJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondSQLError writes a JSON error response matching the Cloud SQL Admin API format.
func respondSQLError(w http.ResponseWriter, statusCode int, message, status, reason string) {
	errResp := sqladmin.APIError{
		Error: sqladmin.ErrorDetails{
			Code:    statusCode,
			Message: message,
			Status:  status,
			Errors: []sqladmin.ErrorReason{
				{
					Domain:  "global",
					Reason:  reason,
					Message: message,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errResp)
}

// extractSQLInstanceName extracts the instance name from a path like /sql/v1/projects/{project}/instances/{instance}.
func extractSQLInstanceName(path string) string {
	// Find /instances/ in the path
	const marker = "/instances/"
	idx := strings.Index(path, marker)
	if idx < 0 {
		return ""
	}

	remaining := path[idx+len(marker):]
	// Remove any trailing path segments
	if slashIdx := strings.Index(remaining, "/"); slashIdx >= 0 {
		remaining = remaining[:slashIdx]
	}
	return remaining
}

// extractSQLInstanceNameFromDatabasePath extracts the instance name from a database path.
func extractSQLInstanceNameFromDatabasePath(path string) string {
	return extractSQLInstanceName(path)
}

// extractSQLInstanceNameFromUsersPath extracts the instance name from a users path.
func extractSQLInstanceNameFromUsersPath(path string) string {
	return extractSQLInstanceName(path)
}

// extractInstanceAndDatabase extracts instance and database names from a path like
// /sql/v1/projects/{project}/instances/{instance}/databases/{database}.
func extractInstanceAndDatabase(path string) (string, string) {
	const instanceMarker = "/instances/"
	const dbMarker = "/databases/"

	instanceIdx := strings.Index(path, instanceMarker)
	dbIdx := strings.Index(path, dbMarker)

	if instanceIdx < 0 || dbIdx < 0 {
		return "", ""
	}

	instancePart := path[instanceIdx+len(instanceMarker) : dbIdx]
	dbPart := path[dbIdx+len(dbMarker):]

	// Remove any trailing path segments from db name
	if slashIdx := strings.Index(dbPart, "/"); slashIdx >= 0 {
		dbPart = dbPart[:slashIdx]
	}

	return instancePart, dbPart
}

// extractOperationName extracts the operation name from a path like
// /sql/v1/projects/{project}/operations/{operation}.
func extractOperationName(path string) string {
	const marker = "/operations/"
	idx := strings.Index(path, marker)
	if idx < 0 {
		return ""
	}

	remaining := path[idx+len(marker):]
	// Remove any trailing path segments
	if slashIdx := strings.Index(remaining, "/"); slashIdx >= 0 {
		remaining = remaining[:slashIdx]
	}
	return remaining
}
