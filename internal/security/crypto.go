package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// EncryptAES шифрует строку с помощью AES-256 в режиме CBC и PKCS7 padding
func EncryptAES(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	padded := pkcs7Pad([]byte(plaintext), blockSize)

	iv := make([]byte, blockSize) // Нулевой IV для упрощения, можно заменить на случайный
	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(padded))
	mode.CryptBlocks(encrypted, padded)

	return encodeBase64(encrypted), nil
}

// DecryptAES расшифровывает строку, зашифрованную AES-256 + CBC + PKCS7
func DecryptAES(encoded string, key []byte) (string, error) {
	data, err := decodeBase64(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	if len(data)%blockSize != 0 {
		return "", errors.New("invalid encrypted data length")
	}

	iv := make([]byte, blockSize)
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	unpadded, err := pkcs7Unpad(decrypted)
	if err != nil {
		return "", err
	}
	return string(unpadded), nil
}

// --- padding helpers ---

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, pad...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("invalid padding size")
	}
	padding := int(data[length-1])
	if padding > length {
		return nil, errors.New("invalid padding")
	}
	return data[:length-padding], nil
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func decodeBase64(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}
