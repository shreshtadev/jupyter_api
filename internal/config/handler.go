package config

import (
	"encoding/json"
	"net/http"
)

type CreateAdminClientRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type CreateAdminClientResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type ValidateAdminClientResponse struct {
	IsValid bool `json:"is_valid"`
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) CreateAdminClient(w http.ResponseWriter, r *http.Request) {
	var req CreateAdminClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ClientID == "" || req.ClientSecret == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	adminClient := &AdminClient{
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
	}

	if err := h.repo.Create(adminClient); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateAdminClientResponse{ClientID: adminClient.ClientID, ClientSecret: adminClient.ClientSecret})
}

func (h *Handler) ValidateAdminClient(w http.ResponseWriter, r *http.Request) {
	var req CreateAdminClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
	}

	if req.ClientID == "" || req.ClientSecret == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
	}
	adminClient, err := h.repo.FindBy(req.ClientID, req.ClientSecret)
	if err != nil {
		http.Error(w, "database error", http.StatusUnauthorized)
		return
	}
	if adminClient == nil || adminClient.ClientSecret != req.ClientSecret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateAdminClientResponse{IsValid: true})
}
