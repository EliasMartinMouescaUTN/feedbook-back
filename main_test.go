package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	feedbook "github.com/feedbook/back/internal/feedbook"
	feedbookhttp "github.com/feedbook/back/internal/feedbook/http"
)

func TestHandleLoginSuccess(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"username":"demo","password":"demo","secure_login":true}`),
	)
	recorder := httptest.NewRecorder()

	handleLogin(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response loginResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Token == "" {
		t.Fatal("expected token in response")
	}
}

func TestHandleLoginInvalidCredentials(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"username":"demo","password":"other","secure_login":false}`),
	)
	recorder := httptest.NewRecorder()

	handleLogin(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestResolveAddrDefaultsToLocalhost(t *testing.T) {
	if got := resolveAddr(""); got != defaultAddr {
		t.Fatalf("expected default addr %q, got %q", defaultAddr, got)
	}
}

func TestResolveAddrTrimsConfiguredValue(t *testing.T) {
	if got := resolveAddr(" 0.0.0.0:9090 "); got != "0.0.0.0:9090" {
		t.Fatalf("expected trimmed addr, got %q", got)
	}
}

func TestAPIEndpointsAreMountedUnderAPIPrefix(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", feedbookhttp.NewRouter(feedbook.NewService(feedbook.NewStore()))))

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for mounted api route, got %d", rec.Code)
	}
}
