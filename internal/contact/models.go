package contact

import (
	"time"
)

type Contact struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	FullName       string    `gorm:"type:varchar(75);not null;column:full_name"`
	EmailAddress   string    `gorm:"type:varchar(255);not null;column:email_address"`
	ProjectType    string    `gorm:"type:varchar(40);not null;column:project_type"`
	ProjectDetails string    `gorm:"type:text;not null;column:project_details"`
	CreatedAt      time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (Contact) TableName() string {
	return "contact_us"
}
