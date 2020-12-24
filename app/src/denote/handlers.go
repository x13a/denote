package denote

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	MaxBodyBytes = 1 << 15

	MinDurationLimit     = 1 * time.Minute
	DefaultDurationLimit = 24 * time.Hour

	keyLen = 1 << 4
)

func getHandler(c *Config, w http.ResponseWriter, r *http.Request) {
	q, err := base64.RawURLEncoding.DecodeString(r.URL.Query().Get("q"))
	if err != nil || len(q) < keyLen {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	key, password := q[:keyLen], q[keyLen:]
	uid, err := uuid.FromBytes(key)
	if err != nil || uid.Version() != 4 {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	data, err := db.get(uid)
	if err != nil {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	if len(password) == 0 {
		password = []byte(c.Password)
	}
	value, err := decrypt(password, data)
	if err != nil {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	w.Write(value)
}

func setHandler(c *Config, w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	if err := r.ParseForm(); err != nil {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	value := r.PostFormValue("value")
	if value == "" {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	password := r.PostFormValue("password")
	isEmptyPassword := password == ""
	if isEmptyPassword {
		password = c.Password
	}
	viewLimit, err := strconv.Atoi(r.PostFormValue("view_limit"))
	if err != nil || viewLimit < 1 {
		viewLimit = 1
	}
	durationLimit, err := time.ParseDuration(r.PostFormValue("duration_limit"))
	if err != nil {
		durationLimit = DefaultDurationLimit
	} else if durationLimit < MinDurationLimit {
		durationLimit = MinDurationLimit
	}
	data, err := encrypt([]byte(password), []byte(value))
	if err != nil {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	uid, err := db.set(data, viewLimit, durationLimit)
	if err != nil {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	w.Write([]byte(c.URLOrigin + "?q="))
	encoder := base64.NewEncoder(base64.RawURLEncoding, w)
	encoder.Write(uid[:])
	if !isEmptyPassword {
		encoder.Write([]byte(password))
	}
	encoder.Close()
}

func makeRootHandleFunc(c *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getHandler(c, w, r)
		case http.MethodPost:
			setHandler(c, w, r)
		default:
			writeErrorStatus(w, http.StatusNotFound)
		}
	}
}

func writeErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
