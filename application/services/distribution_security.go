package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func encryptDistributionSecret(plainText string) (string, error) {
	key, err := distributionCipherKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce failed: %w", err)
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decryptDistributionSecret(cipherText string) (string, error) {
	key, err := distributionCipherKey()
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext failed: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm failed: %w", err)
	}

	if len(decoded) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := decoded[:gcm.NonceSize()]
	data := decoded[gcm.NonceSize():]
	plainText, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secret failed: %w", err)
	}

	return string(plainText), nil
}

func distributionCipherKey() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("DISTRIBUTION_SECRET_KEY"))
	if secret == "" {
		return nil, fmt.Errorf("DISTRIBUTION_SECRET_KEY 未配置")
	}

	sum := sha256.Sum256([]byte(secret))
	return sum[:], nil
}
