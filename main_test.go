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

	handleLogin(newMemoryAccountStore())(recorder, request)

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

	handleLogin(newMemoryAccountStore())(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestHandleRegisterThenLogin(t *testing.T) {
	accounts := newMemoryAccountStore()
	registerRequest := httptest.NewRequest(
		http.MethodPost,
		"/register",
		strings.NewReader(`{"username":"reader@example.com","password":"secret"}`),
	)
	registerRecorder := httptest.NewRecorder()

	handleRegister(accounts)(registerRecorder, registerRequest)

	if registerRecorder.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d", registerRecorder.Code)
	}

	loginRequest := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"username":"reader@example.com","password":"secret"}`),
	)
	loginRecorder := httptest.NewRecorder()

	handleLogin(accounts)(loginRecorder, loginRequest)

	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", loginRecorder.Code)
	}
}

func TestHandleRegisterRejectsDuplicateAccount(t *testing.T) {
	accounts := newMemoryAccountStore()
	body := `{"username":"reader@example.com","password":"secret"}`

	firstRequest := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	firstRecorder := httptest.NewRecorder()
	handleRegister(accounts)(firstRecorder, firstRequest)
	if firstRecorder.Code != http.StatusCreated {
		t.Fatalf("expected first register status 201, got %d", firstRecorder.Code)
	}

	secondRequest := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	secondRecorder := httptest.NewRecorder()
	handleRegister(accounts)(secondRecorder, secondRequest)

	if secondRecorder.Code != http.StatusConflict {
		t.Fatalf("expected duplicate register status 409, got %d", secondRecorder.Code)
	}
}

func TestSQLiteAccountStorePersistsRegisteredAccount(t *testing.T) {
	dbPath := t.TempDir() + "/feedbook.db"
	accounts, err := feedbook.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create sqlite store: %v", err)
	}
	registerRequest := httptest.NewRequest(
		http.MethodPost,
		"/register",
		strings.NewReader(`{"username":"sqlite@example.com","password":"secret"}`),
	)
	registerRecorder := httptest.NewRecorder()

	handleRegister(accounts)(registerRecorder, registerRequest)

	if registerRecorder.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d", registerRecorder.Code)
	}

	reopenedAccounts, err := feedbook.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("reopen sqlite store: %v", err)
	}
	loginRequest := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"username":"sqlite@example.com","password":"secret"}`),
	)
	loginRecorder := httptest.NewRecorder()

	handleLogin(reopenedAccounts)(loginRecorder, loginRequest)

	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected login status 200 after reopening db, got %d", loginRecorder.Code)
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
	mux.Handle("/api/", http.StripPrefix("/api", feedbookhttp.NewRouter(feedbook.NewService(feedbook.NewMemoryStore()))))

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for mounted api route, got %d", rec.Code)
	}
}
