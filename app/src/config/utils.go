package config

import "os"

func getEnv(key, defval string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defval
}
