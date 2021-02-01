package utils

import (
	"crypto/rand"
	"os"
)

func Getenv(key, defval string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defval
}

func RandRead(n int) ([]byte, error) {
	res := make([]byte, n)
	if _, err := rand.Read(res); err != nil {
		return nil, err
	}
	return res, nil
}
