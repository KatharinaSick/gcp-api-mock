package store

// Store defines the interface for resource storage
type Store interface {
	// Get retrieves a resource by ID
	Get(id string) (interface{}, bool)
	
	// Set stores a resource with the given ID
	Set(id string, value interface{})
	
	// Delete removes a resource by ID
	Delete(id string) bool
	
	// List returns all resources
	List() []interface{}
	
	// Exists checks if a resource exists
	Exists(id string) bool
}