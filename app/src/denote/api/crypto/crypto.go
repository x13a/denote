package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32

	SaltLen     = argonKeyLen
	PasswordLen = 1 << 4
)

func RandRead(n int) ([]byte, error) {
	res := make([]byte, n)
	if _, err := rand.Read(res); err != nil {
		return nil, err
	}
	return res, nil
}

func makeKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		var err error
		if salt, err = RandRead(SaltLen); err != nil {
			return nil, nil, err
		}
	}
	return argon2.IDKey(
		password,
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	), salt, nil
}

func Encrypt(password, data []byte) ([]byte, error) {
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

func Decrypt(password, data []byte) ([]byte, error) {
	saltPos := len(data) - SaltLen
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
