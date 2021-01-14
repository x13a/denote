package static

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
)

func serveTemplate(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	data interface{},
) {
	if name == "" {
		name = "index.html"
	}
	value, err := Cache.From(filepath.Join("static", name))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := template.Must(
		template.New(name).Parse(string(value.Content)),
	).Execute(w, data); err != nil {
		http.NotFound(w, r)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, name string) {
	value, err := Cache.From(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, name, value.Time, bytes.NewReader(value.Content))
}
