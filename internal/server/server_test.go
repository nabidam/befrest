package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestNewServesSPA(t *testing.T) {
	handler, err := New(fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("<h1>Befrest</h1>")},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), "Befrest") {
		t.Fatalf("GET / body = %q, want SPA HTML", response.Body.String())
	}
}
