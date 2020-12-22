package denote

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"os"

	"golang.org/x/crypto/argon2"
)

const (
	CryptoTime    = 1
	CryptoMemory  = 64 * 1024
	CryptoThreads = 4
	CryptoKeyLen  = 32
	CryptoSaltLen = CryptoKeyLen
)

func getEnv(key, defval string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defval
}

func makeKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, CryptoSaltLen)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}
	return argon2.IDKey(
		password,
		salt,
		CryptoTime,
		CryptoMemory,
		CryptoThreads,
		CryptoKeyLen,
	), salt, nil
}

func encrypt(password, data []byte) ([]byte, error) {
	key, salt, err := makeKey(password, nil)
	if err != nil {
		return nil, err
	}
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
	res := gcm.Seal(nonce, nonce, data, nil)
	res = append(res, salt...)
	return res, nil
}

func decrypt(password, data []byte) ([]byte, error) {
	saltPos := len(data) - CryptoSaltLen
	salt, data := data[saltPos:], data[:saltPos]
	key, _, err := makeKey(password, salt)
	if err != nil {
		return nil, err
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
