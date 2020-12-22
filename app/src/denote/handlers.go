package denote

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxBodyBytes = 1 << 15

	MinDurationLimit     = 1 * time.Minute
	DefaultDurationLimit = 24 * time.Hour
)

func getHandler(c *Config, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	key := query.Get("key")
	if len(key) != 1<<5 {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	uid, err := uuid.Parse(key)
	if err != nil {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	if uid.Version() != 1<<2 {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	data, err := db.get(uid)
	if err != nil {
		writeErrorStatus(w, http.StatusNotFound)
		return
	}
	password := query.Get("password")
	if password == "" {
		password = c.Password
	}
	value, err := decrypt([]byte(password), data)
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
	res := c.Origin + "?key=" + strings.ReplaceAll(uid.String(), "-", "")
	if !isEmptyPassword {
		res += "&password=" + password
	}
	w.Write([]byte(res))
}

func makeRootHandleFunc(c *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getHandler(c, w, r)
		} else if r.Method == http.MethodPost {
			setHandler(c, w, r)
		} else {
			writeErrorStatus(w, http.StatusNotFound)
		}
	}
}

func writeErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
