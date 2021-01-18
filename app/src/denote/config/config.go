package config

import (
	"net/url"
	"os"
	"strconv"
	"time"
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
	EnvIPLimit      = "IP_LIMIT"
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

type Duration time.Duration

func (d *Duration) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

func (d *Duration) SetDefault(s string, defval time.Duration) {
	if err := d.Set(s); err != nil {
		*d = Duration(defval)
	}
}

func (d Duration) Unwrap() time.Duration {
	return time.Duration(d)
}

type ConfigError struct {
	Key string
	Err error
}

func (e *ConfigError) Error() string {
	return e.Key + " invalid"
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

type Config struct {
	Addr           string
	CertFile       string
	KeyFile        string
	ReadTimeout    Duration
	WriteTimeout   Duration
	IdleTimeout    Duration
	HandlerTimeout Duration

	MaxBodyBytes int64
	EnableStatic bool
	IPLimit      int
	DSN          string

	URL        string
	Path       string
	StaticPath string
}

func (c *Config) FromEnv() error {
	uri, err := url.ParseRequestURI(os.Getenv(EnvURL))
	if err != nil {
		return &ConfigError{EnvURL, err}
	}
	if (uri.Scheme != "http" && uri.Scheme != "https") ||
		uri.Hostname() == "" {

		return &ConfigError{EnvURL, nil}
	}

	c.Path = uri.Path
	if c.Path == "" {
		c.Path = "/"
	} else if c.Path[len(c.Path)-1] != '/' {
		c.Path += "/"
	}
	c.URL = uri.Scheme + "://" + uri.Host + c.Path

	c.Addr = getEnv(EnvAddr, DefaultAddr)
	c.CertFile = os.Getenv(EnvCertFile)
	c.KeyFile = os.Getenv(EnvKeyFile)
	c.ReadTimeout.SetDefault(EnvReadTimeout, DefaultReadTimeout)
	c.WriteTimeout.SetDefault(EnvWriteTimeout, DefaultWriteTimeout)
	c.IdleTimeout.SetDefault(EnvIdleTimeout, DefaultIdleTimeout)
	c.HandlerTimeout.SetDefault(EnvHandlerTimeout, DefaultHandlerTimeout)

	maxBodyBytes, err := strconv.ParseInt(os.Getenv(EnvMaxBodyBytes), 10, 64)
	if err != nil {
		maxBodyBytes = DefaultMaxBodyBytes
	}
	c.MaxBodyBytes = maxBodyBytes

	c.EnableStatic, _ = strconv.ParseBool(os.Getenv(EnvEnableStatic))
	if c.EnableStatic {
		c.StaticPath = c.Path + "static/"
	}

	c.IPLimit, _ = strconv.Atoi(os.Getenv(EnvIPLimit))
	c.DSN = getEnv(EnvDSN, DefaultDSN)
	os.Unsetenv(EnvDSN)
	return nil
}
