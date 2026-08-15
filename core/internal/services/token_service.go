package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type TokenService struct {
	TokenRepo      *repository.TokenRepository
	AppRepo        *repository.AppRepository
	ConsentService *ConsentService
	AuditService   *AuditService
}

func (s *TokenService) GenerateToken(ctx context.Context, userID, appID, appSecret, scope string) (string, error) {
	// 1. Authenticate App
	app, err := s.AppRepo.GetAppByID(ctx, appID)
	if err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "TOKEN_REJECTED", scope)
		return "", errors.New("unauthorized: invalid application credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(app.SecretHash), []byte(appSecret))
	if err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "TOKEN_REJECTED", scope)
		return "", errors.New("unauthorized: invalid application credentials")
	}

	// 2. Validate Scope (Only allow predefined MVP scopes)
	validScopes := map[string]bool{"email": true, "profile": true, "medical_data": true}
	if !validScopes[scope] {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "TOKEN_REJECTED", scope)
		return "", errors.New("forbidden: invalid scope requested")
	}

	// 3. Verify active Consent Manager status
	consentReq := dtos.CheckConsentRequest{
		UserID: userID,
		AppID:  appID,
		Scope:  scope,
	}
	_, err = s.ConsentService.CheckConsent(ctx, consentReq)
	if err != nil {
		s.AuditService.LogAccessAttempt(ctx, userID, appID, "TOKEN_REJECTED_CONSENT", scope)
		return "", fmt.Errorf("forbidden: consent validation failed: %v", err)
	}

	// 4. Generate random 32-byte token
	rawBytes := make([]byte, 32)
	_, err = rand.Read(rawBytes)
	if err != nil {
		return "", errors.New("internal error generating token")
	}
	rawToken := base64.RawURLEncoding.EncodeToString(rawBytes)

	// 5. Hash token for storage
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	// 6. Set short expiration (15 mins)
	expiresAt := time.Now().Add(15 * time.Minute)

	// 7. Store Token
	_, err = s.TokenRepo.CreateToken(ctx, userID, appID, tokenHash, scope, expiresAt)
	if err != nil {
		return "", errors.New("internal error saving token")
	}

	// 8. Audit Event
	s.AuditService.LogAccessAttempt(ctx, userID, appID, "TOKEN_ISSUED", scope)

	return rawToken, nil
}

func (s *TokenService) RevokeToken(ctx context.Context, userID, rawToken string) error {
	// Hash the raw token
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Try to revoke (enforces ownership inside repo)
	err := s.TokenRepo.RevokeToken(ctx, tokenHash, userID)
	if err != nil {
		// Log attempt anyway if possible (we don't have AppID, use "")
		s.AuditService.LogAccessAttempt(ctx, userID, "", "TOKEN_REVOKE_FAILED", "")
		return errors.New("forbidden or invalid token")
	}

	s.AuditService.LogAccessAttempt(ctx, userID, "", "TOKEN_REVOKED", "")
	return nil
}
