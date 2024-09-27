package repository

import "github.com/iagonc/jorge-cli/internal/schemas"

type ResourceRepository interface {
	Create(resource *schemas.Resource) error
	Delete(id uint) error
	FindByName(name string) (*schemas.Resource, error)
	FindByDNS(dns string) (*schemas.Resource, error)
	FindByID(id uint) (*schemas.Resource, error)
	List() ([]*schemas.Resource, error)           // Lista todos os recursos
	Update(resource *schemas.Resource) error      // Atualiza um recurso
	ListByName(name string) ([]*schemas.Resource, error)  // Lista recursos pelo nome
}
