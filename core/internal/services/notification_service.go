package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type NotificationService struct {
	JetStream nats.JetStreamContext
}

type AccessNotification struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	AppID     string `json:"appId"`
	Action    string `json:"action"`
	Scope     string `json:"scope"`
	Timestamp string `json:"timestamp"`
}

func (s *NotificationService) PublishAccessEvent(ctx context.Context, userID, appID, action, scope string) {
	if s.JetStream == nil {
		log.Println("NotificationService: JetStream not configured, skipping notification.")
		return
	}

	event := AccessNotification{
		ID:        uuid.New().String(),
		UserID:    userID,
		AppID:     appID,
		Action:    action,
		Scope:     scope,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("ERROR: Failed to serialize notification event: %v", err)
		return
	}

	// Topic structure: notifications.access.<userID>
	subject := fmt.Sprintf("notifications.access.%s", userID)

	_, err = s.JetStream.Publish(subject, payload)
	if err != nil {
		log.Printf("ERROR: Failed to publish notification to NATS: %v", err)
	} else {
		log.Printf("NotificationService: Published to %s", subject)
	}
}
