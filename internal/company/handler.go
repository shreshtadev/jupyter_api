package company

import (
	"encoding/json"
	"net/http"
	"time"

	"shreshtasmg.in/jupyter/internal/feature"
	"shreshtasmg.in/jupyter/internal/utils"
)

// Incoming request DTO
type CreateCompanyRequest struct {
	CompanyName     string   `json:"company_name"`
	StartDate       *string  `json:"start_date,omitempty"` // "YYYY-MM-DD"
	EndDate         *string  `json:"end_date,omitempty"`
	TotalUsageQuota *int64   `json:"total_usage_quota,omitempty"`
	AwsBucketName   *string  `json:"aws_bucket_name,omitempty"`
	AwsBucketRegion *string  `json:"aws_bucket_region,omitempty"`
	AwsAccessKey    *string  `json:"aws_access_key,omitempty"`
	AwsSecretKey    *string  `json:"aws_secret_key,omitempty"`
	Features        []string `json:"features"`
}

// Response DTO
type CreateCompanyResponse struct {
	CompanyAPIKey string   `json:"company_api_key"`
	CompanySlug   string   `json:"company_slug"`
	Features      []string `json:"features"`
	CreatedAt     string   `json:"created_at"`
}

type Handler struct {
	repo        Repository
	featureRepo feature.Repository
}

func NewHandler(repo Repository, featureRepo feature.Repository) *Handler {
	return &Handler{repo: repo, featureRepo: featureRepo}
}

// CreateCompany godoc
// @Summary      Create a company
// @Description  Creates a company with automatically generated API key & slug
// @Tags         companies
// @Accept       json
// @Produce      json
// @Param        body  body      CreateCompanyRequest  true  "Company payload"
// @Success      201   {object}  CreateCompanyResponse
// @Failure      400   {string}  string "invalid request"
// @Failure      500   {string}  string "internal error"
// @Router       /companies [post]
func (h *Handler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var req CreateCompanyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Required field
	if req.CompanyName == "" {
		http.Error(w, "company_name is required", http.StatusBadRequest)
		return
	}

	// Generate ID & API key
	id := utils.GenerateID()
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		http.Error(w, "failed to generate api key", http.StatusInternalServerError)
		return
	}

	// Generate slug
	slug := utils.Slugify(req.CompanyName)

	// Parse optional dates
	var startDatePtr *time.Time
	var endDatePtr *time.Time

	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err == nil {
			startDatePtr = &t
		}
	}

	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err == nil {
			endDatePtr = &t
		}
	}

	company := &Company{
		ID:              id,
		CompanyName:     req.CompanyName,
		CompanySlug:     slug,
		CompanyAPIKey:   apiKey,
		StartDate:       startDatePtr,
		EndDate:         endDatePtr,
		TotalUsageQuota: req.TotalUsageQuota,
		AwsBucketName:   req.AwsBucketName,
		AwsBucketRegion: req.AwsBucketRegion,
		AwsAccessKey:    req.AwsAccessKey,
		AwsSecretKey:    req.AwsSecretKey,
	}

	// Store in DB
	if err := h.repo.Create(company); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if err := h.featureRepo.SetCompanyFeatures(company.ID, req.Features); err != nil {
		http.Error(w, "failed to set features", http.StatusInternalServerError)
		return
	}

	resp := CreateCompanyResponse{
		CompanyAPIKey: apiKey,
		CompanySlug:   slug,
		CreatedAt:     utils.GetShortDate(company.CreatedAt),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
