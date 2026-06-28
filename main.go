package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	feedbook "github.com/feedbook/back/internal/feedbook"
	feedbookhttp "github.com/feedbook/back/internal/feedbook/http"
)

const (
	defaultAddr              = ":8080"
	defaultFirebaseCredsFile = "firebase-service-account.json"
	defaultFirebaseProjectID = "feedbook-9132b"
	jwtSecret                = "feedbook-local-secret"
)

type loginRequest struct {
	Username    string `json:"username"`
	User        string `json:"user"`
	Password    string `json:"password"`
	SecureLogin bool   `json:"secure_login"`
}

type loginResponse struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type accountStore interface {
	CreateAccount(username string, password string) (bool, error)
	AccountPassword(username string) (string, bool, error)
}

type memoryAccountStore struct {
	mu       sync.RWMutex
	accounts map[string]string
}

func newMemoryAccountStore() *memoryAccountStore {
	store := &memoryAccountStore{accounts: make(map[string]string)}
	store.accounts["demo"] = "demo"
	return store
}

func (s *memoryAccountStore) CreateAccount(username string, password string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accounts[username]; exists {
		return false, nil
	}
	s.accounts[username] = password
	return true, nil
}

func (s *memoryAccountStore) AccountPassword(username string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	password, exists := s.accounts[username]
	return password, exists, nil
}

func main() {
	var store feedbook.Storer
	var accounts accountStore
	switch os.Getenv("FEEDBOOK_STORE") {
	case "memory":
		store = feedbook.NewMemoryStore()
		accounts = newMemoryAccountStore()
	default:
		dbPath := os.Getenv("FEEDBOOK_DB_PATH")
		if dbPath == "" {
			dbPath = "feedbook.db"
		}
		s, err := feedbook.NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatalf("sqlite store: %v", err)
		}
		store = s
		accounts = s
	}

	service := feedbook.NewService(store)
	if sender, err := feedbook.NewFirebasePushSender(
		context.Background(),
		resolveFirebaseCredentialsFile(),
		resolveFirebaseProjectID(),
	); err != nil {
		log.Printf("Firebase push sender disabled: %v", err)
	} else {
		service.SetPushSender(sender)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/register", handleRegister(accounts))
	mux.HandleFunc("/login", handleLogin(accounts))
	mux.Handle("/api/", http.StripPrefix("/api", feedbookhttp.NewRouter(service)))

	addr := resolveAddr(os.Getenv("FEEDBOOK_ADDR"))

	server := &http.Server{
		Addr:              addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("FeedBook backend listening on http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped: %v", err)
	}
}

func resolveAddr(raw string) string {
	addr := strings.TrimSpace(raw)
	if addr == "" {
		return defaultAddr
	}
	return addr
}

func resolveFirebaseProjectID() string {
	if projectID := strings.TrimSpace(os.Getenv("FIREBASE_PROJECT_ID")); projectID != "" {
		return projectID
	}
	if projectID := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT")); projectID != "" {
		return projectID
	}
	return defaultFirebaseProjectID
}

func resolveFirebaseCredentialsFile() string {
	if credentialsFile := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_FILE")); credentialsFile != "" {
		return credentialsFile
	}
	if _, err := os.Stat(defaultFirebaseCredsFile); err == nil {
		return defaultFirebaseCredsFile
	}
	if matches, err := filepath.Glob("*firebase-adminsdk*.json"); err == nil && len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func handleRegister(accounts accountStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}

		var request loginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
			return
		}

		username, password, ok := normalizeCredentials(request)
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "username and password are required"})
			return
		}

		if len(password) < 4 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "password must be at least 4 characters"})
			return
		}

		created, err := accounts.CreateAccount(username, password)
		if err != nil {
			log.Printf("create account: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "unable to create account"})
			return
		}
		if !created {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "account already exists"})
			return
		}

		response, err := createLoginResponse(username, request.SecureLogin)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "unable to create token"})
			return
		}

		writeJSON(w, http.StatusCreated, response)
	}
}

func handleLogin(accounts accountStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}

		var request loginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
			return
		}

		username, password, ok := normalizeCredentials(request)
		if !ok {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "username and password are required"})
			return
		}

		expectedPassword, exists, err := accounts.AccountPassword(username)
		if err != nil {
			log.Printf("read account: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "unable to read account"})
			return
		}
		if !exists || expectedPassword != password {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}

		response, err := createLoginResponse(username, request.SecureLogin)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "unable to create token"})
			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func normalizeCredentials(request loginRequest) (string, string, bool) {
	username := strings.TrimSpace(request.Username)
	if username == "" {
		username = strings.TrimSpace(request.User)
	}
	password := strings.TrimSpace(request.Password)
	return username, password, username != "" && password != ""
}

func createLoginResponse(username string, secureLogin bool) (loginResponse, error) {
	expiresAt := time.Now().Add(30 * 24 * time.Hour).Unix()
	token, err := signJWT(map[string]any{
		"username":     username,
		"secure_login": secureLogin,
		"iat":          time.Now().Unix(),
		"exp":          expiresAt,
	})
	if err != nil {
		return loginResponse{}, err
	}
	return loginResponse{Token: token, Exp: expiresAt}, nil
}

func signJWT(claims map[string]any) (string, error) {
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	unsignedToken := header + "." + payload

	mac := hmac.New(sha256.New, []byte(jwtSecret))
	if _, err := mac.Write([]byte(unsignedToken)); err != nil {
		return "", err
	}

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return unsignedToken + "." + signature, nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}
