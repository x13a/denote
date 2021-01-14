package middleware

import "net/http"

func Security(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Cache-Control", "no-store")
		header.Set("X-Robots-Tag", "noindex, nofollow")

		header.Set("Cross-Origin-Resource-Policy", "same-origin")
		header.Set(
			"Content-Security-Policy",
			"default-src 'none'; "+
				"object-src 'none'; "+
				"base-uri 'none'; "+
				"script-src 'none'; "+
				"style-src 'self'",
		)
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("Strict-Transport-Security", "max-age=63072000")
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("X-XSS-Protection", "1; mode=block")

		header.Set(
			"Feature-Policy",
			"camera 'none'; "+
				"display-capture 'none'; "+
				"document-domain 'none'; "+
				"geolocation 'none'; "+
				"microphone 'none'; "+
				"payment 'none'; "+
				"usb 'none'",
		)
		h.ServeHTTP(w, r)
	})
}
