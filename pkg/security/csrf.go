package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

// CSRFManager manages CSRF token creation, validation, and route exemptions.
type CSRFManager struct {
	CookieName  string
	HeaderNames []string
	FieldNames  []string
	ExemptPaths []string
}

// NewCSRFManager creates a new CSRFManager with Larapak standards.
func NewCSRFManager() *CSRFManager {
	return &CSRFManager{
		CookieName:  "larapak_csrf",
		HeaderNames: []string{"X-CSRF-TOKEN", "X-XSRF-TOKEN", "X-Csrf-Token"},
		FieldNames:  []string{"_token", "csrf_token", "_csrf"},
		ExemptPaths: []string{},
	}
}

// GenerateToken creates a 32-byte cryptographically secure hexadecimal token string.
func (m *CSRFManager) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsSafeMethod checks if the HTTP method is considered safe (GET, HEAD, OPTIONS, TRACE).
func (m *CSRFManager) IsSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// IsExempt checks whether the given request URL path is exempt from CSRF verification.
func (m *CSRFManager) IsExempt(path string) bool {
	for _, exempt := range m.ExemptPaths {
		if exempt == path {
			return true
		}
		if strings.HasSuffix(exempt, "*") && strings.HasPrefix(path, strings.TrimSuffix(exempt, "*")) {
			return true
		}
	}
	return false
}

// Validate compares the client-submitted token against the expected session token.
func (m *CSRFManager) Validate(expectedToken, submittedToken string) bool {
	if expectedToken == "" || submittedToken == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expectedToken), []byte(submittedToken)) == 1
}

// ExtractTokenFromRequest attempts to extract the token from headers or form data.
func (m *CSRFManager) ExtractTokenFromRequest(r *http.Request) string {
	// 1. Check HTTP Headers
	for _, header := range m.HeaderNames {
		if val := r.Header.Get(header); val != "" {
			return val
		}
	}

	// 2. Check URL Form / POST Body
	if r.Form != nil {
		for _, field := range m.FieldNames {
			if val := r.Form.Get(field); val != "" {
				return val
			}
		}
	}

	return ""
}

