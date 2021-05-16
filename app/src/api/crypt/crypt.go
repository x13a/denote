package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

const minDataLen = 12  // nonce

var ErrDataTooShort = errors.New("data too short")

func EncryptGCM(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func DecryptGCM(key, data []byte) ([]byte, error) {
	if len(data) < minDataLen {
		return nil, ErrDataTooShort
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	return gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
}
