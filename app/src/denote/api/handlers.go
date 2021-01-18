package api

import (
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"bitbucket.org/x31a/denote/app/src/denote/config"
)

func GetHandler(w http.ResponseWriter, r *http.Request, c *config.Config) {
	if strings.Contains(r.URL.RawQuery, "rm=") {
		if err := Delete(w, r, c); err != nil {
			http.NotFound(w, r)
		} else {
			io.WriteString(w, "OK")
		}
		return
	}
	res, _ := Get(w, r, c)
	if res == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write(res)
}

func SetHandler(w http.ResponseWriter, r *http.Request, c *config.Config) {
	key, rmKey, password, _ := Set(w, r, c)
	if key == uuid.Nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	WriteGetURL(w, c, key, password)
	io.WriteString(w, "\n")
	WriteDeleteURL(w, c, rmKey)
}

func MakeHandlerFunc(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetHandler(w, r, c)
		case http.MethodPost:
			SetHandler(w, r, c)
		default:
			http.NotFound(w, r)
		}
	}
}
