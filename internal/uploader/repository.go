package uploader

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(cfg *UploaderConfig) error
	FindActiveConfig() (*UploaderConfig, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(cfg *UploaderConfig) error {
	return r.db.Create(cfg).Error
}

func (r *repository) FindActiveConfig() (*UploaderConfig, error) {
	var activeUploaderConfig UploaderConfig
	if err := r.db.Where("is_active = ?", true).First(&activeUploaderConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &activeUploaderConfig, nil
}
