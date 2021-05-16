package api

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"

	"github.com/x13a/denote/api/crypt"
	"github.com/x13a/denote/api/db"
	"github.com/x13a/denote/config"
)

const (
	defaultDurationLimit = 24 * time.Hour
	minDurationLimit     = 1 * time.Minute
	maxDurationLimit     = 7 * defaultDurationLimit

	keyLen      = 1 << 4
	passwordLen = 1 << 4
	totalLen    = keyLen + passwordLen
)

var (
	ErrInvalidKeyLen     = errors.New("invalid key len")
	ErrInvalidKeyVersion = errors.New("invalid key version")
	ErrEmptyValue        = errors.New("empty value")
)

func Get(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	value, err := base64.RawURLEncoding.DecodeString(chi.URLParam(r, "key"))
	if err != nil {
		return nil, err
	} else if len(value) != totalLen {
		return nil, ErrInvalidKeyLen
	}
	uid, password := value[:keyLen], value[keyLen:]
	key, err := uuid.FromBytes(uid)
	if err != nil {
		return nil, err
	} else if key.Version() != 4 {
		return nil, ErrInvalidKeyVersion
	}
	data, err := db.Get(r.Context(), key)
	if err != nil {
		return nil, err
	}
	value, err = crypt.DecryptGCM(password, data)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func Set(w http.ResponseWriter, r *http.Request) (string, string, error) {
	if config.MaxBodyBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, config.MaxBodyBytes)
	}
	if err := r.ParseForm(); err != nil {
		return "", "", err
	}
	value := r.PostFormValue("value")
	if value == "" {
		return "", "", ErrEmptyValue
	}
	password := make([]byte, passwordLen)
	if _, err := rand.Read(password); err != nil {
		return "", "", err
	}
	viewLimit, err := strconv.Atoi(r.PostFormValue("view_limit"))
	if err != nil || viewLimit < 1 {
		viewLimit = 1
	}
	durationLimit, err := time.ParseDuration(r.PostFormValue("duration_limit"))
	if err != nil {
		durationLimit = defaultDurationLimit
	} else if durationLimit < minDurationLimit {
		durationLimit = minDurationLimit
	} else if durationLimit > maxDurationLimit {
		durationLimit = maxDurationLimit
	}
	data, err := crypt.EncryptGCM(password, []byte(value))
	if err != nil {
		return "", "", err
	}
	key, rmKey, err := db.Set(r.Context(), data, viewLimit, durationLimit)
	if err != nil {
		return "", "", err
	}
	buffer := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.RawURLEncoding, buffer)
	encoder.Write(key[:])
	encoder.Write(password)
	encoder.Close()
	var path string
	if r.URL.Path == "/api" && r.URL.Query().Get("static") != "1" {
		path = "/api"
	}
	return config.URL + path + "/get/" + buffer.String(),
		config.URL + path + "/rm/" + rmKey.String(),
		nil
}

func Delete(w http.ResponseWriter, r *http.Request) error {
	uid, err := uuid.Parse(chi.URLParam(r, "key"))
	if err != nil {
		return err
	} else if uid.Version() != 4 {
		return ErrInvalidKeyVersion
	}
	return db.Delete(r.Context(), uid)
}
