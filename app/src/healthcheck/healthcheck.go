package healthcheck

import (
	"net/http"
	"os"
	"strconv"
)

const (
	envPrefix  = "HEALTHCHECK_"
	EnvPath    = envPrefix + "PATH"
	EnvEnabled = envPrefix + "ENABLED"

	DefaultPath = "/ping"
)

func AddHandler(m *http.ServeMux) {
	enabled, err := strconv.ParseBool(os.Getenv(EnvEnabled))
	if err != nil || !enabled {
		return
	}
	path := os.Getenv(EnvPath)
	if path == "" {
		path = DefaultPath
	}
	m.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}
