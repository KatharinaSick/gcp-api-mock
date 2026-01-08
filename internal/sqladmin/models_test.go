package sqladmin

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDatabaseInstanceKindConstant(t *testing.T) {
	instance := &DatabaseInstance{
		Kind: "sql#instance",
		Name: "test-instance",
	}

	if instance.Kind != "sql#instance" {
		t.Errorf("expected kind 'sql#instance', got '%s'", instance.Kind)
	}
}

func TestDatabaseKindConstant(t *testing.T) {
	db := &Database{
		Kind: "sql#database",
		Name: "test-database",
	}

	if db.Kind != "sql#database" {
		t.Errorf("expected kind 'sql#database', got '%s'", db.Kind)
	}
}

func TestUserKindConstant(t *testing.T) {
	user := &User{
		Kind: "sql#user",
		Name: "test-user",
	}

	if user.Kind != "sql#user" {
		t.Errorf("expected kind 'sql#user', got '%s'", user.Kind)
	}
}

func TestOperationKindConstant(t *testing.T) {
	op := &Operation{
		Kind: "sql#operation",
		Name: "test-operation",
	}

	if op.Kind != "sql#operation" {
		t.Errorf("expected kind 'sql#operation', got '%s'", op.Kind)
	}
}

func TestInstancesListResponseKindConstant(t *testing.T) {
	list := &InstancesListResponse{
		Kind:  "sql#instancesList",
		Items: []*DatabaseInstance{},
	}

	if list.Kind != "sql#instancesList" {
		t.Errorf("expected kind 'sql#instancesList', got '%s'", list.Kind)
	}
}

func TestDatabasesListResponseKindConstant(t *testing.T) {
	list := &DatabasesListResponse{
		Kind:  "sql#databasesList",
		Items: []*Database{},
	}

	if list.Kind != "sql#databasesList" {
		t.Errorf("expected kind 'sql#databasesList', got '%s'", list.Kind)
	}
}

func TestUsersListResponseKindConstant(t *testing.T) {
	list := &UsersListResponse{
		Kind:  "sql#usersList",
		Items: []*User{},
	}

	if list.Kind != "sql#usersList" {
		t.Errorf("expected kind 'sql#usersList', got '%s'", list.Kind)
	}
}

func TestOperationsListResponseKindConstant(t *testing.T) {
	list := &OperationsListResponse{
		Kind:  "sql#operationsList",
		Items: []*Operation{},
	}

	if list.Kind != "sql#operationsList" {
		t.Errorf("expected kind 'sql#operationsList', got '%s'", list.Kind)
	}
}

func TestDatabaseInstanceJSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	instance := &DatabaseInstance{
		Kind:            "sql#instance",
		Name:            "test-instance",
		State:           "RUNNABLE",
		DatabaseVersion: "MYSQL_8_0",
		Region:          "us-central1",
		Project:         "test-project",
		BackendType:     "SECOND_GEN",
		InstanceType:    "CLOUD_SQL_INSTANCE",
		CreateTime:      now,
		Settings: &Settings{
			Kind:             "sql#settings",
			Tier:             "db-n1-standard-1",
			AvailabilityType: "ZONAL",
			DataDiskSizeGb:   10,
		},
	}

	data, err := json.Marshal(instance)
	if err != nil {
		t.Fatalf("failed to marshal instance: %v", err)
	}

	var decoded DatabaseInstance
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal instance: %v", err)
	}

	if decoded.Name != instance.Name {
		t.Errorf("expected name '%s', got '%s'", instance.Name, decoded.Name)
	}

	if decoded.Settings == nil {
		t.Fatal("expected settings to be non-nil")
	}

	if decoded.Settings.Tier != instance.Settings.Tier {
		t.Errorf("expected tier '%s', got '%s'", instance.Settings.Tier, decoded.Settings.Tier)
	}
}

func TestDatabaseJSONSerialization(t *testing.T) {
	db := &Database{
		Kind:      "sql#database",
		Name:      "test-db",
		Charset:   "utf8",
		Collation: "utf8_general_ci",
		Instance:  "test-instance",
		Project:   "test-project",
		SelfLink:  "http://localhost:8080/sql/v1/projects/test-project/instances/test-instance/databases/test-db",
	}

	data, err := json.Marshal(db)
	if err != nil {
		t.Fatalf("failed to marshal database: %v", err)
	}

	var decoded Database
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal database: %v", err)
	}

	if decoded.Name != db.Name {
		t.Errorf("expected name '%s', got '%s'", db.Name, decoded.Name)
	}

	if decoded.Charset != db.Charset {
		t.Errorf("expected charset '%s', got '%s'", db.Charset, decoded.Charset)
	}
}

func TestUserJSONSerialization(t *testing.T) {
	user := &User{
		Kind:     "sql#user",
		Name:     "test-user",
		Host:     "%",
		Instance: "test-instance",
		Project:  "test-project",
		Type:     "BUILT_IN",
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal user: %v", err)
	}

	if decoded.Name != user.Name {
		t.Errorf("expected name '%s', got '%s'", user.Name, decoded.Name)
	}

	if decoded.Host != user.Host {
		t.Errorf("expected host '%s', got '%s'", user.Host, decoded.Host)
	}
}

func TestOperationJSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	op := &Operation{
		Kind:          "sql#operation",
		Name:          "operation-123",
		Status:        "DONE",
		OperationType: "CREATE",
		InsertTime:    now,
		StartTime:     now,
		EndTime:       now,
		TargetProject: "test-project",
		TargetId:      "test-instance",
		SelfLink:      "http://localhost:8080/sql/v1/projects/test-project/operations/operation-123",
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("failed to marshal operation: %v", err)
	}

	var decoded Operation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal operation: %v", err)
	}

	if decoded.Name != op.Name {
		t.Errorf("expected name '%s', got '%s'", op.Name, decoded.Name)
	}

	if decoded.Status != op.Status {
		t.Errorf("expected status '%s', got '%s'", op.Status, decoded.Status)
	}
}

func TestSettingsJSONSerialization(t *testing.T) {
	settings := &Settings{
		Kind:                      "sql#settings",
		SettingsVersion:           1,
		Tier:                      "db-n1-standard-1",
		AvailabilityType:          "ZONAL",
		PricingPlan:               "PER_USE",
		ActivationPolicy:          "ALWAYS",
		StorageAutoResize:         true,
		StorageAutoResizeLimit:    0,
		DataDiskType:              "PD_SSD",
		DataDiskSizeGb:            10,
		DeletionProtectionEnabled: false,
		UserLabels: map[string]string{
			"env": "test",
		},
		IPConfiguration: &IPConfiguration{
			IPv4Enabled: true,
		},
		BackupConfiguration: &BackupConfiguration{
			Kind:    "sql#backupConfiguration",
			Enabled: true,
		},
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("failed to marshal settings: %v", err)
	}

	var decoded Settings
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	if decoded.Tier != settings.Tier {
		t.Errorf("expected tier '%s', got '%s'", settings.Tier, decoded.Tier)
	}

	if decoded.DataDiskSizeGb != settings.DataDiskSizeGb {
		t.Errorf("expected dataDiskSizeGb '%d', got '%d'", settings.DataDiskSizeGb, decoded.DataDiskSizeGb)
	}

	if decoded.UserLabels["env"] != "test" {
		t.Errorf("expected userLabels.env 'test', got '%s'", decoded.UserLabels["env"])
	}
}

func TestAPIErrorJSONSerialization(t *testing.T) {
	apiErr := &APIError{
		Error: ErrorDetails{
			Code:    404,
			Message: "Instance not found",
			Status:  "NOT_FOUND",
			Errors: []ErrorReason{
				{
					Domain:  "global",
					Reason:  "notFound",
					Message: "Instance not found",
				},
			},
		},
	}

	data, err := json.Marshal(apiErr)
	if err != nil {
		t.Fatalf("failed to marshal API error: %v", err)
	}

	var decoded APIError
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal API error: %v", err)
	}

	if decoded.Error.Code != 404 {
		t.Errorf("expected code 404, got %d", decoded.Error.Code)
	}

	if len(decoded.Error.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(decoded.Error.Errors))
	}
}

func TestIPMappingJSONSerialization(t *testing.T) {
	ip := &IPMapping{
		Type:      "PRIMARY",
		IPAddress: "10.0.0.1",
	}

	data, err := json.Marshal(ip)
	if err != nil {
		t.Fatalf("failed to marshal IP mapping: %v", err)
	}

	var decoded IPMapping
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal IP mapping: %v", err)
	}

	if decoded.Type != ip.Type {
		t.Errorf("expected type '%s', got '%s'", ip.Type, decoded.Type)
	}

	if decoded.IPAddress != ip.IPAddress {
		t.Errorf("expected IP address '%s', got '%s'", ip.IPAddress, decoded.IPAddress)
	}
}
