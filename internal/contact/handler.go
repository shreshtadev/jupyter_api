package contact

import (
	"encoding/json"
	"net/http"

	"shreshtasmg.in/jupyter/internal/utils"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// DTO for incoming JSON
type CreateContactRequest struct {
	FullName       string `json:"full_name"`
	EmailAddress   string `json:"email_address"`
	ProjectType    string `json:"project_type"`
	ProjectDetails string `json:"project_details"`
}

type CreateContactResponse struct {
	EmailAddress string `json:"email_address"`
	CreatedAt    string `json:"created_at"`
}

// CreateContact godoc
// @Summary      Create a contact-us entry
// @Description  Stores a contact-us form submission
// @Tags         contact
// @Accept       json
// @Produce      json
// @Param        body  body      CreateContactRequest  true  "Contact request"
// @Success      201   {object}  CreateContactResponse
// @Failure      400   {string}  string  "invalid request"
// @Failure      500   {string}  string  "internal server error"
// @Router       /contact-us [post]
func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	var req CreateContactRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Simple validation
	if req.FullName == "" || req.EmailAddress == "" || req.ProjectType == "" || req.ProjectDetails == "" {
		http.Error(w, "full_name, project_type and project_details are required", http.StatusBadRequest)
		return
	}

	contact := &Contact{
		FullName:       req.FullName,
		EmailAddress:   req.EmailAddress,
		ProjectType:    req.ProjectType,
		ProjectDetails: req.ProjectDetails,
	}

	if err := h.repo.Create(contact); err != nil {
		http.Error(w, "failed to create contact", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := CreateContactResponse{EmailAddress: contact.EmailAddress, CreatedAt: utils.GetShortDate(contact.CreatedAt)}
	_ = json.NewEncoder(w).Encode(resp)
}
