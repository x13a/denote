package denote

import (
	"os"
	"strconv"
	"time"
)

const (
	MinPasswordLen = 1 << 4

	EnvAddr           = "ADDR"
	EnvCertFile       = "CERT_FILE"
	EnvKeyFile        = "KEY_FILE"
	EnvReadTimeout    = "READ_TIMEOUT"
	EnvWriteTimeout   = "WRITE_TIMEOUT"
	EnvIdleTimeout    = "IDLE_TIMEOUT"
	EnvHandlerTimeout = "HANDLER_TIMEOUT"

	EnvDsn            = "DSN"
	EnvRunCleanerTask = "RUN_CLEANER_TASK"
	EnvPath           = "ROOT_PATH"
	EnvURLOrigin      = "URL_ORIGIN"
	EnvPassword       = "PASSWORD"

	DefaultAddr           = "127.0.0.1:8000"
	DefaultReadTimeout    = 1 << 2 * time.Second
	DefaultWriteTimeout   = DefaultReadTimeout
	DefaultIdleTimeout    = 1 << 5 * time.Second
	DefaultHandlerTimeout = DefaultIdleTimeout

	DefaultDsn            = "file:denote.db?mode=rwc"
	DefaultRunCleanerTask = true
	DefaultPath           = "/"
)

type (
	Duration time.Duration
	Bool     bool
)

func (d *Duration) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

func (d *Duration) SetDefault(s string, defval time.Duration) error {
	if err := d.Set(s); err != nil {
		*d = Duration(defval)
	}
	return nil
}

func (d Duration) Unwrap() time.Duration {
	return time.Duration(d)
}

func (b *Bool) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*b = Bool(v)
	return nil
}

func (b *Bool) SetDefault(s string, defval bool) error {
	if err := b.Set(s); err != nil {
		*b = Bool(defval)
	}
	return nil
}

func (b Bool) Unwrap() bool {
	return bool(b)
}

type ConfigError struct {
	key string
}

func (e *ConfigError) Error() string {
	return e.key + " invalid"
}

type Config struct {
	Addr           string
	CertFile       string
	KeyFile        string
	ReadTimeout    Duration
	WriteTimeout   Duration
	IdleTimeout    Duration
	HandlerTimeout Duration

	Dsn            string
	RunCleanerTask Bool
	Path           string
	URLOrigin      string
	Password       string
}

func (c *Config) Parse() error {
	c.Addr = getEnv(EnvAddr, DefaultAddr)
	c.CertFile = os.Getenv(EnvCertFile)
	c.KeyFile = os.Getenv(EnvKeyFile)
	c.ReadTimeout.SetDefault(EnvReadTimeout, DefaultReadTimeout)
	c.WriteTimeout.SetDefault(EnvWriteTimeout, DefaultWriteTimeout)
	c.IdleTimeout.SetDefault(EnvIdleTimeout, DefaultIdleTimeout)
	c.HandlerTimeout.SetDefault(EnvHandlerTimeout, DefaultHandlerTimeout)

	c.Dsn = getEnv(EnvDsn, DefaultDsn)
	os.Unsetenv(EnvDsn)
	c.RunCleanerTask.SetDefault(EnvRunCleanerTask, DefaultRunCleanerTask)
	c.Path = getEnv(EnvPath, DefaultPath)
	c.URLOrigin = os.Getenv(EnvURLOrigin)
	if c.URLOrigin == "" {
		return &ConfigError{EnvURLOrigin}
	}
	c.Password = os.Getenv(EnvPassword)
	if len(c.Password) < MinPasswordLen {
		return &ConfigError{EnvPassword}
	}
	os.Unsetenv(EnvPassword)
	return nil
}
