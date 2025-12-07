package uploader

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"shreshtasmg.in/jupyter/internal/company"
	"shreshtasmg.in/jupyter/internal/filemeta"
	"shreshtasmg.in/jupyter/internal/utils"
)

type CreateUploaderConfigRequest struct {
	AwsBucketName   string `json:"aws_bucket_name"`
	AwsBucketRegion string `json:"aws_bucket_region"`
	AwsAccessKey    string `json:"aws_access_key"`
	AwsSecretKey    string `json:"aws_secret_key"`
	TotalQuota      *int64 `json:"total_quota,omitempty"`
	DefaultQuota    *int64 `json:"default_quota,omitempty"`
	IsActive        *int16 `json:"is_active,omitempty"`
}

type CreateUploaderConfigResponse struct {
	CreatedAt string `json:"created_at"`
}

type GenerateUploadURLRequest struct {
	LocTag      string  `json:"loc_tag"`
	FileName    string  `json:"file_name"`               // required
	FileSize    int64   `json:"file_size"`               // required
	FileTxnType int16   `json:"file_txn_type"`           // required (e.g. 1=upload)
	FileTxnMeta *string `json:"file_txn_meta,omitempty"` // optional
}

// GenerateUploadURLResponse is returned to the client.
type GenerateUploadURLResponse struct {
	FileID    string `json:"file_id"`
	FileKey   string `json:"file_key"`
	UploadURL string `json:"upload_url"`
}

type CompanyFileMetaItem struct {
	ID          string  `json:"id"`
	CreatedAt   string  `json:"created_at"`
	FileName    *string `json:"file_name,omitempty"`
	FileSize    int64   `json:"file_size"`
	FileKey     string  `json:"file_key"`
	FileTxnType int16   `json:"file_txn_type"`
	FileTxnMeta *string `json:"file_txn_meta,omitempty"`
}

type ListCompanyFilesResponse struct {
	Items           []CompanyFileMetaItem `json:"items"`
	TotalUsageQuota *int64                `json:"total_usage_quota,omitempty"`
	UsedQuota       int64                 `json:"used_quota"`
	RemainingQuota  *int64                `json:"remaining_quota,omitempty"`
}

// Delete a single file by key
type DeleteFileRequest struct {
	FileKey     string  `json:"file_key"`
	FileTxnMeta *string `json:"file_txn_meta,omitempty"`
}

type DeleteFileResponse struct {
	FileKey string `json:"file_key"`
}

// Delete all files under a folder (prefix)
type DeleteFolderRequest struct {
	FolderPrefix string  `json:"folder_prefix"`
	FileTxnMeta  *string `json:"file_txn_meta,omitempty"`
}

type DeleteFolderResponse struct {
	FolderPrefix string `json:"folder_prefix"`
	DeletedCount int    `json:"deleted_count"`
	DeletedBytes int64  `json:"deleted_bytes"`
}

type ListFoldersResponse struct {
	Items     []string `json:"items"`
	NextToken *string  `json:"next_token,omitempty"`
}

type ListFilesResponse struct {
	Items     []string `json:"items"`
	NextToken *string  `json:"next_token,omitempty"`
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

type Handler struct {
	repo         Repository
	companyRepo  company.Repository
	s3Service    S3Service
	fileMetaRepo filemeta.Repository
}

func NewHandler(repo Repository, companyRepo company.Repository, s3Service S3Service, fileMetaRepo filemeta.Repository) *Handler {
	return &Handler{repo: repo, companyRepo: companyRepo, s3Service: s3Service, fileMetaRepo: fileMetaRepo}
}

// CreateUploaderConfig godoc
// @Summary      Create uploader configuration
// @Description  Creates a new uploader_config entry
// @Tags         uploader
// @Accept       json
// @Produce      json
// @Param        body  body      CreateUploaderConfigRequest  true  "Uploader config payload"
// @Success      201   {object}  CreateUploaderConfigResponse
// @Failure      400   {string}  string "invalid request"
// @Failure      500   {string}  string "internal error"
// @Router       /uploader/config [post]
func (h *Handler) CreateUploaderConfig(w http.ResponseWriter, r *http.Request) {
	var req CreateUploaderConfigRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.AwsBucketName == "" || req.AwsBucketRegion == "" || req.AwsAccessKey == "" || req.AwsSecretKey == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// Create ID if not provided
	id := utils.GenerateID()

	cfg := &UploaderConfig{
		ID:              id,
		AwsBucketName:   req.AwsBucketName,
		AwsBucketRegion: req.AwsBucketRegion,
		AwsAccessKey:    req.AwsAccessKey,
		AwsSecretKey:    req.AwsSecretKey,
	}

	if req.TotalQuota != nil {
		cfg.TotalQuota = *req.TotalQuota
	}

	if req.DefaultQuota != nil {
		cfg.DefaultQuota = *req.DefaultQuota
	}

	if req.IsActive != nil {
		cfg.IsActive = *req.IsActive
	}

	if err := h.repo.Create(cfg); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	resp := CreateUploaderConfigResponse{CreatedAt: utils.GetShortDate(cfg.CreatedAt)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// GenerateUploadURL godoc
// @Summary      Generate S3 presigned upload URL and create file meta
// @Description  Validates API key, generates a presigned S3 upload URL using company AWS config, and stores files_meta
// @Tags         uploader
// @Accept       json
// @Produce      json
// @Param        X-API-Key  header    string                    true  "Company API key"
// @Param        body       body      GenerateUploadURLRequest  true  "File upload request"
// @Success      201        {object}  GenerateUploadURLResponse
// @Failure      400        {string}  string "invalid request"
// @Failure      401        {string}  string "unauthorized"
// @Failure      500        {string}  string "internal error"
// @Router       /uploader/files [post]
func (h *Handler) GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "missing X-API-Key header", http.StatusUnauthorized)
		return
	}

	companyRec, err := h.companyRepo.GetByAPIKey(apiKey)
	if err != nil {
		http.Error(w, "failed to look up company", http.StatusInternalServerError)
		return
	}
	if companyRec == nil {
		http.Error(w, "invalid API key", http.StatusUnauthorized)
		return
	}

	var req GenerateUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.LocTag == "" {
		http.Error(w, "loc_tag is required", http.StatusBadRequest)
		return
	}

	if req.FileName == "" {
		http.Error(w, "file_name is required", http.StatusBadRequest)
		return
	}
	if req.FileSize <= 0 {
		http.Error(w, "file_size must be > 0", http.StatusBadRequest)
		return
	}
	if req.FileTxnType == 0 {
		http.Error(w, "file_txn_type is required", http.StatusBadRequest)
		return
	}

	// ----- QUOTA CHECK -----
	if companyRec.TotalUsageQuota != nil {
		total := *companyRec.TotalUsageQuota
		used := companyRec.UsedQuota
		r := total - used

		if req.FileSize > r {
			// Not enough remaining quota
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":           "quota_exceeded",
				"total_quota":     total,
				"used_quota":      used,
				"remaining_quota": r,
				"requested_size":  req.FileSize,
			})
			return
		}
	}

	safeName := sanitizeFileName(req.FileName)
	fileID := utils.GenerateID()
	fileKey := fmt.Sprintf("%s/%s/%s", companyRec.CompanySlug, req.LocTag, safeName)

	// Generate presigned URL
	uploadURL, err := h.s3Service.GeneratePresignedUploadURL(ctx, companyRec, fileKey, req.FileSize)
	if err != nil {
		http.Error(w, "failed to generate presigned URL", http.StatusInternalServerError)
		return
	}

	// Create files_meta row
	// We use pointers for nullable fields
	fileName := req.FileName
	meta := &filemeta.FileMeta{
		ID:          fileID,
		FileName:    &fileName,
		FileSize:    req.FileSize,
		FileKey:     fileKey,
		FileTxnType: req.FileTxnType,
		FileTxnMeta: req.FileTxnMeta,
		CompanyID:   &companyRec.ID,
	}

	if err := h.fileMetaRepo.Create(meta); err != nil {
		http.Error(w, "failed to create file meta", http.StatusInternalServerError)
		return
	}

	if err := h.companyRepo.IncrementUsedQuota(companyRec.ID, req.FileSize); err != nil {
		// We already created meta, but quota update failed -> log and continue
		// In a stricter system you'd wrap this whole operation in a DB transaction.
		http.Error(w, "failed to update quota", http.StatusInternalServerError)
		return
	}

	resp := GenerateUploadURLResponse{
		FileID:    fileID,
		FileKey:   fileKey,
		UploadURL: uploadURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// ListCompanyFiles godoc
// @Summary      List files for the calling company
// @Description  Uses X-API-Key to identify company and returns its files_meta records
// @Tags         uploader
// @Produce      json
// @Param        X-API-Key  header  string  true   "Company API key"
// @Param        limit      query   int     false  "Max number of items (default 50)"
// @Param        offset     query   int     false  "Offset for pagination (default 0)"
// @Success      200        {object}  ListCompanyFilesResponse
// @Failure      401        {string}  string "unauthorized"
// @Failure      500        {string}  string "internal error"
// @Router       /uploader/files [get]
func (h *Handler) ListCompanyFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "missing X-API-Key header", http.StatusUnauthorized)
		return
	}

	companyRec, err := h.companyRepo.GetByAPIKey(apiKey)
	if err != nil {
		http.Error(w, "failed to look up company", http.StatusInternalServerError)
		return
	}
	if companyRec == nil {
		http.Error(w, "invalid API key", http.StatusUnauthorized)
		return
	}

	// Parse pagination params
	q := r.URL.Query()
	limit := 50
	offset := 0

	if v := q.Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}
	if v := q.Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	metas, err := h.fileMetaRepo.ListByCompanyID(companyRec.ID, limit, offset)
	if err != nil {
		http.Error(w, "failed to list file meta", http.StatusInternalServerError)
		return
	}

	items := make([]CompanyFileMetaItem, 0, len(metas))
	for _, m := range metas {
		item := CompanyFileMetaItem{
			ID:          m.ID,
			CreatedAt:   m.CreatedAt.Format(time.RFC3339Nano),
			FileName:    m.FileName,
			FileSize:    m.FileSize,
			FileKey:     m.FileKey,
			FileTxnType: m.FileTxnType,
			FileTxnMeta: m.FileTxnMeta,
		}
		items = append(items, item)
	}

	var remaining *int64
	if companyRec.TotalUsageQuota != nil {
		total := *companyRec.TotalUsageQuota
		used := companyRec.UsedQuota
		r := total - used
		remaining = &r
	}

	resp := ListCompanyFilesResponse{
		Items:           items,
		TotalUsageQuota: companyRec.TotalUsageQuota,
		UsedQuota:       companyRec.UsedQuota,
		RemainingQuota:  remaining,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)

	_ = ctx // currently unused, but kept if you want to add timeouts/cancellation later
}

// DeleteFile godoc
// @Summary      Delete a single file by key
// @Description  Deletes a file from S3 for the authenticated company and records a files_meta entry with file_txn_type=2
// @Tags         uploader
// @Accept       json
// @Produce      json
// @Param        X-API-Key  header    string             true  "Company API key"
// @Param        body       body      DeleteFileRequest  true  "Delete file request"
// @Success      200        {object}  DeleteFileResponse
// @Failure      400        {string}  string "invalid request"
// @Failure      401        {string}  string "unauthorized"
// @Failure      500        {string}  string "internal error"
// @Router       /uploader/files/delete [post]
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "missing X-API-Key header", http.StatusUnauthorized)
		return
	}

	companyRec, err := h.companyRepo.GetByAPIKey(apiKey)
	if err != nil {
		http.Error(w, "failed to look up company", http.StatusInternalServerError)
		return
	}
	if companyRec == nil {
		http.Error(w, "invalid API key", http.StatusUnauthorized)
		return
	}

	var req DeleteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.FileKey == "" {
		http.Error(w, "file_key is required", http.StatusBadRequest)
		return
	}

	// Basic safety: ensure this key belongs to this company
	expectedPrefix := fmt.Sprintf("companies/%s/", companyRec.ID)
	if !strings.HasPrefix(req.FileKey, expectedPrefix) {
		http.Error(w, "file_key does not belong to this company", http.StatusForbidden)
		return
	}

	// Delete from S3
	if err := h.s3Service.DeleteObject(ctx, companyRec, req.FileKey); err != nil {
		http.Error(w, "failed to delete file from storage", http.StatusInternalServerError)
		return
	}

	// Create a files_meta record for this delete (file_txn_type=2)
	// We don't know size here unless you want to include it in the request; store 0.
	txnMeta := req.FileTxnMeta
	meta := &filemeta.FileMeta{
		ID:          utils.GenerateID(),
		FileName:    nil,
		FileSize:    0,
		FileKey:     req.FileKey,
		FileTxnType: 2,
		FileTxnMeta: txnMeta,
		CompanyID:   &companyRec.ID,
	}
	if err := h.fileMetaRepo.Create(meta); err != nil {
		http.Error(w, "failed to create delete file meta", http.StatusInternalServerError)
		return
	}

	resp := DeleteFileResponse{
		FileKey: req.FileKey,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// DeleteFolder godoc
// @Summary      Delete all files under a folder (prefix)
// @Description  Deletes all objects under folder_prefix for the authenticated company and records a files_meta entry with file_txn_type=3
// @Tags         uploader
// @Accept       json
// @Produce      json
// @Param        X-API-Key  header    string               true  "Company API key"
// @Param        body       body      DeleteFolderRequest  true  "Delete folder request"
// @Success      200        {object}  DeleteFolderResponse
// @Failure      400        {string}  string "invalid request"
// @Failure      401        {string}  string "unauthorized"
// @Failure      500        {string}  string "internal error"
// @Router       /uploader/folders/delete [post]
func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "missing X-API-Key header", http.StatusUnauthorized)
		return
	}

	companyRec, err := h.companyRepo.GetByAPIKey(apiKey)
	if err != nil {
		http.Error(w, "failed to look up company", http.StatusInternalServerError)
		return
	}
	if companyRec == nil {
		http.Error(w, "invalid API key", http.StatusUnauthorized)
		return
	}

	var req DeleteFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.FolderPrefix == "" {
		http.Error(w, "folder_prefix is required", http.StatusBadRequest)
		return
	}

	// Ensure prefix belongs to this company
	expectedPrefix := fmt.Sprintf("%s/%s/", companyRec.CompanySlug, req.FolderPrefix)

	deletedCount, deletedBytes, err := h.s3Service.DeletePrefix(ctx, companyRec, expectedPrefix)
	if err != nil {
		http.Error(w, "failed to delete files from storage", http.StatusInternalServerError)
		return
	}

	// Record a single files_meta entry, representing this bulk delete (file_txn_type=3)
	// file_key stores the folder_prefix, file_size = deletedBytes
	txnMeta := req.FileTxnMeta
	meta := &filemeta.FileMeta{
		ID:          utils.GenerateID(),
		FileName:    nil,
		FileSize:    deletedBytes,
		FileKey:     req.FolderPrefix,
		FileTxnType: 3,
		FileTxnMeta: txnMeta,
		CompanyID:   &companyRec.ID,
	}
	if err := h.fileMetaRepo.Create(meta); err != nil {
		http.Error(w, "failed to create folder delete meta", http.StatusInternalServerError)
		return
	}

	if err := h.companyRepo.ResetUsedQuota(companyRec.ID); err != nil {
		// We already created meta, but quota update failed -> log and continue
		// In a stricter system you'd wrap this whole operation in a DB transaction.
		http.Error(w, "failed to update quota", http.StatusInternalServerError)
		return
	}

	resp := DeleteFolderResponse{
		FolderPrefix: req.FolderPrefix,
		DeletedCount: deletedCount,
		DeletedBytes: deletedBytes,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
