package rest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestETagMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		handler       func(w http.ResponseWriter, r *http.Request)
		ifNoneMatch   string
		wantStatus    int
		wantETag      bool
		wantBody      bool
		wantBodyBytes []byte
	}{
		{
			name:   "non-GET passes through without ETag",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
			wantETag:   false,
			wantBody:   false,
		},
		{
			name:   "GET 200 includes ETag and Cache-Control",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"test":"data"}`))
			},
			wantStatus:    http.StatusOK,
			wantETag:      true,
			wantBody:      true,
			wantBodyBytes: []byte(`{"test":"data"}`),
		},
		{
			name:   "GET with matching If-None-Match returns 304",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"test":"data"}`))
			},
			ifNoneMatch: `"` + hash(`{"test":"data"}`) + `"`,
			wantStatus:  http.StatusNotModified,
			wantETag:    true,
			wantBody:    false,
		},
		{
			name:   "GET with non-matching If-None-Match returns 200 with body",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"test":"data"}`))
			},
			ifNoneMatch:   `"wrong-etag"`,
			wantStatus:    http.StatusOK,
			wantETag:      true,
			wantBody:      true,
			wantBodyBytes: []byte(`{"test":"data"}`),
		},
		{
			name:   "error responses do not get ETag",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantETag:   false,
			wantBody:   true,
		},
		{
			name:   "same body produces same ETag",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"consistent":"data"}`))
			},
			wantStatus: http.StatusOK,
			wantETag:   true,
			wantBody:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(tt.handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			if tt.ifNoneMatch != "" {
				req.Header.Set("If-None-Match", tt.ifNoneMatch)
			}

			ETagMiddleware(handler).ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			etag := w.Header().Get("ETag")
			if tt.wantETag && etag == "" {
				t.Error("expected ETag header to be set")
			}
			if !tt.wantETag && etag != "" {
				t.Errorf("expected no ETag, got %s", etag)
			}

			if tt.wantETag && etag != "" {
				if etag[0] != '"' || etag[len(etag)-1] != '"' {
					t.Errorf("ETag should be quoted, got %s", etag)
				}
				cc := w.Header().Get("Cache-Control")
				if cc == "" {
					t.Error("expected Cache-Control header to be set")
				}
			}

			if tt.wantBody {
				respBody, _ := io.ReadAll(w.Body)
				if tt.wantBodyBytes != nil && !bytes.Equal(respBody, tt.wantBodyBytes) {
					t.Errorf("body = %s, want %s", respBody, tt.wantBodyBytes)
				}
			}
		})
	}
}

func hash(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:8])
}
