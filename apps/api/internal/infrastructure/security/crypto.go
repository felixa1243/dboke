package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

// Encrypt encrypts a plaintext string using AES-256-GCM.
// The masterKey must be exactly 32 bytes long.
func Encrypt(plaintext, masterKey string) (string, error) {
	if len(masterKey) != 32 {
		return "", errors.New("master key must be exactly 32 bytes")
	}

	block, err := aes.NewCipher([]byte(masterKey))
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

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts an AES-256-GCM encrypted hex string.
// The masterKey must be exactly 32 bytes long.
func Decrypt(encryptedHex, masterKey string) (string, error) {
	if len(masterKey) != 32 {
		return "", errors.New("master key must be exactly 32 bytes")
	}

	enc, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher([]byte(masterKey))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(enc) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
