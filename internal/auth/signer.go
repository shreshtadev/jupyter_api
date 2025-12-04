// internal/auth/signer.go
package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTSigner struct {
	privateKey *rsa.PrivateKey
	issuer     string
	audience   string
	ttl        time.Duration
}

func NewJWTSignerFromFile(path, issuer, audience string, ttl time.Duration) (*JWTSigner, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM data")
	}

	var privKey *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse PKCS1 private key: %w", err)
		}
	default:
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse PKCS8 private key: %w", err2)
		}
		var ok bool
		privKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA")
		}
	}

	return &JWTSigner{
		privateKey: privKey,
		issuer:     issuer,
		audience:   audience,
		ttl:        ttl,
	}, nil
}

func randomJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *JWTSigner) GenerateToken(userID, companyID, email string, roles []string, publicKeyPath string) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(s.ttl)

	jti, err := randomJTI()
	if err != nil {
		return "", fmt.Errorf("generate jti: %w", err)
	}

	claims := UserClaims{
		UserID:    userID,
		CompanyID: companyID,
		Email:     email,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	if publicKeyPath == "" {
		return "", errors.New("PUBLIC_KEY_PATH is not set")
	}

	pemBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", errors.New("cannot read PUBLIC_KEY_PATH")
	}

	pub, err := ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return "", errors.New("cannot read PEM")
	}

	// create JWK
	jwk := JwkFromRSAPublicKey(pub, pemBytes)
	token.Header["kid"] = jwk.Kid
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}
