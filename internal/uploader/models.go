package uploader

import (
	"time"
)

type UploaderConfig struct {
	ID              string    `gorm:"type:varchar(40);primaryKey;column:id"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	AwsBucketName   string    `gorm:"type:varchar(64);not null;column:aws_bucket_name"`
	AwsBucketRegion string    `gorm:"type:varchar(50);not null;column:aws_bucket_region"`
	AwsAccessKey    string    `gorm:"type:varchar(25);not null;column:aws_access_key"`
	AwsSecretKey    string    `gorm:"type:varchar(45);not null;column:aws_secret_key"`
	TotalQuota      int64     `gorm:"column:total_quota;default:5368709120"`  // 5GB
	DefaultQuota    int64     `gorm:"column:default_quota;default:262144000"` // 250MB
	IsActive        int16     `gorm:"column:is_active;default:0"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UploaderConfig) TableName() string {
	return "uploader_config"
}
