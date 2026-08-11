package services

import (
	"context"
	"errors"

	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"
)

type AccessService struct {
	AppRepo        *repository.AppRepository
	ConsentService *ConsentService
	VaultService   *VaultService
	AuditService   *AuditService
	TokenRepo      *repository.TokenRepository
}

func (s *AccessService) AccessData(ctx context.Context, userID string, req dtos.AccessDataRequest) (*dtos.VaultDataResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Authenticate Application
	appRecord, err := s.AppRepo.GetAppByID(ctx, req.AppID)
	if err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, req.AppID, "ACCESS_DENIED_INVALID_APP", req.Scope)
		return nil, errors.New("unauthorized: invalid application credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(appRecord.SecretHash), []byte(req.Secret)); err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, req.AppID, "ACCESS_DENIED_INVALID_SECRET", req.Scope)
		return nil, errors.New("unauthorized: invalid application credentials")
	}

	// 2. Check Consent
	consentReq := dtos.CheckConsentRequest{
		UserID: userID,
		AppID:  req.AppID,
		Scope:  req.Scope,
	}
	checkRes, err := s.ConsentService.CheckConsent(ctx, consentReq)
	if err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, req.AppID, "ACCESS_DENIED_CONSENT_ERROR", req.Scope)
		return nil, errors.New("forbidden: error checking consent")
	}
	if !checkRes.Allowed {
		s.AuditService.LogAccessAttempt(ctx, userID, req.AppID, "ACCESS_DENIED_CONSENT_REJECTED", req.Scope)
		return nil, errors.New("forbidden: access denied by consent manager")
	}

	// 3. Retrieve and Decrypt Vault Data
	// VaultService.GetVaultData enforces IDOR because it accepts userID.
	vaultData, err := s.VaultService.GetVaultData(ctx, req.VaultDataID, userID)
	if err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, req.AppID, "ACCESS_DENIED_VAULT_NOT_FOUND", req.Scope)
		return nil, errors.New("forbidden: vault data not found or access denied")
	}

	// 4. Validate Scope vs DataType
	if vaultData.DataType != req.Scope {
		s.AuditService.LogAccessAttempt(ctx, userID, req.AppID, "ACCESS_DENIED_SCOPE_MISMATCH", req.Scope)
		return nil, errors.New("forbidden: requested scope does not match vault data type")
	}

	s.AuditService.LogAccessAttempt(ctx, userID, req.AppID, "ACCESS_GRANTED", req.Scope)
	return vaultData, nil
}

func (s *AccessService) AccessDataWithToken(ctx context.Context, rawToken string, req dtos.AccessDataRequest) (*dtos.VaultDataResponse, error) {
	if rawToken == "" || req.VaultDataID == "" || req.Scope == "" {
		return nil, errors.New("bad request: missing required fields")
	}

	// 1. Hash the token
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	// 2. Retrieve Token
	tokenRecord, err := s.TokenRepo.GetTokenByHash(ctx, tokenHash)
	if err != nil {
		// Cannot audit: no valid user/app UUIDs available for an unknown token
		return nil, errors.New("unauthorized: invalid or missing token")
	}

	userID := tokenRecord.UserID
	appID := tokenRecord.AppID

	// 3. Validate Token Status
	if tokenRecord.RevokedAt != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "ACCESS_DENIED_REVOKED_TOKEN", req.Scope)
		return nil, errors.New("unauthorized: token has been revoked")
	}
	if time.Now().After(tokenRecord.ExpiresAt) {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "ACCESS_DENIED_EXPIRED_TOKEN", req.Scope)
		return nil, errors.New("unauthorized: token has expired")
	}

	// 4. Validate Scope
	if !strings.Contains(tokenRecord.Scopes, req.Scope) {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "ACCESS_DENIED_TOKEN_SCOPE_MISMATCH", req.Scope)
		return nil, errors.New("forbidden: token does not grant requested scope")
	}

	// 5. Re-Check Consent Manager
	consentReq := dtos.CheckConsentRequest{
		UserID: userID,
		AppID:  appID,
		Scope:  req.Scope,
	}
	checkRes, err := s.ConsentService.CheckConsent(ctx, consentReq)
	if err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "ACCESS_DENIED_CONSENT_ERROR", req.Scope)
		return nil, errors.New("forbidden: error checking consent")
	}
	if !checkRes.Allowed {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "ACCESS_DENIED_CONSENT_REJECTED", req.Scope)
		return nil, errors.New("forbidden: access denied by consent manager")
	}

	// 6. Retrieve and Decrypt Vault Data
	vaultData, err := s.VaultService.GetVaultData(ctx, req.VaultDataID, userID)
	if err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "ACCESS_DENIED_VAULT_NOT_FOUND", req.Scope)
		return nil, errors.New("forbidden: vault data not found or access denied")
	}

	// 7. Validate Data Scope Match
	if vaultData.DataType != req.Scope {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "ACCESS_DENIED_DATA_SCOPE_MISMATCH", req.Scope)
		return nil, errors.New("forbidden: requested scope does not match vault data type")
	}

	s.AuditService.LogAccessAttempt(ctx, userID, appID, "ACCESS_GRANTED_WITH_TOKEN", req.Scope)
	return vaultData, nil
}

