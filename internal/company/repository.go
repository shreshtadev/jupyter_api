package company

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Create(c *Company) error
	GetByAPIKey(apiKey string) (*Company, error)
	GetByID(companyId string) (*Company, error)
	IncrementUsedQuota(companyID string, delta int64) error
	ResetUsedQuota(companyID string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(c *Company) error {
	return r.db.Create(c).Error
}

func (r *repository) GetByID(companyId string) (*Company, error) {
	var c Company
	if err := r.db.Where("company_id = ?", companyId).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *repository) GetByAPIKey(apiKey string) (*Company, error) {
	var c Company
	if err := r.db.Where("company_api_key = ?", apiKey).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *repository) IncrementUsedQuota(companyID string, delta int64) error {
	return r.db.Model(&Company{}).
		Where("id = ?", companyID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		UpdateColumn("used_quota", gorm.Expr("used_quota + ?", delta)).Error
}

func (r *repository) ResetUsedQuota(companyID string) error {
	return r.db.Model(&Company{}).
		Where("id = ?", companyID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		UpdateColumn("used_quota", 0).Error
}
