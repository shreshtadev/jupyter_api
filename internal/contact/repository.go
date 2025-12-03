package contact

import "gorm.io/gorm"

type Repository interface {
	Create(c *Contact) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(c *Contact) error {
	return r.db.Create(c).Error
}
