package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RiskService struct {
	DB *sql.DB
}

type DevicePosture struct {
	IPAddress string
	UserAgent string
}

const (
	RiskLevelLow  = 0
	RiskLevelMed  = 50
	RiskLevelHigh = 100
)

// EvaluateLogin checks if the device is known. If not, assigns a risk score and saves the device.
func (s *RiskService) EvaluateLogin(ctx context.Context, userID string, posture DevicePosture) (int, error) {
	if posture.IPAddress == "" || posture.UserAgent == "" {
		return RiskLevelHigh, nil // High risk if posture data is missing
	}

	query := `
		SELECT id, risk_score
		FROM user_devices
		WHERE user_id = $1 AND ip_address = $2 AND user_agent = $3
	`
	var deviceID string
	var currentRisk int

	err := s.DB.QueryRowContext(ctx, query, userID, posture.IPAddress, posture.UserAgent).Scan(&deviceID, &currentRisk)
	if err == sql.ErrNoRows {
		// New device seen, save it as a high risk event initially
		err = s.recordNewDevice(ctx, userID, posture)
		if err != nil {
			return RiskLevelHigh, err
		}
		return RiskLevelHigh, nil
	} else if err != nil {
		return RiskLevelHigh, fmt.Errorf("failed to check device posture: %w", err)
	}

	// Known device, update last_seen
	_, _ = s.DB.ExecContext(ctx, "UPDATE user_devices SET last_seen_at = $1 WHERE id = $2", time.Now(), deviceID)

	return RiskLevelLow, nil
}

func (s *RiskService) recordNewDevice(ctx context.Context, userID string, posture DevicePosture) error {
	query := `
		INSERT INTO user_devices (id, user_id, ip_address, user_agent, risk_score, created_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`
	_, err := s.DB.ExecContext(ctx, query, uuid.New().String(), userID, posture.IPAddress, posture.UserAgent, RiskLevelHigh, time.Now())
	return err
}
