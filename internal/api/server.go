// Package api exposes the catalog over HTTP.
package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/devsecops-playground-org/botanary-backend/internal/catalog"
)

const version = "0.1.0"

type Server struct {
	catalog *catalog.Catalog
	logger  *slog.Logger
}

func NewServer(c *catalog.Catalog, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{catalog: c, logger: logger}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/listings", s.listListings)
	mux.HandleFunc("POST /api/listings", s.createListing)
	mux.HandleFunc("GET /api/listings/{id}", s.getListing)
	return securityHeaders(mux)
}

// securityHeaders applies the baseline every service in the organisation sets.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": version})
}

func (s *Server) listListings(w http.ResponseWriter, r *http.Request) {
	inStockOnly := strings.EqualFold(r.URL.Query().Get("in_stock"), "true")
	writeJSON(w, http.StatusOK, s.catalog.List(inStockOnly))
}

func (s *Server) getListing(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be an integer")
		return
	}

	listing, err := s.catalog.Get(id)
	if errors.Is(err, catalog.ErrNotFound) {
		writeError(w, http.StatusNotFound, "listing not found")
		return
	}
	writeJSON(w, http.StatusOK, listing)
}

func (s *Server) createListing(w http.ResponseWriter, r *http.Request) {
	var payload catalog.Listing

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "body is not a valid listing")
		return
	}

	created, err := s.catalog.Add(payload)
	if errors.Is(err, catalog.ErrInvalid) {
		writeError(w, http.StatusUnprocessableEntity, "name is required and price must not be negative")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Default().Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
