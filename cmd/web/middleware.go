package main

import (
	"fmt"
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/rs/zerolog/log"
)

type StatusCodeRecorder struct {
	StatusCode int
	http.ResponseWriter
}

func (scr *StatusCodeRecorder) WriteHeader(status int) {
	scr.StatusCode = status
	scr.ResponseWriter.WriteHeader(status)
}

var secureHeaders = map[string]string{
	"Cache-Control":           "no-store",
	"Content-Security-Policy": "frame-ancestors 'none'",
	"X-Content-Type-Options":  "nosniff",
	"X-Frame-Options":         "DENY",
}

func (app *application) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("method", r.Method).Str("requestURI", r.RequestURI).Str("remote_addr", r.RemoteAddr).Send()
		next.ServeHTTP(w, r)
	})
}

func (app *application) RequestCounterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &StatusCodeRecorder{ResponseWriter: w, StatusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)
		s := fmt.Sprintf(`requests_total{method="%s", path=%q, status="%d"}`, r.Method, r.URL.Path, recorder.StatusCode)
		metrics.GetOrCreateCounter(s).Inc()
	})
}

func (app *application) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Add("Connection", "close")
				app.serverError(w, fmt.Errorf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for headerKey, headerValue := range secureHeaders {
			w.Header().Set(headerKey, headerValue)
		}
		next.ServeHTTP(w, r)
	})
}
