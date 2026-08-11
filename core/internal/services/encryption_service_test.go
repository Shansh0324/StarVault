package services

import (
	"context"
	"bytes"
	"crypto/rand"
	"strings"
	"testing"
)

func TestEncryptionService(t *testing.T) {
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	svc, err := NewEncryptionService(keyBytes)
	if err != nil {
		t.Fatalf("Failed to init service: %v", err)
	}

	plaintext := []byte("secret medical data")

	// 1. Encrypt + decrypt returns original plaintext
	payload, err := svc.Encrypt(context.Background(), plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	decrypted, err := svc.Decrypt(context.Background(), payload)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("Expected %s, got %s", string(plaintext), string(decrypted))
	}

	// 2. Fresh nonces (same plaintext -> different ciphertext)
	payload2, _ := svc.Encrypt(context.Background(), plaintext)
	if payload == payload2 {
		t.Fatal("Ciphertexts match for same plaintext. Nonce is not fresh!")
	}

	// 3. Tampered ciphertext fails
	parts := strings.Split(payload, ":")
	parts[1] = "A" + parts[1][1:] // tamper first character of ciphertext
	tamperedPayload := parts[0] + ":" + parts[1]
	_, err = svc.Decrypt(context.Background(), tamperedPayload)
	if err == nil {
		t.Fatal("Expected decryption to fail on tampered ciphertext, but it succeeded")
	}

	// 4. Invalid key length
	_, err = NewEncryptionService([]byte("too_short"))
	if err == nil {
		t.Fatal("Expected failure for invalid key length")
	}

	// 5. Empty plaintext
	emptyPayload, _ := svc.Encrypt(context.Background(), []byte(""))
	emptyDecrypted, _ := svc.Decrypt(context.Background(), emptyPayload)
	if len(emptyDecrypted) != 0 {
		t.Fatalf("Expected empty decrypted result, got length %d", len(emptyDecrypted))
	}
}
