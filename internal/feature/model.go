package feature

type Feature struct {
	Key         string `gorm:"type:varchar(64);primaryKey;column:feature_key"`
	Name        string `gorm:"type:varchar(128);not null;column:name"`
	Description string `gorm:"type:text;column:description"`
	IsActive    bool   `gorm:"column:is_active;default:true"`
}

func (Feature) TableName() string { return "features" }

type CompanyFeature struct {
	CompanyID  string `gorm:"type:varchar(40);primaryKey;column:company_id"`
	FeatureKey string `gorm:"type:varchar(64);primaryKey;column:feature_key"`
	IsEnabled  bool   `gorm:"column:is_enabled;default:false"`
	ConfigJSON string `gorm:"type:text;column:config_json"`
}

func (CompanyFeature) TableName() string { return "company_features" }
