package dtos

import (
	"errors"
	"strings"
)

type CreateVaultDataRequest struct {
	DataType string `json:"dataType"`
	Data     string `json:"data"`
}

func (r *CreateVaultDataRequest) Validate() error {
	r.DataType = strings.TrimSpace(r.DataType)
	r.Data = strings.TrimSpace(r.Data)
	if r.DataType == "" || r.Data == "" {
		return errors.New("dataType and data are required")
	}
	if len(r.Data) > 50000 {
		return errors.New("data payload exceeds maximum allowed size")
	}
	return nil
}

type VaultDataResponse struct {
	ID        string `json:"id"`
	DataType  string `json:"dataType"`
	Data      string `json:"data,omitempty"`
	CreatedAt string `json:"createdAt"`
}
