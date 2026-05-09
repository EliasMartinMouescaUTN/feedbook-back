package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleLoginSuccess(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"username":"demo","password":"demo","easy_login":true}`),
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
		strings.NewReader(`{"username":"demo","password":"other","easy_login":false}`),
	)
	recorder := httptest.NewRecorder()

	handleLogin(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}
