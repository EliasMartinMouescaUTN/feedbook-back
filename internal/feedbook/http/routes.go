package feedbookhttp

import (
	"net/http"

	feedbook "github.com/feedbook/back/internal/feedbook"
)

func NewRouter(service *feedbook.Service) http.Handler {
	handler := &Handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("/books", handler.handleBooks)
	mux.HandleFunc("/books/", handler.handleBookRoutes)
	mux.HandleFunc("/explore/users", handler.handleExploreUsers)
	mux.HandleFunc("/authors", handler.handleAuthors)
	mux.HandleFunc("/authors/", handler.handleAuthorRoutes)
	mux.HandleFunc("/home", handler.handleHome)
	mux.HandleFunc("/library/me", handler.handleOwnLibrary)
	mux.HandleFunc("/library/me/books", handler.handleLibraryBooks)
	mux.HandleFunc("/profile/me", handler.handleOwnProfile)
	mux.HandleFunc("/profile/me/preview", handler.handleOwnPublicPreview)
	mux.HandleFunc("/profile/public", handler.handlePublicProfile)
	mux.HandleFunc("/stats", handler.handleStats)
	mux.HandleFunc("/notifications", handler.handleNotifications)
	mux.HandleFunc("/push/register", handler.handlePushRegister)
	mux.HandleFunc("/push/send", handler.handlePushSend)
	mux.HandleFunc("/push/tokens", handler.handlePushTokens)
	return mux
}
