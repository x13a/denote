package healthcheck

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/x13a/denote/api/db"
)

const (
	envPrefix  = "HEALTHCHECK_"
	EnvPath    = envPrefix + "PATH"
	EnvEnabled = envPrefix + "ENABLED"

	DefaultPath = "/ping"
)

func AddHandler(m *chi.Mux) {
	enabled, err := strconv.ParseBool(os.Getenv(EnvEnabled))
	if err != nil || !enabled {
		return
	}
	path := os.Getenv(EnvPath)
	if path == "" {
		path = DefaultPath
	}
	m.Get(path, func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, "OK")
	})
}
