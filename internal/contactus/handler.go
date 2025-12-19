package contactus

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"
)

type CreateContactUsRequest struct {
	FullName       string `json:"full_name"`
	ContactEmail   string `json:"contact_email"`
	ContactNumber  string `json:"contact_number"`
	ProjectType    string `json:"project_type"`
	ProjectDetails string `json:"project_details"`
}

type CreateContactUsResponse struct {
	Message   string `json:"msg"`
	CreatedAt string `json:"createdAt"`
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateContactUs godoc
// @Summary      Create contact us
// @Description  Creates a new contact_us entry
// @Tags         contactus
// @Accept       json
// @Produce      json
// @Param        body  body      CreateContactUsRequest  true  "ContactUs payload"
// @Success      201   {object}  CreateContactUsResponse
// @Failure      400   {string}  string "invalid request"
// @Failure      500   {string}  string "internal error"
// @Router       /contactus [post]
func (h *Handler) CreateContactUs(w http.ResponseWriter, r *http.Request) {
	var req CreateContactUsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ContactEmail == "" || req.ContactNumber == "" || req.FullName == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	emailRegex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
	if matched, _ := regexp.MatchString(emailRegex, req.ContactEmail); !matched {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}

	phoneRegex := `^(?:\+91|0)?[6789]\d{9}$`
	if matched, _ := regexp.MatchString(phoneRegex, req.ContactNumber); !matched {
		http.Error(w, "invalid phone number", http.StatusBadRequest)
		return
	}

	contactUs := &ContactUs{
		FullName:        req.FullName,
		ContactEmail:    req.ContactEmail,
		ContactNumber:   req.ContactNumber,
		ProjectType:     req.ProjectType,
		ProjectDetails:  req.ProjectDetails,
		ContactUsStatus: "submitted",
		Remarks:         "Contact Us Request submitted for review",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.repo.Create(contactUs); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateContactUsResponse{Message: "Contact us created successfully", CreatedAt: contactUs.CreatedAt.Format(time.RFC3339)})
}
