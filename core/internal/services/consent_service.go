package services

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"
)

var allowedScopes = map[string]bool{
	"email":        true,
	"profile":      true,
	"medical_data": true,
}

type ConsentService struct {
	ConsentRepo *repository.ConsentRepository
	AppRepo     *repository.AppRepository
}

func (s *ConsentService) CreateConsent(ctx context.Context, userID string, req dtos.CreateConsentRequest) (*dtos.ConsentResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify application exists
	if _, err := s.AppRepo.GetAppByID(ctx, req.AppID); err != nil {
		return nil, errors.New("invalid app_id: application does not exist")
	}

	// Validate scopes
	for _, scope := range req.Scopes {
		if !allowedScopes[scope] {
			return nil, errors.New("unsupported scope requested: " + scope)
		}
	}

	scopesJSON, err := json.Marshal(req.Scopes)
	if err != nil {
		return nil, errors.New("failed to serialize scopes")
	}

	var policiesJSON []byte
	if req.Policies == nil {
		policiesJSON = []byte("{}")
	} else {
		policiesJSON, err = json.Marshal(req.Policies)
		if err != nil {
			return nil, errors.New("failed to serialize policies")
		}
	}

	expiresAt, _ := time.Parse(time.RFC3339, req.ExpiresAt)

	id, createdAt, err := s.ConsentRepo.CreateConsent(ctx, userID, req.AppID, string(scopesJSON), req.Purpose, string(policiesJSON), expiresAt)
	if err != nil {
		return nil, err
	}

	return &dtos.ConsentResponse{
		ConsentID: id,
		AppID:     req.AppID,
		Scopes:    req.Scopes,
		Purpose:   req.Purpose,
		Status:    "ACTIVE",
		ExpiresAt: expiresAt.Format(time.RFC3339),
		CreatedAt: createdAt.Format(time.RFC3339),
		Policies:  req.Policies,
	}, nil
}

func (s *ConsentService) GetConsent(ctx context.Context, id, userID string) (*dtos.ConsentResponse, error) {
	record, err := s.ConsentRepo.GetConsentByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err // Returns "consent not found"
	}
	return s.mapToDTO(record), nil
}

func (s *ConsentService) ListConsents(ctx context.Context, userID string) ([]dtos.ConsentResponse, error) {
	records, err := s.ConsentRepo.ListConsentsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dtos.ConsentResponse, 0, len(records))
	for _, r := range records {
		responses = append(responses, *s.mapToDTO(r))
	}
	return responses, nil
}

func (s *ConsentService) RevokeConsent(ctx context.Context, id, userID string) error {
	return s.ConsentRepo.RevokeConsent(ctx, id, userID)
}

func (s *ConsentService) CheckConsent(ctx context.Context, req dtos.CheckConsentRequest) (*dtos.CheckConsentResponse, error) {
	record, err := s.ConsentRepo.GetActiveConsentForApp(ctx, req.UserID, req.AppID)
	if err != nil {
		// No active consent found
		return &dtos.CheckConsentResponse{Allowed: false}, nil
	}

	// Double check expiration at read time
	if record.ExpiresAt.Before(time.Now()) || record.ExpiresAt.Equal(time.Now()) {
		return &dtos.CheckConsentResponse{Allowed: false}, nil
	}
	if record.Status != "ACTIVE" {
		return &dtos.CheckConsentResponse{Allowed: false}, nil
	}

	var scopes []string
	if err := json.Unmarshal([]byte(record.Scopes), &scopes); err != nil {
		return &dtos.CheckConsentResponse{Allowed: false}, nil
	}

	var policies map[string]interface{}
	if record.Policies != "" {
		if err := json.Unmarshal([]byte(record.Policies), &policies); err == nil {
			if timeStr, ok := policies["time_of_day"].(string); ok {
				// Evaluate time_of_day
				parts := strings.Split(timeStr, "-")
				if len(parts) == 2 {
					now := time.Now().UTC()
					startParts := strings.Split(parts[0], ":")
					endParts := strings.Split(parts[1], ":")
					
					if len(startParts) == 2 && len(endParts) == 2 {
						startH, _ := strconv.Atoi(startParts[0])
						startM, _ := strconv.Atoi(startParts[1])
						endH, _ := strconv.Atoi(endParts[0])
						endM, _ := strconv.Atoi(endParts[1])
						
						currentMins := now.Hour()*60 + now.Minute()
						startMins := startH*60 + startM
						endMins := endH*60 + endM
						
						if currentMins < startMins || currentMins > endMins {
							return &dtos.CheckConsentResponse{Allowed: false}, nil
						}
					}
				}
			}
		}
	}

	// Exact scope match
	for _, grantedScope := range scopes {
		if grantedScope == req.Scope {
			return &dtos.CheckConsentResponse{Allowed: true}, nil
		}
	}

	return &dtos.CheckConsentResponse{Allowed: false}, nil
}

func (s *ConsentService) mapToDTO(record *repository.ConsentRecord) *dtos.ConsentResponse {
	var scopes []string
	json.Unmarshal([]byte(record.Scopes), &scopes)

	var policies map[string]interface{}
	if record.Policies != "" {
		json.Unmarshal([]byte(record.Policies), &policies)
	}

	status := record.Status
	// Read-time expiration evaluation
	if status == "ACTIVE" && (record.ExpiresAt.Before(time.Now()) || record.ExpiresAt.Equal(time.Now())) {
		status = "EXPIRED"
	}

	resp := &dtos.ConsentResponse{
		ConsentID: record.ID,
		AppID:     record.AppID,
		Scopes:    scopes,
		Purpose:   record.Purpose,
		Status:    status,
		ExpiresAt: record.ExpiresAt.Format(time.RFC3339),
		CreatedAt: record.CreatedAt.Format(time.RFC3339),
		Policies:  policies,
	}
	if record.RevokedAt.Valid {
		rAt := record.RevokedAt.Time.Format(time.RFC3339)
		resp.RevokedAt = &rAt
	}
	return resp
}
