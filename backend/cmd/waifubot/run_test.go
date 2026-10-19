package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLimitDiscordBodyRejectsOversizedBody(t *testing.T) {
	var readErr error
	handler := limitDiscordBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	body := bytes.NewReader(make([]byte, maxDiscordInteractionBytes+1))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if readErr == nil {
		t.Fatal("expected the body read to be bounded")
	}
}

func TestNewHTTPServerSetsTransportDeadlines(t *testing.T) {
	srv := newHTTPServer(":0", http.NotFoundHandler())

	if srv.ReadHeaderTimeout == 0 {
		t.Error("ReadHeaderTimeout must be set")
	}
	if srv.ReadTimeout == 0 {
		t.Error("ReadTimeout must be set")
	}
	if srv.WriteTimeout == 0 {
		t.Error("WriteTimeout must be set")
	}
	if srv.IdleTimeout == 0 {
		t.Error("IdleTimeout must be set")
	}
	if srv.MaxHeaderBytes == 0 {
		t.Error("MaxHeaderBytes must be set")
	}
}
