package feedbookhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	feedbook "github.com/feedbook/back/internal/feedbook"
)

func newTestRouter() http.Handler {
	return NewRouter(feedbook.NewService(feedbook.NewMemoryStore()))
}

func newPushTestRouter(sender feedbook.PushSender) http.Handler {
	service := feedbook.NewService(feedbook.NewMemoryStore())
	service.SetPushSender(sender)
	return NewRouter(service)
}

func TestBooksEndpointReturnsData(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var books []feedbook.Book
	if err := json.NewDecoder(rec.Body).Decode(&books); err != nil {
		t.Fatalf("decode books: %v", err)
	}
	if len(books) == 0 {
		t.Fatal("expected at least one book")
	}
}

func TestProfileUpdateRoundTrip(t *testing.T) {
	payload := feedbook.UpdateProfileRequest{
		Name:                 "Backend Reader",
		Handle:               "@backendreader",
		Quote:                "Updated from test",
		AvatarTopColorHex:    1,
		AvatarBottomColorHex: 2,
		TargetPagesPerDay:    intPtr(55),
	}
	body, _ := json.Marshal(payload)

	router := newTestRouter()
	updateReq := httptest.NewRequest(http.MethodPut, "/profile/me", bytes.NewReader(body))
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d", updateRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/profile/me", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	var profile feedbook.Profile
	if err := json.NewDecoder(getRec.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if profile.Name != payload.Name || profile.Handle != payload.Handle {
		t.Fatalf("expected updated profile, got %+v", profile)
	}
	if profile.ReadingGoal == nil || profile.ReadingGoal.TargetPagesPerDay != 55 {
		t.Fatalf("expected updated reading goal, got %+v", profile.ReadingGoal)
	}
}

func TestProfilePreviewReflectsUpdatedProfile(t *testing.T) {
	payload := feedbook.UpdateProfileRequest{
		Name:                 "Preview Reader",
		Handle:               "@previewreader",
		Quote:                "Preview updated",
		AvatarTopColorHex:    10,
		AvatarBottomColorHex: 20,
		TargetPagesPerDay:    intPtr(35),
	}
	body, _ := json.Marshal(payload)

	router := newTestRouter()
	updateReq := httptest.NewRequest(http.MethodPut, "/profile/me", bytes.NewReader(body))
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d", updateRec.Code)
	}

	previewReq := httptest.NewRequest(http.MethodGet, "/profile/me/preview", nil)
	previewRec := httptest.NewRecorder()
	router.ServeHTTP(previewRec, previewReq)

	if previewRec.Code != http.StatusOK {
		t.Fatalf("expected preview status 200, got %d", previewRec.Code)
	}

	var preview feedbook.Profile
	if err := json.NewDecoder(previewRec.Body).Decode(&preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if preview.Name != payload.Name || preview.Handle != payload.Handle {
		t.Fatalf("expected preview to reflect updated profile, got %+v", preview)
	}
}

func TestPublicProfileEndpointReturnsSnapshot(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/profile/public", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var profile feedbook.Profile
	if err := json.NewDecoder(rec.Body).Decode(&profile); err != nil {
		t.Fatalf("decode public profile: %v", err)
	}
	if profile.Handle == "" {
		t.Fatal("expected public profile handle")
	}
}

func TestMissingBookReturnsNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/books/unknown", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestPushRegisterAcceptsToken(t *testing.T) {
	body := bytes.NewReader([]byte(`{"token":"fcm-token","platform":"android"}`))
	req := httptest.NewRequest(http.MethodPost, "/push/register", body)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestPushSendUsesRegisteredTokens(t *testing.T) {
	sender := &fakePushSender{}
	router := newPushTestRouter(sender)

	registerReq := httptest.NewRequest(
		http.MethodPost,
		"/push/register",
		bytes.NewReader([]byte(`{"token":"fcm-token","platform":"android"}`)),
	)
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusNoContent {
		t.Fatalf("expected register status 204, got %d", registerRec.Code)
	}

	sendReq := httptest.NewRequest(
		http.MethodPost,
		"/push/send",
		bytes.NewReader([]byte(`{"title":"FeedBook","body":"Nueva actividad"}`)),
	)
	sendRec := httptest.NewRecorder()
	router.ServeHTTP(sendRec, sendReq)

	if sendRec.Code != http.StatusOK {
		t.Fatalf("expected send status 200, got %d", sendRec.Code)
	}
	if sender.token != "fcm-token" || sender.title != "FeedBook" || sender.body != "Nueva actividad" {
		t.Fatalf("unexpected push send call: %+v", sender)
	}

	var response feedbook.SendPushResponse
	if err := json.NewDecoder(sendRec.Body).Decode(&response); err != nil {
		t.Fatalf("decode push response: %v", err)
	}
	if response.Sent != 1 || response.Failed != 0 || len(response.IDs) != 1 {
		t.Fatalf("unexpected push response: %+v", response)
	}
}

func TestPushTokensReturnsRegisteredTokens(t *testing.T) {
	router := newTestRouter()
	registerReq := httptest.NewRequest(
		http.MethodPost,
		"/push/register",
		bytes.NewReader([]byte(`{"token":"fcm-token","platform":"android"}`)),
	)
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusNoContent {
		t.Fatalf("expected register status 204, got %d", registerRec.Code)
	}

	tokensReq := httptest.NewRequest(http.MethodGet, "/push/tokens", nil)
	tokensRec := httptest.NewRecorder()
	router.ServeHTTP(tokensRec, tokensReq)

	if tokensRec.Code != http.StatusOK {
		t.Fatalf("expected tokens status 200, got %d", tokensRec.Code)
	}

	var response struct {
		Count  int                      `json:"count"`
		Tokens []feedbook.PushTokenInfo `json:"tokens"`
	}
	if err := json.NewDecoder(tokensRec.Body).Decode(&response); err != nil {
		t.Fatalf("decode tokens response: %v", err)
	}
	if response.Count != 1 || len(response.Tokens) != 1 || response.Tokens[0].Token != "fcm-token" {
		t.Fatalf("unexpected tokens response: %+v", response)
	}
}

func intPtr(v int) *int {
	return &v
}

type fakePushSender struct {
	token string
	title string
	body  string
}

func (s *fakePushSender) Send(token string, title string, body string, data map[string]string) (string, error) {
	s.token = token
	s.title = title
	s.body = body
	return "projects/feedbook/messages/test", nil
}
