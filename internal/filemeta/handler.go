package filemeta

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"shreshtasmg.in/jupyter/internal/utils"
)

// Request DTO
type CreateFileMetaRequest struct {
	FileName    *string `json:"file_name,omitempty"`
	FileSize    int64   `json:"file_size"`     // required
	FileKey     string  `json:"file_key"`      // required (e.g. S3 key)
	FileTxnType int16   `json:"file_txn_type"` // required (e.g. 1=upload,2=delete, etc.)
	FileTxnMeta *string `json:"file_txn_meta,omitempty"`
	CompanyID   *string `json:"company_id,omitempty"` // can be nil if not tied to a company
}

type FileMetaResponse struct {
	ID          string  `json:"id"`
	CreatedAt   string  `json:"created_at"`
	FileName    *string `json:"file_name,omitempty"`
	FileSize    int64   `json:"file_size"`
	FileKey     string  `json:"file_key"`
	FileTxnType int16   `json:"file_txn_type"`
	FileTxnMeta *string `json:"file_txn_meta,omitempty"`
	CompanyID   *string `json:"company_id,omitempty"`
}

// Response DTO
type CreateFileMetaResponse struct {
	ID string `json:"id"`
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateFileMeta godoc
// @Summary      Create file metadata
// @Description  Stores metadata for a file transaction (e.g. upload to S3)
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        body  body      CreateFileMetaRequest  true  "File meta payload"
// @Success      201   {object}  CreateFileMetaResponse
// @Failure      400   {string}  string "invalid request"
// @Failure      500   {string}  string "internal error"
// @Router       /files/meta [post]
func (h *Handler) CreateFileMeta(w http.ResponseWriter, r *http.Request) {
	var req CreateFileMetaRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.FileSize <= 0 {
		http.Error(w, "file_size must be > 0", http.StatusBadRequest)
		return
	}
	if req.FileKey == "" {
		http.Error(w, "file_key is required", http.StatusBadRequest)
		return
	}
	if req.FileTxnType == 0 {
		http.Error(w, "file_txn_type is required", http.StatusBadRequest)
		return
	}

	id := utils.GenerateID()

	meta := &FileMeta{
		ID:          id,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		FileKey:     req.FileKey,
		FileTxnType: req.FileTxnType,
		FileTxnMeta: req.FileTxnMeta,
		CompanyID:   req.CompanyID,
	}

	if err := h.repo.Create(meta); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	resp := CreateFileMetaResponse{ID: id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetFileMetaByID godoc
// @Summary      Get file meta by ID
// @Description  Fetch a single files_meta row by its ID
// @Tags         files
// @Produce      json
// @Param        id   path      string             true  "File ID"
// @Success      200  {object}  FileMetaResponse
// @Failure      404  {string}  string "not found"
// @Failure      500  {string}  string "internal error"
// @Router       /files/meta/{id} [get]
func (h *Handler) GetFileMetaByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	meta, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "failed to fetch file meta", http.StatusInternalServerError)
		return
	}
	if meta == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp := FileMetaResponse{
		ID:          meta.ID,
		CreatedAt:   meta.CreatedAt.Format(time.RFC3339Nano),
		FileName:    meta.FileName,
		FileSize:    meta.FileSize,
		FileKey:     meta.FileKey,
		FileTxnType: meta.FileTxnType,
		FileTxnMeta: meta.FileTxnMeta,
		CompanyID:   meta.CompanyID,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
