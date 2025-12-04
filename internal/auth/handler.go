// internal/auth/handler.go
package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"shreshtasmg.in/jupyter/internal/user"
	"shreshtasmg.in/jupyter/internal/utils"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"` // "Bearer"
	ExpiresIn   int64    `json:"expires_in"` // seconds
	UserID      string   `json:"user_id"`
	CompanyID   string   `json:"company_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	CompanyID string `json:"company_id"`
	Role      string `json:"role"` // "admin" or "user"
}

type RegisterResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	CompanyID string `json:"company_id"`
	Role      string `json:"role"`
}

type MeResponse struct {
	UserID    string   `json:"user_id"`
	CompanyID *string  `json:"company_id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
}

type AuthHandler struct {
	userRepo user.Repository
	signer   *JWTSigner
	ttl      int64 // seconds
}

func NewAuthHandler(userRepo user.Repository, signer *JWTSigner) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		signer:   signer,
		ttl:      int64(signer.ttl.Seconds()),
	}
}

// Login godoc
// @Summary      Login with email and password
// @Description  Validates credentials and returns a JWT access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Login payload"
// @Success      200   {object}  LoginResponse
// @Failure      400   {string}  string "invalid request"
// @Failure      401   {string}  string "invalid credentials"
// @Failure      500   {string}  string "internal error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	u, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "failed to lookup user", http.StatusInternalServerError)
		return
	}
	if u == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := VerifyPassword(u.PasswordHash, req.Password); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	roles := []string{u.Role} // simple: single role
	var companyID string
	if u.CompanyID != nil {
		companyID = *u.CompanyID
	} else {
		companyID = "" // superadmin has no company
	}
	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		log.Fatal("PUBLIC_KEY_PATH is not set")
		return
	}
	token, err := h.signer.GenerateToken(u.ID, companyID, u.Email, roles, publicKeyPath)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   h.ttl,
		UserID:      u.ID,
		CompanyID:   companyID,
		Email:       u.Email,
		Roles:       roles,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a user with Argon2 password hash
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Register payload"
// @Success      201   {object}  RegisterResponse
// @Failure      400   {string}  string "invalid request"
// @Failure      409   {string}  string "email already exists"
// @Failure      500   {string}  string "internal error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.CompanyID == "" || req.Role == "" {
		http.Error(w, "email, password, company_id and role are required", http.StatusBadRequest)
		return
	}

	// Check for existing user
	existing, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "failed to check user", http.StatusInternalServerError)
		return
	}
	if existing != nil {
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	u := &user.User{
		ID:           utils.GenerateID(),
		Email:        req.Email,
		PasswordHash: hash,
		CompanyID:    &req.CompanyID,
		Role:         req.Role,
	}

	if err := h.userRepo.Create(u); err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	resp := RegisterResponse{
		UserID:    u.ID,
		Email:     u.Email,
		CompanyID: *u.CompanyID,
		Role:      u.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// Me godoc
// @Summary      Get current user info from JWT
// @Description  Returns user_id, company_id, email and roles from the access token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  MeResponse
// @Failure      401  {string}  string "unauthorized"
// @Router       /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := UserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ci := CompanyIDFromContext(r.Context())
	var companyIDPtr *string
	if ci != "" {
		companyIDPtr = &ci
	}

	resp := MeResponse{
		UserID:    userID,
		CompanyID: companyIDPtr,
		Email:     EmailFromContext(ctx),
		Roles:     RolesFromContext(ctx),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) JWKSJson(w http.ResponseWriter, r *http.Request) {
	JWKSHandler(w, r)
}
