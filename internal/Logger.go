package internal

import (
	"log"
	"net/http"
	"time"
)

// Logger returns a simple Logger that wraps a http.handler and logs any incoming request
func Logger(inner appHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			inner.name,
			time.Since(start),
		)
	})
}