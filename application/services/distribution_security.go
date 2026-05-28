/**
 * 模块说明：数字丝路分发敏感信息加密工具。
 * 业务场景：Discord Webhook 等分发凭证需要随账号绑定保存，但不能以明文落库。
 * 核心职责：使用环境变量派生的 AES-GCM 密钥加解密分发密钥，供分发服务保存和发布时复用。
 */
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

/**
 * 功能：加密分发渠道敏感字段。
 * 参数：plainText 为用户录入的 Webhook 或第三方分发密钥。
 * 返回：可安全落库的密文；密钥未配置或加密失败时返回错误。
 */
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

/**
 * 功能：解密分发渠道敏感字段。
 * 参数：cipherText 为数据库中保存的密文。
 * 返回：发布任务调用外部平台所需的明文凭证。
 */
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

/**
 * 功能：从环境变量派生分发凭证加密密钥。
 * 参数：无。
 * 返回：32 字节 AES 密钥；缺少 DISTRIBUTION_SECRET_KEY 时返回错误，防止静默明文保存。
 */
func distributionCipherKey() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("DISTRIBUTION_SECRET_KEY"))
	if secret == "" {
		return nil, fmt.Errorf("DISTRIBUTION_SECRET_KEY 未配置")
	}

	sum := sha256.Sum256([]byte(secret))
	return sum[:], nil
}
