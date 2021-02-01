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
	m.Group(func(r chi.Router) {
		r.Use(middleware.MakeSecurity())
		apiPath := "/"
		if config.EnableStatic {
			apiPath += "api"
			r.Get("/", static.Index)
			r.Get("/{name}", static.Index)
			r.Post("/", static.Set)
			r.Get("/get/{key}", static.Get)
			r.Get("/rm/{key}", static.Delete)
			r.Get("/static/{name}", static.Static)
		}
		r.Route(apiPath, func(r chi.Router) {
			r.Post("/", api.SetHandler)
			r.Get("/get/{key}", api.GetHandler)
			r.Get("/rm/{key}", api.DeleteHandler)
		})
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go db.Cleaner(ctx, cleanerInterval)
	mux := chi.NewRouter()
	addHandlers(mux)
	handlerTimeout := config.HandlerTimeout.Unwrap()
	srv := &http.Server{
		Addr:           config.Addr,
		ReadTimeout:    config.ReadTimeout.Unwrap(),
		WriteTimeout:   config.WriteTimeout.Unwrap(),
		IdleTimeout:    config.IdleTimeout.Unwrap(),
		MaxHeaderBytes: maxHeaderBytes,
		Handler:        http.TimeoutHandler(mux, handlerTimeout, ""),
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
