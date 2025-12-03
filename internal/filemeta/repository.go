package filemeta

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(f *FileMeta) error
	GetByID(id string) (*FileMeta, error)
	ListByCompanyID(companyID string, limit, offset int) ([]FileMeta, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(f *FileMeta) error {
	return r.db.Create(f).Error
}

func (r *repository) GetByID(id string) (*FileMeta, error) {
	var meta FileMeta
	if err := r.db.Where("id = ?", id).First(&meta).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &meta, nil
}

func (r *repository) ListByCompanyID(companyID string, limit, offset int) ([]FileMeta, error) {
	var metas []FileMeta
	q := r.db.Where("company_id = ?", companyID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := q.Find(&metas).Error; err != nil {
		return nil, err
	}
	return metas, nil
}
