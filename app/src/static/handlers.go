package static

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/go-chi/chi"

	"github.com/x13a/denote/api"
	"github.com/x13a/denote/static/filecache"
)

var cache = filecache.New()

func Get(w http.ResponseWriter, r *http.Request) {
	value, err := api.Get(w, r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	serveTemplate(w, r, "", string(value))
}

func Set(w http.ResponseWriter, r *http.Request) {
	getURL, deleteURL, err := api.Set(w, r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var buffer strings.Builder
	buffer.WriteString(getURL)
	buffer.WriteString("\n")
	buffer.WriteString(deleteURL)
	serveTemplate(w, r, "", buffer.String())
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if api.Delete(w, r) != nil {
		http.NotFound(w, r)
		return
	}
	serveTemplate(w, r, "", "OK")
}

func Index(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	switch name {
	case "":
		w.Header().Set("X-Robots-Tag", "nofollow, noarchive, notranslate")
		serveTemplate(w, r, "", nil)
	case "security.txt", "robots.txt":
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		serveFile(w, r, filepath.Join("static", name))
	default:
		http.NotFound(w, r)
	}
}

func Static(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" || name != strings.Trim(name, ".") {
		http.NotFound(w, r)
		return
	}
	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
			!unicode.IsDigit(c) && !strings.ContainsRune(".-_", c) {

			http.NotFound(w, r)
			return
		}
	}
	if strings.HasSuffix(name, ".html") ||
		strings.HasSuffix(name, ".txt") {

		http.NotFound(w, r)
		return
	}
	name = filepath.Join("static", name)
	if !cache.Has(name) {
		fileInfo, err := os.Lstat(name)
		if err != nil || !fileInfo.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
	}
	w.Header().Set("Cache-Control", "no-cache, private")
	serveFile(w, r, name)
}
