// internal/user/model.go
package user

import "time"

type User struct {
	ID           string    `gorm:"type:varchar(40);primaryKey;column:id"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null;column:email"`
	PasswordHash string    `gorm:"type:varchar(255);not null;column:password_hash"`
	CompanyID    *string   `gorm:"type:varchar(40);column:company_id"`
	Role         string    `gorm:"type:varchar(32);not null;column:role"` // "admin", "user", etc.
}

func (User) TableName() string {
	return "users"
}
