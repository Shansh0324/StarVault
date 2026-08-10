package dtos

import (
	"errors"
	"strings"
	"time"
)

type CreateConsentRequest struct {
	AppID     string   `json:"appId"`
	Scopes    []string `json:"scopes"`
	Purpose   string   `json:"purpose"`
	ExpiresAt string   `json:"expiresAt"`
}

func (r *CreateConsentRequest) Validate() error {
	r.AppID = strings.TrimSpace(r.AppID)
	r.Purpose = strings.TrimSpace(r.Purpose)

	if r.AppID == "" {
		return errors.New("appId is required")
	}
	if len(r.Scopes) == 0 {
		return errors.New("scopes are required")
	}
	if r.Purpose == "" {
		return errors.New("purpose is required")
	}
	if len(r.Purpose) > 500 {
		return errors.New("purpose is too long")
	}

	exp, err := time.Parse(time.RFC3339, r.ExpiresAt)
	if err != nil {
		return errors.New("invalid expiresAt format")
	}
	if exp.Before(time.Now()) || exp.Equal(time.Now()) {
		return errors.New("expiresAt must be in the future")
	}

	// Check for duplicate scopes
	seen := make(map[string]bool)
	for i, scope := range r.Scopes {
		s := strings.TrimSpace(scope)
		if s == "" {
			return errors.New("empty scope is not allowed")
		}
		if seen[s] {
			return errors.New("duplicate scopes are not allowed")
		}
		seen[s] = true
		r.Scopes[i] = s
	}

	return nil
}

type ConsentResponse struct {
	ConsentID string   `json:"consentId"`
	AppID     string   `json:"appId"`
	Scopes    []string `json:"scopes"`
	Purpose   string   `json:"purpose"`
	Status    string   `json:"status"`
	ExpiresAt string   `json:"expiresAt"`
	CreatedAt string   `json:"createdAt"`
	RevokedAt *string  `json:"revokedAt,omitempty"`
}

type CheckConsentRequest struct {
	UserID string `json:"userId"`
	AppID  string `json:"appId"`
	Scope  string `json:"scope"`
}

type CheckConsentResponse struct {
	Allowed bool `json:"allowed"`
}
