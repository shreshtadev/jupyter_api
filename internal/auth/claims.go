package auth

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	UserID    string   `json:"sub"`
	CompanyID string   `json:"company_id"` // tenant scope
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}
