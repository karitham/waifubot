package rest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type eTagResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *eTagResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *eTagResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *eTagResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *eTagResponseWriter) ETag() string {
	if w.statusCode >= 400 {
		return ""
	}
	hash := sha256.Sum256(w.body.Bytes())
	return `"` + hex.EncodeToString(hash[:8]) + `"`
}

// ETagMiddleware adds ETag support to GET responses.
// It computes a hash of the response body and includes it in the ETag header.
// When a client sends If-None-Match matching the current ETag, returns 304 Not Modified.
// Non-GET requests pass through unchanged.
func ETagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		buf := &bytes.Buffer{}
		wrapped := &eTagResponseWriter{
			ResponseWriter: w,
			body:           buf,
			statusCode:     200,
		}

		next.ServeHTTP(wrapped, r)

		etag := wrapped.ETag()

		if etag != "" {
			w.Header().Set("ETag", etag)
			w.Header().Set("Cache-Control", "private, max-age=0, must-revalidate")

			if r.Header.Get("If-None-Match") == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.WriteHeader(wrapped.statusCode)
		io.Copy(w, buf)
	})
}
