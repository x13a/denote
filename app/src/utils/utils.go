package utils

import "os"

func Getenv(key, defval string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defval
}
