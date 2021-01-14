package static

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"bitbucket.org/x31a/denote/app/src/denote/api"
	"bitbucket.org/x31a/denote/app/src/denote/config"
	"bitbucket.org/x31a/denote/app/src/denote/static/cache"
)

var Cache = cache.NewFileCache()

func get(w http.ResponseWriter, r *http.Request, c *config.Config) {
	if strings.Contains(r.URL.RawQuery, "api=1") {
		api.GetHandler(w, r, c)
		return
	}
	isDelete := strings.Contains(r.URL.RawQuery, "rm=")
	if !strings.Contains(r.URL.RawQuery, "q=") && !isDelete {
		name := r.URL.Path[len(c.Path):]
		switch name {
		case "":
			header := w.Header()
			header.Set("X-Robots-Tag", "nofollow, noarchive, notranslate")
			serveTemplate(w, r, "", nil)
		case "security.txt", "robots.txt":
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			serveFile(w, r, filepath.Join("static", name))
		default:
			http.NotFound(w, r)
		}
		return
	}
	var value []byte
	if isDelete {
		if api.Delete(w, r, c) == nil {
			value = []byte("OK")
		}
	} else {
		value, _ = api.Get(w, r, c)
	}
	if value == nil {
		http.NotFound(w, r)
		return
	}
	serveTemplate(w, r, "", string(value))
}

func set(w http.ResponseWriter, r *http.Request, c *config.Config) {
	if strings.Contains(r.URL.RawQuery, "api=1") {
		api.SetHandler(w, r, c)
		return
	}
	key, rmKey, password, _ := api.Set(w, r, c)
	if key == uuid.Nil {
		http.NotFound(w, r)
		return
	}
	buffer := &bytes.Buffer{}
	api.WriteGetURL(buffer, c, key, password)
	buffer.WriteString("\n")
	api.WriteDeleteURL(buffer, c, rmKey)
	serveTemplate(w, r, "", buffer.String())
}

func MakeRootHandlerFunc(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			get(w, r, c)
		case http.MethodPost:
			set(w, r, c)
		default:
			http.NotFound(w, r)
		}
	}
}

func MakeStaticHandlerFunc(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		name := strings.Trim(r.URL.Path[len(c.StaticPath):], ".-_")
		for _, c := range name {
			if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
				!unicode.IsDigit(c) && !strings.ContainsRune(".-_", c) {

				http.NotFound(w, r)
				return
			}
		}
		if name == "" || strings.HasSuffix(name, ".html") ||
			strings.HasSuffix(name, ".txt") {

			http.NotFound(w, r)
			return
		}
		name = filepath.Join("static", name)
		if !Cache.Has(name) {
			fileInfo, err := os.Lstat(name)
			if err != nil || !fileInfo.Mode().IsRegular() {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Cache-Control", "no-cache, private")
		serveFile(w, r, name)
	}
}
