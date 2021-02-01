package config

import (
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/x13a/denote/utils"
)

const (
	EnvAddr           = "ADDR"
	EnvCertFile       = "CERT_FILE"
	EnvKeyFile        = "KEY_FILE"
	EnvReadTimeout    = "READ_TIMEOUT"
	EnvWriteTimeout   = "WRITE_TIMEOUT"
	EnvIdleTimeout    = "IDLE_TIMEOUT"
	EnvHandlerTimeout = "HANDLER_TIMEOUT"

	EnvMaxBodyBytes = "MAX_BODY_BYTES"
	EnvEnableStatic = "ENABLE_STATIC"
	EnvEnableJS     = "ENABLE_JS"
	EnvDSN          = "DSN"
	EnvURL          = "URL"

	DefaultAddr           = "127.0.0.1:8000"
	DefaultReadTimeout    = 1 << 2 * time.Second
	DefaultWriteTimeout   = DefaultReadTimeout
	DefaultIdleTimeout    = 1 << 5 * time.Second
	DefaultHandlerTimeout = DefaultIdleTimeout

	DefaultMaxBodyBytes = 1 << 12
	DefaultDSN          = "file:./db/denote.db?mode=rwc"
)

var (
	Addr           string
	CertFile       string
	KeyFile        string
	ReadTimeout    duration
	WriteTimeout   duration
	IdleTimeout    duration
	HandlerTimeout duration

	MaxBodyBytes int64
	EnableStatic bool
	EnableJS     bool
	DSN          string

	URL string
)

type duration time.Duration

func (d *duration) set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = duration(v)
	return nil
}

func (d *duration) setDefault(s string, defval time.Duration) {
	if err := d.set(s); err != nil {
		*d = duration(defval)
	}
}

func (d duration) Unwrap() time.Duration {
	return time.Duration(d)
}

type Error struct {
	Key string
	Err error
}

func (e *Error) Error() string {
	return e.Key + " invalid value"
}

func (e *Error) Unwrap() error {
	return e.Err
}

func LoadEnv() error {
	uri, err := url.ParseRequestURI(os.Getenv(EnvURL))
	if err != nil {
		return &Error{EnvURL, err}
	}
	if (uri.Scheme != "http" && uri.Scheme != "https") ||
		uri.Hostname() == "" {

		return &Error{EnvURL, nil}
	}
	URL = uri.Scheme + "://" + uri.Host

	Addr = utils.Getenv(EnvAddr, DefaultAddr)
	CertFile = os.Getenv(EnvCertFile)
	KeyFile = os.Getenv(EnvKeyFile)
	ReadTimeout.setDefault(EnvReadTimeout, DefaultReadTimeout)
	WriteTimeout.setDefault(EnvWriteTimeout, DefaultWriteTimeout)
	IdleTimeout.setDefault(EnvIdleTimeout, DefaultIdleTimeout)
	HandlerTimeout.setDefault(EnvHandlerTimeout, DefaultHandlerTimeout)

	maxBodyBytes, err := strconv.ParseInt(os.Getenv(EnvMaxBodyBytes), 10, 64)
	if err != nil {
		maxBodyBytes = DefaultMaxBodyBytes
	}
	MaxBodyBytes = maxBodyBytes

	EnableStatic, _ = strconv.ParseBool(os.Getenv(EnvEnableStatic))
	EnableJS, _ = strconv.ParseBool(os.Getenv(EnvEnableJS))

	DSN = utils.Getenv(EnvDSN, DefaultDSN)
	os.Unsetenv(EnvDSN)
	return nil
}
