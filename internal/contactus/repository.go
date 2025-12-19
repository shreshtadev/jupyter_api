package contactus

import (
	"gorm.io/gorm"
)

type Repository interface {
	Create(f *ContactUs) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(f *ContactUs) error {
	return r.db.Create(f).Error
}
