// internal/feature/repository.go
package feature

import "gorm.io/gorm"

type Repository interface {
	IsFeatureEnabled(companyID, featureKey string) (bool, error)
	SetCompanyFeatures(companyID string, featureKeys []string) error
}

type repo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository {
	return &repo{db: db}
}

func (r *repo) IsFeatureEnabled(companyID, featureKey string) (bool, error) {
	var cf CompanyFeature
	if err := r.db.
		Where("company_id = ? AND feature_key = ?", companyID, featureKey).
		First(&cf).Error; err != nil {
		return false, err
	}
	return cf.IsEnabled, nil
}

// SetCompanyFeatures sets the enabled features for a company to exactly the given set.
//
// Semantics:
//   - For each featureKey in featureKeys:
//   - If a row exists in company_features -> set is_enabled = true
//   - If not -> create row with is_enabled = true
//   - For existing company_features rows whose feature_key is NOT in featureKeys:
//   - set is_enabled = false
//
// This runs in a transaction.
func (r *repo) SetCompanyFeatures(companyID string, featureKeys []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Load existing feature mappings for this company
		var existing []CompanyFeature
		if err := tx.
			Where("company_id = ?", companyID).
			Find(&existing).Error; err != nil {
			return err
		}

		existingByKey := make(map[string]*CompanyFeature, len(existing))
		for i := range existing {
			cf := &existing[i]
			existingByKey[cf.FeatureKey] = cf
		}

		// Build a set of desired feature keys
		desired := make(map[string]struct{}, len(featureKeys))
		for _, key := range featureKeys {
			if key == "" {
				continue
			}
			desired[key] = struct{}{}
		}

		// Enable or create each desired feature
		for key := range desired {
			if cf, ok := existingByKey[key]; ok {
				// Update existing row to enabled
				if !cf.IsEnabled {
					if err := tx.Model(cf).
						Update("is_enabled", true).Error; err != nil {
						return err
					}
				}
			} else {
				// Create a new row
				newCF := CompanyFeature{
					CompanyID:  companyID,
					FeatureKey: key,
					IsEnabled:  true,
				}
				if err := tx.Create(&newCF).Error; err != nil {
					return err
				}
			}
		}

		// Disable any features that are no longer desired
		for key, cf := range existingByKey {
			if _, keep := desired[key]; !keep && cf.IsEnabled {
				if err := tx.Model(cf).
					Update("is_enabled", false).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
