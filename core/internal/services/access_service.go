package services

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"
)

type AccessService struct {
	AppRepo        *repository.AppRepository
	ConsentService *ConsentService
	VaultService   *VaultService
}

func (s *AccessService) AccessData(userID string, req dtos.AccessDataRequest) (*dtos.VaultDataResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Authenticate Application
	appRecord, err := s.AppRepo.GetAppByID(req.AppID)
	if err != nil {
		return nil, errors.New("unauthorized: invalid application credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(appRecord.SecretHash), []byte(req.Secret)); err != nil {
		return nil, errors.New("unauthorized: invalid application credentials")
	}

	// 2. Check Consent
	consentReq := dtos.CheckConsentRequest{
		UserID: userID,
		AppID:  req.AppID,
		Scope:  req.Scope,
	}
	checkRes, err := s.ConsentService.CheckConsent(consentReq)
	if err != nil {
		return nil, errors.New("forbidden: error checking consent")
	}
	if !checkRes.Allowed {
		return nil, errors.New("forbidden: access denied by consent manager")
	}

	// 3. Retrieve and Decrypt Vault Data
	// VaultService.GetVaultData enforces IDOR because it accepts userID.
	vaultData, err := s.VaultService.GetVaultData(req.VaultDataID, userID)
	if err != nil {
		return nil, errors.New("forbidden: vault data not found or access denied")
	}

	// 4. Validate Scope vs DataType
	if vaultData.DataType != req.Scope {
		return nil, errors.New("forbidden: requested scope does not match vault data type")
	}

	return vaultData, nil
}
