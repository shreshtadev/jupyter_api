package config

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(adminClient *AdminClient) error
	FindBy(client_id, client_secret string) (*AdminClient, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(adminClient *AdminClient) error {
	if adminClient.ClientID == "" || adminClient.ClientSecret == "" {
		return errors.New("client_id and client_secret are required")
	}
	adminClient.IsActive = true
	return r.db.Create(adminClient).Error
}

func (r *repository) FindBy(client_id, client_secret string) (*AdminClient, error) {
	var ac AdminClient
	if err := r.db.Where("client_id = ? AND client_secret = ? AND is_active=?", client_id, client_secret, true).First(&ac).Error; err != nil {
		return nil, err
	}
	return &ac, nil
}
