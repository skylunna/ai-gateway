package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/skylunna/luner/internal/policy"
	"github.com/skylunna/luner/internal/storage"
)

// RestServer is the Web Console REST API server.
type RestServer struct {
	mux          *http.ServeMux
	store        storage.Storage
	logger       *slog.Logger
	policyEngine policy.Engine
}

// NewRestServer wires up all REST routes.
func NewRestServer(store storage.Storage, logger *slog.Logger) *RestServer {
	return NewRestServerWithEngine(store, logger, nil)
}

// NewRestServerWithEngine wires up REST routes with an optional policy engine.
func NewRestServerWithEngine(store storage.Storage, logger *slog.Logger, engine policy.Engine) *RestServer {
	s := &RestServer{
		mux:          http.NewServeMux(),
		store:        store,
		logger:       logger,
		policyEngine: engine,
	}
	s.routes()
	return s
}

func (s *RestServer) routes() {
	wrap := func(h http.HandlerFunc) http.Handler { return cors(h) }

	s.mux.Handle("GET /api/health", wrap(s.handleHealth))
	s.mux.Handle("GET /api/traces", wrap(s.handleListTraces))
	s.mux.Handle("GET /api/traces/{trace_id}", wrap(s.handleGetTrace))
	s.mux.Handle("GET /api/dashboard/summary", wrap(s.handleDashboardSummary))
	s.mux.Handle("GET /api/dashboard/cost", wrap(s.handleCostBreakdown))

	// Policy CRUD
	s.mux.Handle("GET /api/policies", wrap(s.handleListPolicies))
	s.mux.Handle("POST /api/policies", wrap(s.handleCreatePolicy))
	s.mux.Handle("GET /api/policies/{id}", wrap(s.handleGetPolicy))
	s.mux.Handle("PUT /api/policies/{id}", wrap(s.handleUpdatePolicy))
	s.mux.Handle("DELETE /api/policies/{id}", wrap(s.handleDeletePolicy))
	s.mux.Handle("POST /api/policies/reload", wrap(s.handleReloadPolicies))

	// OPTIONS preflight for all /api/ paths
	s.mux.Handle("OPTIONS /api/", cors(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
}

func (s *RestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// cors wraps a handler with permissive CORS headers for the Vite dev server.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
}

func (s *RestServer) json(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *RestServer) errJSON(w http.ResponseWriter, status int, msg string) {
	s.json(w, status, map[string]string{"error": msg})
}
