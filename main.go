package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	feedbook "github.com/feedbook/back/internal/feedbook"
	feedbookhttp "github.com/feedbook/back/internal/feedbook/http"
)

const (
	defaultAddr = "127.0.0.1:8080"
	jwtSecret   = "feedbook-local-secret"
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", handleLogin)
	mux.Handle("/api/", http.StripPrefix("/api", feedbookhttp.NewRouter(feedbook.NewService(feedbook.NewStore()))))

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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	username := strings.TrimSpace(request.Username)
	if username == "" {
		username = strings.TrimSpace(request.User)
	}
	password := strings.TrimSpace(request.Password)

	if username == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "username and password are required"})
		return
	}

	if username != password {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		return
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour).Unix()
	token, err := signJWT(map[string]any{
		"username":     username,
		"password":     password,
		"secure_login": request.SecureLogin,
		"iat":          time.Now().Unix(),
		"exp":          expiresAt,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "unable to create token"})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token, Exp: expiresAt})
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
