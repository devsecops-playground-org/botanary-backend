package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devsecops-playground-org/botanary-backend/internal/catalog"
)

func newTestServer() http.Handler {
	return NewServer(catalog.NewSeeded(), nil).Routes()
}

func do(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	rec := httptest.NewRecorder()
	newTestServer().ServeHTTP(rec, req)
	return rec
}

func TestHealthIsOK(t *testing.T) {
	rec := do(t, http.MethodGet, "/health", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("health did not return JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("got status %q, want ok", body["status"])
	}
}

func TestSecurityHeadersArePresent(t *testing.T) {
	rec := do(t, http.MethodGet, "/health", "")

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("got %q, want nosniff", got)
	}
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("got %q, want DENY", got)
	}
}

func TestListListings(t *testing.T) {
	rec := do(t, http.MethodGet, "/api/listings", "")

	var listings []catalog.Listing
	if err := json.Unmarshal(rec.Body.Bytes(), &listings); err != nil {
		t.Fatalf("unexpected body: %v", err)
	}
	if len(listings) != 3 {
		t.Fatalf("got %d listings, want 3", len(listings))
	}
}

func TestListListingsFilteredToInStock(t *testing.T) {
	rec := do(t, http.MethodGet, "/api/listings?in_stock=true", "")

	var listings []catalog.Listing
	_ = json.Unmarshal(rec.Body.Bytes(), &listings)
	for _, l := range listings {
		if !l.InStock {
			t.Fatalf("out-of-stock listing %q leaked into the filtered result", l.Name)
		}
	}
	if len(listings) != 2 {
		t.Fatalf("got %d in-stock listings, want 2", len(listings))
	}
}

func TestGetListing(t *testing.T) {
	if rec := do(t, http.MethodGet, "/api/listings/1", ""); rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	if rec := do(t, http.MethodGet, "/api/listings/999", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}
	if rec := do(t, http.MethodGet, "/api/listings/abc", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
}

func TestCreateListing(t *testing.T) {
	rec := do(t, http.MethodPost, "/api/listings", `{"name":"Pothos","species":"epipremnum","price_cts":1500,"in_stock":true}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want 201: %s", rec.Code, rec.Body.String())
	}
	var created catalog.Listing
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.ID == 0 {
		t.Fatal("created listing has no ID")
	}
}

func TestCreateListingRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"not json", `{`, http.StatusBadRequest},
		{"unknown field", `{"name":"x","sneaky":true}`, http.StatusBadRequest},
		{"empty name", `{"name":"  "}`, http.StatusUnprocessableEntity},
		{"negative price", `{"name":"x","price_cts":-5}`, http.StatusUnprocessableEntity},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if rec := do(t, http.MethodPost, "/api/listings", tc.body); rec.Code != tc.want {
				t.Fatalf("got %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
