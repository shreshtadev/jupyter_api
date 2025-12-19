package contactus

import (
	"time"
)

type ContactUs struct {
	ID              int64     `gorm:"column:id;type:bigint(20);primary_key;AUTO_INCREMENT"`
	FullName        string    `gorm:"column:full_name;type:varchar(75);NOT NULL"`
	ContactEmail    string    `gorm:"column:contact_email;type:varchar(255);NOT NULL"`
	ContactNumber   string    `gorm:"column:contact_number;type:varchar(25);NOT NULL"`
	ProjectType     string    `gorm:"column:project_type;type:varchar(40)"`
	ProjectDetails  string    `gorm:"column:project_details;type:text"`
	ContactUsStatus string    `gorm:"column:contact_us_status;type:varchar(40);default:submit;NOT NULL"`
	Remarks         string    `gorm:"column:remarks;type:text"`
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime(3)"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime(3)"`
}

func (ContactUs) TableName() string {
	return "contactus"
}
