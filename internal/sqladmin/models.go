// Package sqladmin provides data models for the Google Cloud SQL Admin API mock.
package sqladmin

import "time"

// DatabaseInstance represents a Cloud SQL database instance.
// Based on the official Cloud SQL Admin API v1 specification.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/instances
type DatabaseInstance struct {
	// Kind is the kind of resource. For instances, this is always "sql#instance".
	Kind string `json:"kind"`
	// State is the current serving state of the instance.
	State string `json:"state"`
	// DatabaseVersion is the database engine type and version (e.g., "MYSQL_8_0").
	DatabaseVersion string `json:"databaseVersion"`
	// Settings contains user settings for the instance.
	Settings *Settings `json:"settings,omitempty"`
	// Etag is the HTTP 1.1 Entity tag for the resource.
	Etag string `json:"etag,omitempty"`
	// MasterInstanceName is the name of the instance acting as master in the replication setup.
	MasterInstanceName string `json:"masterInstanceName,omitempty"`
	// ReplicaNames contains the replicas of the instance.
	ReplicaNames []string `json:"replicaNames,omitempty"`
	// MaxDiskSize is deprecated. This field is not used.
	MaxDiskSize int64 `json:"maxDiskSize,string,omitempty"`
	// CurrentDiskSize is deprecated. This field is not used.
	CurrentDiskSize int64 `json:"currentDiskSize,string,omitempty"`
	// IPAddresses contains the assigned IP addresses for the instance.
	IPAddresses []*IPMapping `json:"ipAddresses,omitempty"`
	// ServerCaCert contains SSL configuration for the instance.
	ServerCaCert *SSLCert `json:"serverCaCert,omitempty"`
	// InstanceType is the instance type.
	InstanceType string `json:"instanceType"`
	// Project is the project ID of the project containing the Cloud SQL instance.
	Project string `json:"project"`
	// IPv6Address is the IPv6 address assigned to the instance (deprecated).
	IPv6Address string `json:"ipv6Address,omitempty"`
	// ServiceAccountEmailAddress is the service account email address assigned to the instance.
	ServiceAccountEmailAddress string `json:"serviceAccountEmailAddress,omitempty"`
	// OnPremisesConfiguration contains configuration specific to on-premises instances.
	OnPremisesConfiguration *OnPremisesConfiguration `json:"onPremisesConfiguration,omitempty"`
	// ReplicaConfiguration contains configuration for failover replicas.
	ReplicaConfiguration *ReplicaConfiguration `json:"replicaConfiguration,omitempty"`
	// BackendType is the backend type. SECOND_GEN is the only valid value.
	BackendType string `json:"backendType"`
	// SelfLink is the URI of this resource.
	SelfLink string `json:"selfLink"`
	// ConnectionName is the connection name of the instance used in connection strings.
	ConnectionName string `json:"connectionName"`
	// Name is the name of the instance which will also be used as the database name.
	Name string `json:"name"`
	// Region is the geographical region.
	Region string `json:"region"`
	// GceZone is the Compute Engine zone that the instance is currently serving from.
	GceZone string `json:"gceZone,omitempty"`
	// SecondaryGceZone is the Compute Engine zone that the failover instance is currently serving from.
	SecondaryGceZone string `json:"secondaryGceZone,omitempty"`
	// DiskEncryptionConfiguration contains disk encryption configuration.
	DiskEncryptionConfiguration *DiskEncryptionConfiguration `json:"diskEncryptionConfiguration,omitempty"`
	// DiskEncryptionStatus contains disk encryption status.
	DiskEncryptionStatus *DiskEncryptionStatus `json:"diskEncryptionStatus,omitempty"`
	// RootPassword is the initial root password (only available on insert).
	RootPassword string `json:"rootPassword,omitempty"`
	// CreateTime is the time when the instance was created in RFC 3339 format.
	CreateTime time.Time `json:"createTime"`
}

// Settings contains database instance settings.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/instances#settings
type Settings struct {
	// SettingsVersion is the version of settings. Required for update.
	SettingsVersion int64 `json:"settingsVersion,string"`
	// AuthorizedGaeApplications contains App Engine app IDs (deprecated).
	AuthorizedGaeApplications []string `json:"authorizedGaeApplications,omitempty"`
	// Tier is the tier (or machine type) for this instance.
	Tier string `json:"tier"`
	// Kind is the kind of resource. This is always "sql#settings".
	Kind string `json:"kind"`
	// UserLabels contains user-provided labels as key:value pairs.
	UserLabels map[string]string `json:"userLabels,omitempty"`
	// AvailabilityType specifies whether the instance should be zonal or regional.
	AvailabilityType string `json:"availabilityType"`
	// PricingPlan is the pricing plan for this instance.
	PricingPlan string `json:"pricingPlan"`
	// ReplicationType is the type of replication this instance uses (deprecated).
	ReplicationType string `json:"replicationType,omitempty"`
	// StorageAutoResizeLimit is the maximum size to which storage capacity can be automatically increased.
	StorageAutoResizeLimit int64 `json:"storageAutoResizeLimit,string,omitempty"`
	// ActivationPolicy specifies when the instance should be activated.
	ActivationPolicy string `json:"activationPolicy"`
	// IPConfiguration contains the settings for IP management.
	IPConfiguration *IPConfiguration `json:"ipConfiguration,omitempty"`
	// StorageAutoResize specifies whether automatic storage increase is enabled.
	StorageAutoResize bool `json:"storageAutoResize"`
	// LocationPreference contains the location preference settings.
	LocationPreference *LocationPreference `json:"locationPreference,omitempty"`
	// DatabaseFlags contains the database flags passed to the instance at startup.
	DatabaseFlags []*DatabaseFlags `json:"databaseFlags,omitempty"`
	// DataDiskType is the type of data disk.
	DataDiskType string `json:"dataDiskType"`
	// MaintenanceWindow contains the maintenance window for the instance.
	MaintenanceWindow *MaintenanceWindow `json:"maintenanceWindow,omitempty"`
	// BackupConfiguration contains the daily backup configuration.
	BackupConfiguration *BackupConfiguration `json:"backupConfiguration,omitempty"`
	// DatabaseReplicationEnabled specifies if database replication is enabled (deprecated).
	DatabaseReplicationEnabled bool `json:"databaseReplicationEnabled,omitempty"`
	// CrashSafeReplicationEnabled specifies if crash-safe replication is enabled (deprecated).
	CrashSafeReplicationEnabled bool `json:"crashSafeReplicationEnabled,omitempty"`
	// DataDiskSizeGb is the size of data disk in GB.
	DataDiskSizeGb int64 `json:"dataDiskSizeGb,string"`
	// ActiveDirectoryConfig contains Active Directory configuration.
	ActiveDirectoryConfig *SqlActiveDirectoryConfig `json:"activeDirectoryConfig,omitempty"`
	// Collation is the server collation.
	Collation string `json:"collation,omitempty"`
	// DenyMaintenancePeriods contains deny maintenance periods.
	DenyMaintenancePeriods []*DenyMaintenancePeriod `json:"denyMaintenancePeriods,omitempty"`
	// InsightsConfig contains Query Insights configuration.
	InsightsConfig *InsightsConfig `json:"insightsConfig,omitempty"`
	// PasswordValidationPolicy contains the local user password validation policy.
	PasswordValidationPolicy *PasswordValidationPolicy `json:"passwordValidationPolicy,omitempty"`
	// SqlServerAuditConfig contains SQL Server specific audit configuration.
	SqlServerAuditConfig *SqlServerAuditConfig `json:"sqlServerAuditConfig,omitempty"`
	// Edition is the edition of the instance.
	Edition string `json:"edition,omitempty"`
	// TimeZone is the time zone of the instance.
	TimeZone string `json:"timeZone,omitempty"`
	// DeletionProtectionEnabled specifies if deletion protection is enabled.
	DeletionProtectionEnabled bool `json:"deletionProtectionEnabled,omitempty"`
}

// IPMapping contains database instance IP mapping.
type IPMapping struct {
	// Type is the type of this IP address.
	Type string `json:"type"`
	// IPAddress is the IP address assigned.
	IPAddress string `json:"ipAddress"`
	// TimeToRetire is the due time for this IP to be retired in RFC 3339 format.
	TimeToRetire time.Time `json:"timeToRetire,omitempty"`
}

// SSLCert contains SslCerts Resource.
type SSLCert struct {
	// Kind is the kind of resource. This is always "sql#sslCert".
	Kind string `json:"kind"`
	// CertSerialNumber is the serial number of the certificate.
	CertSerialNumber string `json:"certSerialNumber"`
	// Cert contains PEM representation.
	Cert string `json:"cert"`
	// CreateTime is the time when the certificate was created in RFC 3339 format.
	CreateTime time.Time `json:"createTime"`
	// CommonName is the user supplied name.
	CommonName string `json:"commonName"`
	// ExpirationTime is the time when the certificate expires in RFC 3339 format.
	ExpirationTime time.Time `json:"expirationTime"`
	// Sha1Fingerprint contains Sha1 Fingerprint.
	Sha1Fingerprint string `json:"sha1Fingerprint"`
	// Instance is the name of the database instance.
	Instance string `json:"instance"`
	// SelfLink is the URI of this resource.
	SelfLink string `json:"selfLink,omitempty"`
}

// OnPremisesConfiguration contains on-premises instance configuration.
type OnPremisesConfiguration struct {
	// HostPort is the host and port of the on-premises instance in host:port format.
	HostPort string `json:"hostPort"`
	// Kind is the kind of resource. This is always "sql#onPremisesConfiguration".
	Kind string `json:"kind"`
}

// ReplicaConfiguration contains read replica configuration.
type ReplicaConfiguration struct {
	// Kind is the kind of resource. This is always "sql#replicaConfiguration".
	Kind string `json:"kind"`
	// MysqlReplicaConfiguration contains MySQL specific replica configuration.
	MysqlReplicaConfiguration *MySQLReplicaConfiguration `json:"mysqlReplicaConfiguration,omitempty"`
	// FailoverTarget specifies if the replica is the failover target.
	FailoverTarget bool `json:"failoverTarget"`
}

// MySQLReplicaConfiguration contains MySQL specific configuration.
type MySQLReplicaConfiguration struct {
	// DumpFilePath is the path to a SQL dump file in GCS.
	DumpFilePath string `json:"dumpFilePath,omitempty"`
	// Username is the replication user account on the master.
	Username string `json:"username,omitempty"`
	// Password is the replication user account password.
	Password string `json:"password,omitempty"`
	// ConnectRetryInterval is the number of seconds between connect retries.
	ConnectRetryInterval int32 `json:"connectRetryInterval,omitempty"`
	// MasterHeartbeatPeriod is the interval in milliseconds between heartbeats to the master.
	MasterHeartbeatPeriod int64 `json:"masterHeartbeatPeriod,string,omitempty"`
	// CaCertificate contains PEM representation of the trusted CA's x509 certificate.
	CaCertificate string `json:"caCertificate,omitempty"`
	// ClientCertificate contains PEM representation of the replica's x509 certificate.
	ClientCertificate string `json:"clientCertificate,omitempty"`
	// ClientKey contains PEM representation of the replica's private key.
	ClientKey string `json:"clientKey,omitempty"`
	// SslCipher is a list of permissible ciphers.
	SslCipher string `json:"sslCipher,omitempty"`
	// VerifyServerCertificate specifies whether to check the server certificate CA.
	VerifyServerCertificate bool `json:"verifyServerCertificate,omitempty"`
	// Kind is the kind of resource. This is always "sql#mysqlReplicaConfiguration".
	Kind string `json:"kind"`
}

// DiskEncryptionConfiguration contains disk encryption configuration.
type DiskEncryptionConfiguration struct {
	// KmsKeyName contains the resource name of KMS key.
	KmsKeyName string `json:"kmsKeyName"`
	// Kind is the kind of resource. This is always "sql#diskEncryptionConfiguration".
	Kind string `json:"kind"`
}

// DiskEncryptionStatus contains disk encryption status.
type DiskEncryptionStatus struct {
	// KmsKeyVersionName contains the KMS key version used for encryption.
	KmsKeyVersionName string `json:"kmsKeyVersionName"`
	// Kind is the kind of resource. This is always "sql#diskEncryptionStatus".
	Kind string `json:"kind"`
}

// IPConfiguration contains IP management configuration.
type IPConfiguration struct {
	// IPv4Enabled specifies whether the instance should be assigned an IP address.
	IPv4Enabled bool `json:"ipv4Enabled"`
	// PrivateNetwork is the resource link for the VPC network.
	PrivateNetwork string `json:"privateNetwork,omitempty"`
	// RequireSsl specifies whether SSL connections over IP should be enforced.
	RequireSsl bool `json:"requireSsl,omitempty"`
	// AuthorizedNetworks contains the list of external networks.
	AuthorizedNetworks []*ACLEntry `json:"authorizedNetworks,omitempty"`
	// AllocatedIpRange is the name of the allocated IP range.
	AllocatedIpRange string `json:"allocatedIpRange,omitempty"`
	// EnablePrivatePathForGoogleCloudServices specifies whether private path is enabled.
	EnablePrivatePathForGoogleCloudServices bool `json:"enablePrivatePathForGoogleCloudServices,omitempty"`
	// SslMode specifies the SSL/TLS mode for connections.
	SslMode string `json:"sslMode,omitempty"`
}

// ACLEntry contains access control entry.
type ACLEntry struct {
	// Value is the allowlisted value for the access control list.
	Value string `json:"value"`
	// ExpirationTime is when this access control entry expires in RFC 3339 format.
	ExpirationTime time.Time `json:"expirationTime,omitempty"`
	// Name is a label to identify this entry.
	Name string `json:"name,omitempty"`
	// Kind is the kind of resource. This is always "sql#aclEntry".
	Kind string `json:"kind"`
}

// LocationPreference contains preferred location settings.
type LocationPreference struct {
	// FollowGaeApplication is deprecated.
	FollowGaeApplication string `json:"followGaeApplication,omitempty"`
	// Zone is the preferred Compute Engine zone.
	Zone string `json:"zone,omitempty"`
	// SecondaryZone is the preferred Compute Engine zone for the failover.
	SecondaryZone string `json:"secondaryZone,omitempty"`
	// Kind is the kind of resource. This is always "sql#locationPreference".
	Kind string `json:"kind"`
}

// DatabaseFlags contains database flags.
type DatabaseFlags struct {
	// Name is the name of the flag.
	Name string `json:"name"`
	// Value is the value of the flag.
	Value string `json:"value"`
}

// MaintenanceWindow contains maintenance window settings.
type MaintenanceWindow struct {
	// Hour is the hour of day (0-23) for maintenance window.
	Hour int32 `json:"hour,omitempty"`
	// Day is the day of week (1-7) for maintenance window.
	Day int32 `json:"day,omitempty"`
	// UpdateTrack is the maintenance timing setting.
	UpdateTrack string `json:"updateTrack,omitempty"`
	// Kind is the kind of resource. This is always "sql#maintenanceWindow".
	Kind string `json:"kind"`
}

// BackupConfiguration contains database instance backup configuration.
type BackupConfiguration struct {
	// StartTime is the start time for the daily backup configuration in 24-hour format.
	StartTime string `json:"startTime,omitempty"`
	// Enabled specifies whether this configuration is enabled.
	Enabled bool `json:"enabled"`
	// Kind is the kind of resource. This is always "sql#backupConfiguration".
	Kind string `json:"kind"`
	// BinaryLogEnabled specifies whether binary log is enabled.
	BinaryLogEnabled bool `json:"binaryLogEnabled,omitempty"`
	// ReplicationLogArchivingEnabled specifies whether replication log archiving is enabled.
	ReplicationLogArchivingEnabled bool `json:"replicationLogArchivingEnabled,omitempty"`
	// Location is the location of the backup.
	Location string `json:"location,omitempty"`
	// PointInTimeRecoveryEnabled specifies whether point in time recovery is enabled.
	PointInTimeRecoveryEnabled bool `json:"pointInTimeRecoveryEnabled,omitempty"`
	// TransactionLogRetentionDays is the number of days of transaction logs retained.
	TransactionLogRetentionDays int32 `json:"transactionLogRetentionDays,omitempty"`
	// BackupRetentionSettings contains backup retention settings.
	BackupRetentionSettings *BackupRetentionSettings `json:"backupRetentionSettings,omitempty"`
}

// BackupRetentionSettings contains backup retention settings.
type BackupRetentionSettings struct {
	// RetainedBackups is the number of backups to retain.
	RetainedBackups int32 `json:"retainedBackups,omitempty"`
	// RetentionUnit is the unit that 'retainedBackups' represents.
	RetentionUnit string `json:"retentionUnit,omitempty"`
}

// SqlActiveDirectoryConfig contains Active Directory configuration.
type SqlActiveDirectoryConfig struct {
	// Kind is the kind of resource. This is always "sql#activeDirectoryConfig".
	Kind string `json:"kind"`
	// Domain is the name of the domain.
	Domain string `json:"domain"`
}

// DenyMaintenancePeriod contains deny maintenance period.
type DenyMaintenancePeriod struct {
	// StartDate is the start date.
	StartDate string `json:"startDate"`
	// EndDate is the end date.
	EndDate string `json:"endDate"`
	// Time is the time in UTC when the deny maintenance period starts.
	Time string `json:"time"`
}

// InsightsConfig contains Query Insights configuration.
type InsightsConfig struct {
	// QueryInsightsEnabled specifies whether Query Insights is enabled.
	QueryInsightsEnabled bool `json:"queryInsightsEnabled"`
	// RecordClientAddress specifies whether the client IP address should be recorded.
	RecordClientAddress bool `json:"recordClientAddress,omitempty"`
	// RecordApplicationTags specifies whether application tags should be recorded.
	RecordApplicationTags bool `json:"recordApplicationTags,omitempty"`
	// QueryStringLength is the maximum query length stored in bytes.
	QueryStringLength int32 `json:"queryStringLength,omitempty"`
	// QueryPlansPerMinute is the number of query plans generated by Insights per minute.
	QueryPlansPerMinute int32 `json:"queryPlansPerMinute,omitempty"`
}

// PasswordValidationPolicy contains password validation policy.
type PasswordValidationPolicy struct {
	// MinLength is the minimum length of the password.
	MinLength int32 `json:"minLength,omitempty"`
	// Complexity specifies the complexity of the password.
	Complexity string `json:"complexity,omitempty"`
	// ReuseInterval is the number of previous passwords that cannot be reused.
	ReuseInterval int32 `json:"reuseInterval,omitempty"`
	// DisallowUsernameSubstring specifies if username is not allowed in the password.
	DisallowUsernameSubstring bool `json:"disallowUsernameSubstring,omitempty"`
	// PasswordChangeInterval is the minimum interval after which password can be changed.
	PasswordChangeInterval string `json:"passwordChangeInterval,omitempty"`
	// EnablePasswordPolicy specifies whether the password policy is enabled.
	EnablePasswordPolicy bool `json:"enablePasswordPolicy,omitempty"`
}

// SqlServerAuditConfig contains SQL Server specific audit configuration.
type SqlServerAuditConfig struct {
	// Kind is the kind of resource. This is always "sql#sqlServerAuditConfig".
	Kind string `json:"kind"`
	// Bucket is the Cloud Storage bucket name where the audit files are stored.
	Bucket string `json:"bucket,omitempty"`
	// RetentionInterval specifies how long to keep generated audit files.
	RetentionInterval string `json:"retentionInterval,omitempty"`
	// UploadInterval specifies how often to upload generated audit files.
	UploadInterval string `json:"uploadInterval,omitempty"`
}

// InstancesListResponse represents a response from listing instances.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/instances/list
type InstancesListResponse struct {
	// Kind is the kind of resource. This is always "sql#instancesList".
	Kind string `json:"kind"`
	// Items contains the list of Cloud SQL instances.
	Items []*DatabaseInstance `json:"items"`
	// NextPageToken is used to continue a previous list request.
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// Database represents a Cloud SQL database resource.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/databases
type Database struct {
	// Kind is the kind of resource. This is always "sql#database".
	Kind string `json:"kind"`
	// Charset is the Cloud SQL charset value.
	Charset string `json:"charset"`
	// Collation is the Cloud SQL collation value.
	Collation string `json:"collation"`
	// Etag is the HTTP 1.1 Entity tag for the resource.
	Etag string `json:"etag,omitempty"`
	// Name is the name of the database in the Cloud SQL instance.
	Name string `json:"name"`
	// Instance is the name of the Cloud SQL instance.
	Instance string `json:"instance"`
	// SelfLink is the URI of this resource.
	SelfLink string `json:"selfLink"`
	// Project is the project ID of the project containing the Cloud SQL database.
	Project string `json:"project"`
}

// DatabasesListResponse represents a response from listing databases.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/databases/list
type DatabasesListResponse struct {
	// Kind is the kind of resource. This is always "sql#databasesList".
	Kind string `json:"kind"`
	// Items contains the list of databases.
	Items []*Database `json:"items"`
}

// User represents a Cloud SQL user resource.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/users
type User struct {
	// Kind is the kind of resource. This is always "sql#user".
	Kind string `json:"kind"`
	// Password is the user's password.
	Password string `json:"password,omitempty"`
	// Etag is the HTTP 1.1 Entity tag for the resource.
	Etag string `json:"etag,omitempty"`
	// Name is the name of the user in the Cloud SQL instance.
	Name string `json:"name"`
	// Host is the host name from which the user can connect.
	Host string `json:"host,omitempty"`
	// Instance is the name of the Cloud SQL instance.
	Instance string `json:"instance"`
	// Project is the project ID of the project containing the Cloud SQL database.
	Project string `json:"project"`
	// Type is the user type.
	Type string `json:"type,omitempty"`
	// PasswordPolicy contains the user's password policy.
	PasswordPolicy *UserPasswordValidationPolicy `json:"passwordPolicy,omitempty"`
}

// UserPasswordValidationPolicy contains user-specific password validation policy.
type UserPasswordValidationPolicy struct {
	// AllowedFailedAttempts is the number of failed attempts allowed before the user is locked.
	AllowedFailedAttempts int32 `json:"allowedFailedAttempts,omitempty"`
	// PasswordExpirationDuration is the expiration duration of the current password.
	PasswordExpirationDuration string `json:"passwordExpirationDuration,omitempty"`
	// EnableFailedAttemptsCheck specifies if failed login attempts check is enabled.
	EnableFailedAttemptsCheck bool `json:"enableFailedAttemptsCheck,omitempty"`
	// EnablePasswordVerification specifies if the user must provide current password before updating.
	EnablePasswordVerification bool `json:"enablePasswordVerification,omitempty"`
	// Status contains password status.
	Status *PasswordStatus `json:"status,omitempty"`
}

// PasswordStatus contains password status.
type PasswordStatus struct {
	// Locked specifies if the user is locked because of too many failed attempts.
	Locked bool `json:"locked,omitempty"`
	// PasswordExpirationTime is when the password expires in RFC 3339 format.
	PasswordExpirationTime time.Time `json:"passwordExpirationTime,omitempty"`
}

// UsersListResponse represents a response from listing users.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/users/list
type UsersListResponse struct {
	// Kind is the kind of resource. This is always "sql#usersList".
	Kind string `json:"kind"`
	// Items contains the list of users.
	Items []*User `json:"items"`
	// NextPageToken is used to continue a previous list request.
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// Operation represents a Cloud SQL operation resource.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/operations
type Operation struct {
	// Kind is the kind of resource. This is always "sql#operation".
	Kind string `json:"kind"`
	// TargetLink is the URI of the target resource.
	TargetLink string `json:"targetLink,omitempty"`
	// Status is the status of the operation.
	Status string `json:"status"`
	// User is the email address of the user who initiated the operation.
	User string `json:"user,omitempty"`
	// InsertTime is the time the operation was created in RFC 3339 format.
	InsertTime time.Time `json:"insertTime"`
	// StartTime is the time the operation started in RFC 3339 format.
	StartTime time.Time `json:"startTime,omitempty"`
	// EndTime is the time the operation ended in RFC 3339 format.
	EndTime time.Time `json:"endTime,omitempty"`
	// Error contains the error information if the operation failed.
	Error *OperationErrors `json:"error,omitempty"`
	// OperationType is the type of the operation.
	OperationType string `json:"operationType"`
	// Name is the unique identifier of the operation.
	Name string `json:"name"`
	// TargetId is the database instance identifier.
	TargetId string `json:"targetId,omitempty"`
	// SelfLink is the URI of this resource.
	SelfLink string `json:"selfLink"`
	// TargetProject is the project ID of the target instance.
	TargetProject string `json:"targetProject,omitempty"`
}

// OperationErrors contains operation error information.
type OperationErrors struct {
	// Kind is the kind of resource. This is always "sql#operationErrors".
	Kind string `json:"kind"`
	// Errors contains the list of errors.
	Errors []*OperationError `json:"errors,omitempty"`
}

// OperationError contains a single operation error.
type OperationError struct {
	// Kind is the kind of resource. This is always "sql#operationError".
	Kind string `json:"kind"`
	// Code is the error code.
	Code string `json:"code"`
	// Message is the error message.
	Message string `json:"message,omitempty"`
}

// OperationsListResponse represents a response from listing operations.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/operations/list
type OperationsListResponse struct {
	// Kind is the kind of resource. This is always "sql#operationsList".
	Kind string `json:"kind"`
	// Items contains the list of operations.
	Items []*Operation `json:"items"`
	// NextPageToken is used to continue a previous list request.
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// InstanceInsertRequest represents the request body for creating an instance.
type InstanceInsertRequest struct {
	// Name is the name of the instance.
	Name string `json:"name"`
	// DatabaseVersion is the database engine type and version.
	DatabaseVersion string `json:"databaseVersion"`
	// Region is the geographical region.
	Region string `json:"region,omitempty"`
	// Settings contains the settings for the instance.
	Settings *Settings `json:"settings,omitempty"`
	// MasterInstanceName is the name of the instance acting as master.
	MasterInstanceName string `json:"masterInstanceName,omitempty"`
	// RootPassword is the initial root password.
	RootPassword string `json:"rootPassword,omitempty"`
}

// InstancePatchRequest represents the request body for patching an instance.
type InstancePatchRequest struct {
	// Settings contains the settings for the instance.
	Settings *Settings `json:"settings,omitempty"`
}

// DatabaseInsertRequest represents the request body for creating a database.
type DatabaseInsertRequest struct {
	// Name is the name of the database.
	Name string `json:"name"`
	// Charset is the charset value.
	Charset string `json:"charset,omitempty"`
	// Collation is the collation value.
	Collation string `json:"collation,omitempty"`
}

// DatabasePatchRequest represents the request body for patching a database.
type DatabasePatchRequest struct {
	// Charset is the charset value.
	Charset string `json:"charset,omitempty"`
	// Collation is the collation value.
	Collation string `json:"collation,omitempty"`
}

// UserInsertRequest represents the request body for creating a user.
type UserInsertRequest struct {
	// Name is the name of the user.
	Name string `json:"name"`
	// Password is the user's password.
	Password string `json:"password,omitempty"`
	// Host is the host name from which the user can connect.
	Host string `json:"host,omitempty"`
	// Type is the user type.
	Type string `json:"type,omitempty"`
}

// UserUpdateRequest represents the request body for updating a user.
type UserUpdateRequest struct {
	// Password is the user's password.
	Password string `json:"password,omitempty"`
	// Host is the host name from which the user can connect.
	Host string `json:"host,omitempty"`
}

// APIError represents an error response from the Cloud SQL Admin API.
// Reference: https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1/ErrorResponse
type APIError struct {
	Error ErrorDetails `json:"error"`
}

// ErrorDetails contains the details of an API error.
type ErrorDetails struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Status  string        `json:"status,omitempty"`
	Errors  []ErrorReason `json:"errors,omitempty"`
}

// ErrorReason contains the reason for an error.
type ErrorReason struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
