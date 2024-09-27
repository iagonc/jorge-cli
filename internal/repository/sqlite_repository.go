package repository

import (
	"errors"

	"github.com/iagonc/jorge-cli/internal/schemas"
	"gorm.io/gorm"
)
var ErrResourceNotFound = errors.New("resource not found")

type sqliteResourceRepository struct {
	db *gorm.DB
}

func NewSQLiteResourceRepository(db *gorm.DB) ResourceRepository {
	return &sqliteResourceRepository{db: db}
}

func (r *sqliteResourceRepository) Create(resource *schemas.Resource) error {
	return r.db.Create(resource).Error
}

func (r *sqliteResourceRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&schemas.Resource{}).Error
}

func (r *sqliteResourceRepository) FindByName(name string) (*schemas.Resource, error) {
	var resource schemas.Resource
	if err := r.db.Where("name = ?", name).First(&resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrResourceNotFound // Retorna um erro personalizado
        }
		return nil, err
	}
	return &resource, nil
}

func (r *sqliteResourceRepository) FindByDNS(dns string) (*schemas.Resource, error) {
	var resource schemas.Resource
	if err := r.db.Where("dns = ?", dns).First(&resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrResourceNotFound // Retorna um erro personalizado
        }
		return nil, err
	}
	return &resource, nil
}

func (r *sqliteResourceRepository) FindByID(id uint) (*schemas.Resource, error) {
    var resource schemas.Resource
    if err := r.db.First(&resource, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrResourceNotFound // Retorna um erro personalizado
        }
        return nil, err
    }
    return &resource, nil
}

func (r *sqliteResourceRepository) ListByName(name string) ([]*schemas.Resource, error) {
	var resources []*schemas.Resource
	if err := r.db.Where("name = ?", name).Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *sqliteResourceRepository) List() ([]*schemas.Resource, error) {
	var resources []*schemas.Resource
	if err := r.db.Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *sqliteResourceRepository) Update(resource *schemas.Resource) error {
	return r.db.Save(resource).Error
}