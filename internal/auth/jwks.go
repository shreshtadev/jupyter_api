package auth

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"
)

// JWK represents a single JSON Web Key
type JWK struct {
	Kty string `json:"kty"` // Key Type
	Use string `json:"use"` // Public Key Use
	Alg string `json:"alg"` // Algorithm
	Kid string `json:"kid"` // Key ID
	N   string `json:"n"`   // RSA Modulus
	E   string `json:"e"`   // RSA Exponent
}

// JWKSResponse represents the JWKS response structure
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// cachedJWKS will hold the parsed JWKS structure ready to send.
var (
	cachedJWKS      *JWKSResponse
	cachedJWKSBytes []byte // Pre-marshaled for performance
	cachedJWKSETag  string // ETag for conditional requests
	cachedJWKSMu    sync.RWMutex
)

// InitJWKS loads public key PEM from file (path from env or param) and prepares JWKS JSON.
// Call once at startup.
func InitJWKS() error {
	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		return errors.New("PUBLIC_KEY_PATH is not set")
	}

	pemBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return err
	}

	pub, err := ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return err
	}

	// create JWK
	jwk := JwkFromRSAPublicKey(pub, pemBytes)

	jwks := &JWKSResponse{
		Keys: []JWK{jwk},
	}

	// Pre-marshal for performance
	b, err := json.Marshal(jwks)
	if err != nil {
		return err
	}

	// Generate ETag from content hash
	etag := fmt.Sprintf(`"%s"`, base64.RawURLEncoding.EncodeToString(
		sha256.New().Sum(b)[:16],
	))

	cachedJWKSMu.Lock()
	cachedJWKS = jwks
	cachedJWKSBytes = b
	cachedJWKSETag = etag
	cachedJWKSMu.Unlock()

	return nil
}

// JWKSHandler serves the cached JWKS JSON with cache-control headers and conditional request support.
func JWKSHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET and HEAD methods
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cachedJWKSMu.RLock()
	jwks := cachedJWKS
	b := cachedJWKSBytes
	etag := cachedJWKSETag
	cachedJWKSMu.RUnlock()

	if jwks == nil || b == nil {
		// try to load once more (but log failures)
		if err := InitJWKS(); err != nil {
			log.Printf("JWKSHandler: initJWKS failed: %v", err)
			http.Error(w, "JWKS not available", http.StatusInternalServerError)
			return
		}
		cachedJWKSMu.RLock()
		jwks = cachedJWKS
		b = cachedJWKSBytes
		etag = cachedJWKSETag
		cachedJWKSMu.RUnlock()
		if jwks == nil || b == nil {
			http.Error(w, "JWKS not available", http.StatusInternalServerError)
			return
		}
	}

	// Set security headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=60")
	w.Header().Set("ETag", etag)

	// Check if client has cached version (conditional request)
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Use http.ServeContent for proper range request and conditional request support
	http.ServeContent(w, r, "jwks.json", time.Time{}, bytes.NewReader(b))
}

// helper: parse RSA public key from PEM bytes
func ParseRSAPublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	var pubInterface interface{}
	var err error
	switch block.Type {
	case "RSA PUBLIC KEY":
		pubInterface, err = x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		// try PKIX parse anyway
		pubInterface, err = x509.ParsePKIXPublicKey(block.Bytes)
	}
	if err != nil {
		return nil, err
	}
	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return pub, nil
}

// helper: build JWK from rsa.PublicKey
func JwkFromRSAPublicKey(pub *rsa.PublicKey, pemBytes []byte) JWK {
	// n: base64url(modulus)
	// e: base64url(exponent)
	nBytes := pub.N.Bytes()
	n := base64.RawURLEncoding.EncodeToString(nBytes)

	// exponent to big-endian bytes
	eInt := pub.E
	eBytes := big.NewInt(int64(eInt)).Bytes()
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	// simple kid: SHA256 of PEM (base64url). For RFC7638 thumbprint you'd use canonical JWK JSON + sha256.
	sum := sha256.Sum256(pemBytes)
	kid := base64.RawURLEncoding.EncodeToString(sum[:])

	return JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: kid,
		N:   n,
		E:   e,
	}
}

// GetJWKS returns the current cached JWKS response (useful for testing or direct access)
func GetJWKS() (*JWKSResponse, error) {
	cachedJWKSMu.RLock()
	defer cachedJWKSMu.RUnlock()

	if cachedJWKS == nil {
		return nil, errors.New("JWKS not initialized")
	}

	// Return a copy to prevent external modification
	response := &JWKSResponse{
		Keys: make([]JWK, len(cachedJWKS.Keys)),
	}
	copy(response.Keys, cachedJWKS.Keys)
	return response, nil
}

// StreamJWKSResponse streams the JWKS response using json.Encoder (alternative to pre-marshaled bytes)
// This can be useful if you want to dynamically modify the response before sending
func StreamJWKSResponse(w io.Writer) error {
	cachedJWKSMu.RLock()
	jwks := cachedJWKS
	cachedJWKSMu.RUnlock()

	if jwks == nil {
		return errors.New("JWKS not initialized")
	}

	return json.NewEncoder(w).Encode(jwks)
}
