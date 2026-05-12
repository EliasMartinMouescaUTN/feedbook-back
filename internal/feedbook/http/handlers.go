package feedbookhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	feedbook "github.com/feedbook/back/internal/feedbook"
)

type Handler struct {
	service *feedbook.Service
}

func (h *Handler) handleBooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetBooks())
}

func (h *Handler) handleBookRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/books/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusNotFound, feedbook.ErrorResponse{Error: "not found"})
		return
	}
	bookID := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		book, err := h.service.GetBookByID(bookID)
		writeServiceResponse(w, book, err)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "progress":
			if r.Method == http.MethodGet {
				progress, err := h.service.GetReadingProgress(bookID)
				writeServiceResponse(w, progress, err)
				return
			}
			if r.Method == http.MethodPut {
				var req struct {
					CurrentPage int `json:"current_page"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeJSON(w, http.StatusBadRequest, feedbook.ErrorResponse{Error: "invalid json body"})
					return
				}
				progress, err := h.service.SaveReadingProgress(bookID, req.CurrentPage)
				writeServiceResponse(w, progress, err)
				return
			}
		case "reviews":
			switch r.Method {
			case http.MethodGet:
				writeJSON(w, http.StatusOK, h.service.GetReviews(bookID))
				return
			case http.MethodPost:
				var req struct {
					Rating float32 `json:"rating"`
					Text   string  `json:"text"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeJSON(w, http.StatusBadRequest, feedbook.ErrorResponse{Error: "invalid json body"})
					return
				}
				review, err := h.service.SaveReview(bookID, req.Rating, req.Text)
				writeServiceResponse(w, review, err)
				return
			}
		}
	}
	writeJSON(w, http.StatusNotFound, feedbook.ErrorResponse{Error: "not found"})
}

func (h *Handler) handleExploreUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetExploreUsers())
}

func (h *Handler) handleAuthors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetAuthors())
}

func (h *Handler) handleAuthorRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/authors/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusNotFound, feedbook.ErrorResponse{Error: "not found"})
		return
	}
	authorID := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		author, err := h.service.GetAuthorByID(authorID)
		writeServiceResponse(w, author, err)
		return
	}
	if len(parts) == 2 && parts[1] == "follow-toggle" && r.Method == http.MethodPost {
		err := h.service.ToggleAuthorFollow(authorID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusNotFound, feedbook.ErrorResponse{Error: "not found"})
}

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetHome())
}

func (h *Handler) handleOwnLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetOwnLibrary())
}

func (h *Handler) handleOwnProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.service.GetOwnProfile())
	case http.MethodPut:
		var request feedbook.UpdateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, feedbook.ErrorResponse{Error: "invalid json body"})
			return
		}
		profile, err := h.service.UpdateOwnProfile(request)
		writeServiceResponse(w, profile, err)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
	}
}

func (h *Handler) handleLibraryBooks(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID string `json:"book_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, feedbook.ErrorResponse{Error: "invalid json body"})
		return
	}
	if req.BookID == "" {
		writeJSON(w, http.StatusBadRequest, feedbook.ErrorResponse{Error: "book_id is required"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		err := h.service.AddBookToLibrary(req.BookID)
		if errors.Is(err, feedbook.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, feedbook.ErrorResponse{Error: "book not found"})
			return
		}
		if errors.Is(err, feedbook.ErrAlreadyInLibrary) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, feedbook.ErrorResponse{Error: "internal server error"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		err := h.service.RemoveBookFromLibrary(req.BookID)
		if errors.Is(err, feedbook.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, feedbook.ErrorResponse{Error: "book not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, feedbook.ErrorResponse{Error: "internal server error"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
	}
}

func (h *Handler) handleOwnPublicPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetOwnPublicPreview())
}

func (h *Handler) handlePublicProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetPublicProfile())
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetStats())
}

func (h *Handler) handleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, feedbook.ErrorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, h.service.GetNotifications())
}

func writeServiceResponse(w http.ResponseWriter, payload any, err error) {
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, feedbook.ErrNotFound):
		writeJSON(w, http.StatusNotFound, feedbook.ErrorResponse{Error: "not found"})
	case errors.Is(err, feedbook.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, feedbook.ErrorResponse{Error: "invalid request"})
	default:
		writeJSON(w, http.StatusInternalServerError, feedbook.ErrorResponse{Error: "internal server error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
