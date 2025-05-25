package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateHMAC генерирует HMAC-SHA256 подпись
func GenerateHMAC(data string, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyHMAC проверяет HMAC подпись
func VerifyHMAC(data, signature string, key []byte) (bool, error) {
	expected, err := GenerateHMAC(data, key)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(expected), []byte(signature)), nil
}
