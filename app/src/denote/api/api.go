package api

import (
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"bitbucket.org/x31a/denote/app/src/denote/api/crypto"
	"bitbucket.org/x31a/denote/app/src/denote/api/limiter"
	"bitbucket.org/x31a/denote/app/src/denote/api/storage"
	"bitbucket.org/x31a/denote/app/src/denote/config"
)

const (
	defaultDurationLimit = 24 * time.Hour
	minDurationLimit     = 1 * time.Minute
	maxDurationLimit     = 7 * defaultDurationLimit

	keyLen = 1 << 4
)

var (
	DB            = &storage.Database{}
	SetLimiter    = limiter.NewIPLimiter(0)
	DeleteLimiter = limiter.NewIPLimiter(5)
)

func Get(
	w http.ResponseWriter,
	r *http.Request,
	c *config.Config,
) ([]byte, error) {
	q := r.URL.Query().Get("q")
	if q == "" {
		return nil, nil
	}
	value, err := base64.RawURLEncoding.DecodeString(q)
	if err != nil || len(value) != keyLen+crypto.PasswordLen {
		return nil, err
	}
	uid, password := value[:keyLen], value[keyLen:]
	key, err := uuid.FromBytes(uid)
	if err != nil || key.Version() != 4 {
		return nil, err
	}
	data, err := DB.Get(r.Context(), key)
	if err != nil {
		return nil, err
	}
	value, err = crypto.Decrypt(password, data)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func Set(
	w http.ResponseWriter,
	r *http.Request,
	c *config.Config,
) (uuid.UUID, uuid.UUID, []byte, error) {
	if !SetLimiter.Allow(r) {
		return uuid.Nil, uuid.Nil, nil, nil
	}
	if c.MaxBodyBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, c.MaxBodyBytes)
	}
	if err := r.ParseForm(); err != nil {
		return uuid.Nil, uuid.Nil, nil, err
	}
	value := r.PostFormValue("value")
	if value == "" {
		return uuid.Nil, uuid.Nil, nil, nil
	}
	password, err := crypto.RandRead(crypto.PasswordLen)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, err
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
	data, err := crypto.Encrypt(password, []byte(value))
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, err
	}
	key, rmKey, err := DB.Set(r.Context(), data, viewLimit, durationLimit)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, err
	}
	return key, rmKey, password, nil
}

func Delete(
	w http.ResponseWriter,
	r *http.Request,
	c *config.Config,
) error {
	if !DeleteLimiter.Allow(r) {
		return nil
	}
	rm := r.URL.Query().Get("rm")
	if rm == "" {
		return nil
	}
	rmKey, err := uuid.Parse(rm)
	if err != nil || rmKey.Version() != 4 {
		return err
	}
	return DB.Delete(r.Context(), rmKey)
}

func WriteGetURL(
	w io.Writer,
	c *config.Config,
	key uuid.UUID,
	password []byte,
) {
	io.WriteString(w, c.URL+"?q=")
	encoder := base64.NewEncoder(base64.RawURLEncoding, w)
	encoder.Write(key[:])
	encoder.Write(password)
	encoder.Close()
}

func WriteDeleteURL(w io.Writer, c *config.Config, rmKey uuid.UUID) {
	io.WriteString(w,
		c.URL+"?rm="+strings.ReplaceAll(rmKey.String(), "-", ""))
}
