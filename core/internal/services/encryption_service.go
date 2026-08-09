package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

type EncryptionService struct {
	key []byte
}

func NewEncryptionService(hexKey []byte) (*EncryptionService, error) {
	if len(hexKey) != 32 {
		return nil, errors.New("master key must be exactly 32 bytes")
	}
	return &EncryptionService{key: hexKey}, nil
}

func (s *EncryptionService) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	
	// Format: base64(nonce):base64(ciphertext)
	return fmt.Sprintf("%s:%s", base64.StdEncoding.EncodeToString(nonce), base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func (s *EncryptionService) Decrypt(payload string) ([]byte, error) {
	parts := strings.Split(payload, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid encrypted payload format")
	}

	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("failed to decode nonce")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("failed to decode ciphertext")
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aesGCM.NonceSize() {
		return nil, errors.New("invalid nonce length")
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed or authentication tag mismatch")
	}

	return plaintext, nil
}
