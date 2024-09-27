package usecase

import (
	"github.com/iagonc/jorge-cli/internal/schemas"
)

type ResourceUseCase interface {
    CreateResource(resource *schemas.Resource) error
    DeleteResource(id uint) error
    ListResources() ([]*schemas.Resource, error)
    ListResourcesByName(name string) ([]*schemas.Resource, error)
    UpdateResource(resource *schemas.Resource) error
}
