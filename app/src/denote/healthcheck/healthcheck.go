package healthcheck

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/x13a/denote/api/db"
	"github.com/x13a/denote/utils"
)

const (
	envPrefix = "HEALTHCHECK_"
	EnvPath   = envPrefix + "PATH"
	EnvEnable = envPrefix + "ENABLE"

	DefaultPath = "/ping"
)

func AddHandler(m *chi.Mux) {
	enable, err := strconv.ParseBool(os.Getenv(EnvEnable))
	if err != nil || !enable {
		return
	}
	m.Get(
		utils.Getenv(EnvPath, DefaultPath),
		func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			io.WriteString(w, "OK")
		},
	)
}
