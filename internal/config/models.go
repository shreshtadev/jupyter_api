package config

import "time"

type AdminClient struct {
	ClientID     string     `gorm:"column:client_id;type:varchar(40);primary_key"`
	ClientSecret string     `gorm:"column:client_secret;type:varchar(100);primary_key"`
	IsActive     bool       `gorm:"column:is_active;default:true"`
	CreatedAt    *time.Time `gorm:"column:created_at;type:datetime(3)"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:datetime(3)"`
}

func (AdminClient) TableName() string {
	return "admin_client"
}
