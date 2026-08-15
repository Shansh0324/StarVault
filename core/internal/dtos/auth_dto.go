package dtos

import (
	"errors"
	"strings"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password  string `json:"password"`
	IPAddress string `json:"ipAddress,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`
}

func (r *AuthRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	if r.Email == "" || r.Password == "" {
		return errors.New("Email and password are required")
	}
	return nil
}

type AuthResponse struct {
	UserID string `json:"userId,omitempty"`
	Error  string `json:"error,omitempty"`
}
