package middleware

import (
	"net/http"

	"github.com/x13a/denote/config"
)

const (
	featurePolicy = "autoplay 'none'; " +
		"camera 'none'; " +
		"display-capture 'none'; " +
		"document-domain 'none'; " +
		"geolocation 'none'; " +
		"microphone 'none'; " +
		"payment 'none'; " +
		"usb 'none'"
	csp = "default-src 'none'; " +
		"object-src 'none'; " +
		"base-uri 'none'; " +
		"style-src 'self'; " +
		"script-src "
)

func MakeSecurity() func(h http.Handler) http.Handler {
	csp := csp
	if config.EnableJS {
		csp += "'self'"
	} else {
		csp += "'none'"
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			header.Set("Cache-Control", "no-store")
			header.Set("X-Robots-Tag", "noindex, nofollow")

			header.Set("Cross-Origin-Resource-Policy", "same-origin")
			header.Set("Content-Security-Policy", csp)

			header.Set("Referrer-Policy", "no-referrer")
			header.Set("Strict-Transport-Security", "max-age=63072000")
			header.Set("X-Content-Type-Options", "nosniff")
			header.Set("X-Frame-Options", "DENY")
			header.Set("X-XSS-Protection", "1; mode=block")

			header.Set("Feature-Policy", featurePolicy)
			h.ServeHTTP(w, r)
		})
	}
}
