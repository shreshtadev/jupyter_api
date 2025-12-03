// internal/user/repository.go
package user

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	GetByEmail(email string) (*User, error)
	Create(u *User) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByEmail(email string) (*User, error) {
	var u User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *repository) Create(u *User) error {
	return r.db.Create(u).Error
}
