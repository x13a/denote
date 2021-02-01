package api

import (
	"io"
	"net/http"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	value, err := Get(w, r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write(value)
}

func SetHandler(w http.ResponseWriter, r *http.Request) {
	getURL, deleteURL, err := Set(w, r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	io.WriteString(w, getURL)
	io.WriteString(w, "\n")
	io.WriteString(w, deleteURL)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if err := Delete(w, r); err != nil {
		http.NotFound(w, r)
	} else {
		io.WriteString(w, "OK")
	}
}
