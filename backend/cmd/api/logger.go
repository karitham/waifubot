package main

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func loggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				t2 := time.Now()

				// Recover and record stack traces in case of a panic
				if rec := recover(); rec != nil {
					logger.Error("log system error",
						"type", "error",
						"recover_info", rec,
						"debug_stack", debug.Stack())
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				// log end request
				logger.Info("incoming_request",
					"type", "access",
					"remote_ip", r.RemoteAddr,
					"url", r.URL.Path,
					"proto", r.Proto,
					"method", r.Method,
					"user_agent", r.Header.Get("User-Agent"),
					"status", ww.Status(),
					"latency_ms", float64(t2.Sub(t1).Nanoseconds())/1000000.0,
					"bytes_in", r.Header.Get("Content-Length"),
					"bytes_out", ww.BytesWritten())
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
