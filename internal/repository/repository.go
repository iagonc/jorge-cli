package repository

import "github.com/iagonc/jorge-cli/internal/schemas"

// ResourceRepository defines the interface for resource-related database operations
type ResourceRepository interface {
    Create(resource *schemas.Resource) error                   // Creates a new resource
    Delete(id uint) error                                       // Deletes a resource by ID
    FindByName(name string) (*schemas.Resource, error)          // Finds a resource by its name
    FindByDNS(dns string) (*schemas.Resource, error)            // Finds a resource by its DNS
    FindByID(id uint) (*schemas.Resource, error)                // Finds a resource by its ID
    List() ([]*schemas.Resource, error)                         // Lists all resources
    Update(resource *schemas.Resource) error                    // Updates an existing resource
    ListByName(name string) ([]*schemas.Resource, error)        // Lists resources by name
}
