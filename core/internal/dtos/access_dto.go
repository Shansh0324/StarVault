package dtos

import "errors"

type AccessDataRequest struct {
	AppID       string `json:"appId"`
	Secret      string `json:"secret"`
	Scope       string `json:"scope"`
	VaultDataID string `json:"vaultDataId"`
}

func (r *AccessDataRequest) Validate() error {
	if r.AppID == "" {
		return errors.New("appId is required")
	}
	if r.Secret == "" {
		return errors.New("secret is required")
	}
	if r.Scope == "" {
		return errors.New("scope is required")
	}
	if r.VaultDataID == "" {
		return errors.New("vaultDataId is required")
	}
	return nil
}
