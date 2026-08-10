package services

import (
	"starvault/core/internal/dtos"
	"testing"
)

func TestAccessService_DTOValidation(t *testing.T) {
	req := dtos.AccessDataRequest{
		AppID:       "app-123",
		Secret:      "sec-123",
		Scope:       "email",
		VaultDataID: "vault-123",
	}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid DTO, got error: %v", err)
	}

	reqBad := dtos.AccessDataRequest{
		AppID:       "",
		Secret:      "sec-123",
		Scope:       "email",
		VaultDataID: "vault-123",
	}
	if err := reqBad.Validate(); err == nil {
		t.Errorf("expected validation error for empty AppID")
	}
}
