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

// NewSQLiteResourceRepository creates a new repository for SQLite
func NewSQLiteResourceRepository(db *gorm.DB) ResourceRepository {
    return &sqliteResourceRepository{db: db}
}

// Create inserts a new resource into the database
func (r *sqliteResourceRepository) Create(resource *schemas.Resource) error {
    return r.db.Create(resource).Error
}

// Delete removes a resource from the database by ID
func (r *sqliteResourceRepository) Delete(id uint) error {
    return r.db.Where("id = ?", id).Delete(&schemas.Resource{}).Error
}

// FindByName searches for a resource by its name and returns it, or an error if not found
func (r *sqliteResourceRepository) FindByName(name string) (*schemas.Resource, error) {
    var resource schemas.Resource
    if err := r.db.Where("name = ?", name).First(&resource).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrResourceNotFound // Custom error for not found records
        }
        return nil, err
    }
    return &resource, nil
}

// FindByDNS searches for a resource by its DNS and returns it, or an error if not found
func (r *sqliteResourceRepository) FindByDNS(dns string) (*schemas.Resource, error) {
    var resource schemas.Resource
    if err := r.db.Where("dns = ?", dns).First(&resource).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrResourceNotFound // Custom error for not found records
        }
        return nil, err
    }
    return &resource, nil
}

// FindByID searches for a resource by its ID and returns it, or an error if not found
func (r *sqliteResourceRepository) FindByID(id uint) (*schemas.Resource, error) {
    var resource schemas.Resource
    if err := r.db.First(&resource, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrResourceNotFound // Custom error for not found records
        }
        return nil, err
    }
    return &resource, nil
}

// ListByName retrieves all resources that match the given name
func (r *sqliteResourceRepository) ListByName(name string) ([]*schemas.Resource, error) {
    var resources []*schemas.Resource
    if err := r.db.Where("name = ?", name).Find(&resources).Error; err != nil {
        return nil, err
    }
    return resources, nil
}

// List retrieves all resources from the database
func (r *sqliteResourceRepository) List() ([]*schemas.Resource, error) {
    var resources []*schemas.Resource
    if err := r.db.Find(&resources).Error; err != nil {
        return nil, err
    }
    return resources, nil
}

// Update updates an existing resource in the database
func (r *sqliteResourceRepository) Update(resource *schemas.Resource) error {
    return r.db.Save(resource).Error
}
