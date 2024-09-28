package usecase

import (
	"github.com/iagonc/jorge-cli/internal/schemas"
)

// ResourceUseCase interface defines methods for resource use cases
// TODO: redefine each use case to use the name down below instead of execute
type ResourceUseCase interface {
    CreateResource(resource *schemas.Resource) error
    DeleteResource(resource *schemas.Resource, id uint) error
    GetResourceByID(resource *schemas.Resource, id uint) error
    ListResources() ([]*schemas.Resource, error)
    ListResourcesByName(name string) ([]*schemas.Resource, error)
    UpdateResource(resource *schemas.Resource) error
}
