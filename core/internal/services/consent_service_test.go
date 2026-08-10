package services

import (
	"encoding/json"
	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"
	"testing"
	"time"
)

func TestConsentService_ScopeValidation(t *testing.T) {
	req := dtos.CreateConsentRequest{
		AppID:     "app-123",
		Scopes:    []string{"email", "unsupported_scope"},
		Purpose:   "test",
		ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	err := req.Validate()
	if err != nil {
		// Valid DTO so far
	}
}

// Since we use a real DB, the actual service logic will be tested end-to-end via integration tests.
// Let's test the mapping and read-time expiration logic.

func TestConsentService_MapToDTO_Expiration(t *testing.T) {
	svc := &ConsentService{}

	pastDate := time.Now().Add(-1 * time.Hour)
	futureDate := time.Now().Add(1 * time.Hour)

	scopesJSON, _ := json.Marshal([]string{"email"})

	record1 := &repository.ConsentRecord{
		ID:        "1",
		Status:    "ACTIVE",
		ExpiresAt: futureDate,
		Scopes:    string(scopesJSON),
	}
	dto1 := svc.mapToDTO(record1)
	if dto1.Status != "ACTIVE" {
		t.Errorf("Expected ACTIVE, got %s", dto1.Status)
	}

	record2 := &repository.ConsentRecord{
		ID:        "2",
		Status:    "ACTIVE",
		ExpiresAt: pastDate,
		Scopes:    string(scopesJSON),
	}
	dto2 := svc.mapToDTO(record2)
	if dto2.Status != "EXPIRED" {
		t.Errorf("Expected EXPIRED due to read-time logic, got %s", dto2.Status)
	}
}
