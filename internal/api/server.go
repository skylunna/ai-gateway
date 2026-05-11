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

	// Always available — no storage required
	s.mux.Handle("GET /api/health", wrap(s.handleHealth))
	s.mux.Handle("GET /api/metrics/live", wrap(s.handleLiveMetrics))

	// Storage-dependent endpoints
	s.mux.Handle("GET /api/traces", wrap(s.requireStore(s.handleListTraces)))
	s.mux.Handle("GET /api/traces/{trace_id}", wrap(s.requireStore(s.handleGetTrace)))
	s.mux.Handle("GET /api/dashboard/summary", wrap(s.requireStore(s.handleDashboardSummary)))
	s.mux.Handle("GET /api/dashboard/cost", wrap(s.requireStore(s.handleCostBreakdown)))

	// Policy CRUD (storage-dependent)
	s.mux.Handle("GET /api/policies", wrap(s.requireStore(s.handleListPolicies)))
	s.mux.Handle("POST /api/policies", wrap(s.requireStore(s.handleCreatePolicy)))
	s.mux.Handle("GET /api/policies/{id}", wrap(s.requireStore(s.handleGetPolicy)))
	s.mux.Handle("PUT /api/policies/{id}", wrap(s.requireStore(s.handleUpdatePolicy)))
	s.mux.Handle("DELETE /api/policies/{id}", wrap(s.requireStore(s.handleDeletePolicy)))
	s.mux.Handle("POST /api/policies/reload", wrap(s.requireStore(s.handleReloadPolicies)))

	// OPTIONS preflight for all /api/ paths
	s.mux.Handle("OPTIONS /api/", cors(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
}

// requireStore wraps a handler and returns 503 when storage is not configured.
func (s *RestServer) requireStore(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.store == nil {
			s.errJSON(w, http.StatusServiceUnavailable, "storage not configured")
			return
		}
		h(w, r)
	}
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
