package denote

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/x13a/denote/api"
	"github.com/x13a/denote/api/db"
	"github.com/x13a/denote/config"
	"github.com/x13a/denote/healthcheck"
	"github.com/x13a/denote/middleware"
	"github.com/x13a/denote/static"
)

const (
	Version = "0.2.6"

	cleanerInterval = 24 * time.Hour
	maxHeaderBytes  = 1 << 10
)

func addHandlers(m *chi.Mux) {
	m.Use(middleware.MakeSecurity())
	apiPath := "/"
	if config.EnableStatic {
		apiPath += "api"
		m.Get("/", static.Index)
		m.Get("/{name}", static.Index)
		m.Post("/", static.Set)
		m.Get("/get/{key}", static.Get)
		m.Get("/rm/{key}", static.Delete)
		m.Get("/static/{name}", static.Static)
	}
	m.Route(apiPath, func(r chi.Router) {
		r.Post("/", api.SetHandler)
		r.Get("/get/{key}", api.GetHandler)
		r.Get("/rm/{key}", api.DeleteHandler)
	})
	healthcheck.AddHandler(m)
}

func Run(ctx context.Context) (err error) {
	if err = db.Init(ctx); err != nil {
		return
	}
	defer func() {
		if err1 := db.Close(); err == nil || err == http.ErrServerClosed {
			err = err1
		}
	}()
	stopChan := make(chan struct{})
	go db.Cleaner(ctx, cleanerInterval, stopChan)
	defer func() {
		stopChan <- struct{}{}
		<-stopChan
	}()
	router := chi.NewRouter()
	addHandlers(router)
	handlerTimeout := config.HandlerTimeout.Unwrap()
	srv := &http.Server{
		Addr:           config.Addr,
		ReadTimeout:    config.ReadTimeout.Unwrap(),
		WriteTimeout:   config.WriteTimeout.Unwrap(),
		IdleTimeout:    config.IdleTimeout.Unwrap(),
		MaxHeaderBytes: maxHeaderBytes,
		Handler:        http.TimeoutHandler(router, handlerTimeout, ""),
	}
	errChan := make(chan error, 1)
	go func() {
		if config.CertFile != "" && config.KeyFile != "" {
			errChan <- srv.ListenAndServeTLS(config.CertFile, config.KeyFile)
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
