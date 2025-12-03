package company

import "time"

type Company struct {
	ID              string     `gorm:"type:varchar(40);primaryKey;column:id"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
	CompanyName     string     `gorm:"type:varchar(144);not null;column:company_name"`
	CompanySlug     string     `gorm:"type:varchar(255);not null;column:company_slug"`
	CompanyAPIKey   string     `gorm:"type:varchar(255);not null;column:company_api_key"`
	StartDate       *time.Time `gorm:"column:start_date"`
	EndDate         *time.Time `gorm:"column:end_date"`
	TotalUsageQuota *int64     `gorm:"column:total_usage_quota"`
	UsedQuota       int64      `gorm:"column:used_quota;default:0"`
	AwsBucketName   *string    `gorm:"type:varchar(64);column:aws_bucket_name"`
	AwsBucketRegion *string    `gorm:"type:varchar(50);column:aws_bucket_region"`
	AwsAccessKey    *string    `gorm:"type:varchar(25);column:aws_access_key"`
	AwsSecretKey    *string    `gorm:"type:varchar(45);column:aws_secret_key"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Company) TableName() string {
	return "companies"
}
