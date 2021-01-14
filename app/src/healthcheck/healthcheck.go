package healthcheck

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"bitbucket.org/x31a/denote/app/src/denote/api"
)

const (
	envPrefix  = "HEALTHCHECK_"
	EnvPath    = envPrefix + "PATH"
	EnvEnabled = envPrefix + "ENABLED"

	DefaultPath = "/ping/"
)

func AddHandler(m *http.ServeMux) {
	enabled, err := strconv.ParseBool(os.Getenv(EnvEnabled))
	if err != nil || !enabled {
		return
	}
	path := os.Getenv(EnvPath)
	if path == "" {
		path = DefaultPath
	} else if path[len(path)-1] != '/' {
		path += "/"
	}
	m.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if err := api.DB.Ping(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, "OK")
	})
}
