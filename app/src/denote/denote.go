package denote

import (
	"context"
	"net/http"

	"bitbucket.org/x31a/denote/app/src/healthcheck"
)

const (
	Version = "0.2.0"

	MaxHeaderBytes = 1 << 10
)

func addHandlers(m *http.ServeMux, c *Config) {
	m.HandleFunc(c.Path, makeRootHandleFunc(c))
	healthcheck.AddHandler(m)
}

func Run(ctx context.Context, c Config) (err error) {
	if err = db.open(c.Dsn); err != nil {
		return
	}
	defer func() {
		if err1 := db.close(); err == nil || err == http.ErrServerClosed {
			err = err1
		}
	}()
	if err = db.create(); err != nil {
		return
	}
	if c.RunCleanerTask {
		stopChan := make(chan struct{}, 1)
		go db.cleaner(stopChan)
		defer func() {
			stopChan <- struct{}{}
			<-stopChan
		}()
	}
	mux := http.NewServeMux()
	addHandlers(mux, &c)
	handlerTimeout := c.HandlerTimeout.Unwrap()
	srv := &http.Server{
		Addr:           c.Addr,
		ReadTimeout:    c.ReadTimeout.Unwrap(),
		WriteTimeout:   c.WriteTimeout.Unwrap(),
		IdleTimeout:    c.IdleTimeout.Unwrap(),
		MaxHeaderBytes: MaxHeaderBytes,
		Handler:        http.TimeoutHandler(mux, handlerTimeout, ""),
	}
	errChan := make(chan error, 1)
	go func() {
		if c.CertFile != "" && c.KeyFile != "" {
			errChan <- srv.ListenAndServeTLS(c.CertFile, c.KeyFile)
		} else {
			errChan <- srv.ListenAndServe()
		}
	}()
	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(
			context.Background(),
			handlerTimeout,
		)
		defer cancel()
		err = srv.Shutdown(ctx)
	case err = <-errChan:
	}
	return
}
