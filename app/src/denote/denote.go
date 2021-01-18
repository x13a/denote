package denote

import (
	"context"
	"net/http"
	"time"

	"bitbucket.org/x31a/denote/app/src/denote/api"
	"bitbucket.org/x31a/denote/app/src/denote/config"
	"bitbucket.org/x31a/denote/app/src/denote/middleware"
	"bitbucket.org/x31a/denote/app/src/denote/static"
	"bitbucket.org/x31a/denote/app/src/healthcheck"
)

const (
	Version = "0.2.4"

	cleanerInterval = 24 * time.Hour
	maxHeaderBytes  = 1 << 10
)

func addHandlers(m *http.ServeMux, c *config.Config) {
	if c.EnableStatic {
		m.HandleFunc(c.StaticPath, static.MakeStaticHandlerFunc(c))
		m.HandleFunc(c.Path, static.MakeRootHandlerFunc(c))
	} else {
		m.HandleFunc(c.Path, api.MakeHandlerFunc(c))
	}
	healthcheck.AddHandler(m)
}

func runCleaners(
	ctx context.Context,
	c *config.Config,
	stopChan chan struct{},
) {
	n := 1
	stopChan1 := make(chan struct{})
	go api.DB.Cleaner(ctx, cleanerInterval, stopChan1)
	api.SetLimiter.SetLimit(c.IPLimit)
	if api.SetLimiter.IsActive() {
		n++
		go api.SetLimiter.Cleaner(cleanerInterval, stopChan1)
	}
	api.DeleteLimiter.SetLimit(c.IPLimit)
	if api.DeleteLimiter.IsActive() {
		n++
		go api.DeleteLimiter.Cleaner(cleanerInterval, stopChan1)
	}
	<-stopChan
	for i := 0; i < n; i++ {
		stopChan1 <- struct{}{}
		<-stopChan1
	}
	close(stopChan)
}

func Run(ctx context.Context, c config.Config) (err error) {
	if err = api.DB.Open(c.DSN); err != nil {
		return
	}
	defer func() {
		if err1 := api.DB.Close(); err == nil || err == http.ErrServerClosed {
			err = err1
		}
	}()
	if err = api.DB.Create(ctx); err != nil {
		return
	}
	stopChan := make(chan struct{})
	go runCleaners(ctx, &c, stopChan)
	defer func() {
		stopChan <- struct{}{}
		<-stopChan
	}()
	mux := http.NewServeMux()
	addHandlers(mux, &c)
	handlerTimeout := c.HandlerTimeout.Unwrap()
	srv := &http.Server{
		Addr:           c.Addr,
		ReadTimeout:    c.ReadTimeout.Unwrap(),
		WriteTimeout:   c.WriteTimeout.Unwrap(),
		IdleTimeout:    c.IdleTimeout.Unwrap(),
		MaxHeaderBytes: maxHeaderBytes,
		Handler: http.TimeoutHandler(
			middleware.Security(mux),
			handlerTimeout,
			"",
		),
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
