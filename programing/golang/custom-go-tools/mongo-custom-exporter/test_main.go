package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMain_StartHTTPServer(t *testing.T) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	http.DefaultServeMux = http.NewServeMux()
	go main()

	time.Sleep(1 * time.Second) // Allow server to start

	http.DefaultServeMux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected HTTP status 200, got %v", w.Code)
	}
}
