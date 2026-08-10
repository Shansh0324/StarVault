package dtos

import (
	"errors"
	"strings"
)

type CreateAppRequest struct {
	Name string `json:"name"`
}

func (r *CreateAppRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return errors.New("name is required")
	}
	if len(r.Name) > 100 {
		return errors.New("name is too long")
	}
	return nil
}

type CreateAppResponse struct {
	AppID  string `json:"appId"`
	Name   string `json:"name"`
	Secret string `json:"secret"` // Plaintext secret returned only once
}
