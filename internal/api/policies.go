package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/skylunna/luner/internal/storage"
)

type policyRequest struct {
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Expression  string `json:"expression"`
	Action      string `json:"action"`
	Priority    int    `json:"priority"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type policyResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Expression  string    `json:"expression"`
	Action      string    `json:"action"`
	Priority    int       `json:"priority"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func policyToResponse(p *storage.Policy) policyResponse {
	return policyResponse{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Name:        p.Name,
		Expression:  p.Expression,
		Action:      p.Action,
		Priority:    p.Priority,
		Description: p.Description,
		Enabled:     p.Enabled,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (s *RestServer) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := r.URL.Query().Get("tenant_id")

	var (
		policies []*storage.Policy
		err      error
	)
	if tenantID != "" {
		policies, err = s.store.Policies().ListByTenant(ctx, tenantID)
	} else {
		policies, err = s.store.Policies().ListActive(ctx)
	}
	if err != nil {
		s.errJSON(w, http.StatusInternalServerError, "failed to list policies")
		return
	}

	out := make([]policyResponse, 0, len(policies))
	for _, p := range policies {
		out = append(out, policyToResponse(p))
	}
	s.json(w, http.StatusOK, out)
}

func (s *RestServer) handleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req policyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Expression == "" || req.Action == "" {
		s.errJSON(w, http.StatusBadRequest, "name, expression and action are required")
		return
	}
	switch req.Action {
	case "block", "alert", "downgrade":
	default:
		s.errJSON(w, http.StatusBadRequest, "action must be one of: block, alert, downgrade")
		return
	}

	p := &storage.Policy{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		Name:        req.Name,
		Expression:  req.Expression,
		Action:      req.Action,
		Priority:    req.Priority,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	if err := s.store.Policies().Create(r.Context(), p); err != nil {
		s.logger.Error("create policy failed", "err", err)
		s.errJSON(w, http.StatusInternalServerError, "failed to create policy")
		return
	}

	if s.policyEngine != nil {
		s.policyEngine.Reload()
	}
	s.json(w, http.StatusCreated, policyToResponse(p))
}

func (s *RestServer) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.store.Policies().Get(r.Context(), id)
	if err == storage.ErrNotFound {
		s.errJSON(w, http.StatusNotFound, "policy not found")
		return
	}
	if err != nil {
		s.errJSON(w, http.StatusInternalServerError, "failed to get policy")
		return
	}
	s.json(w, http.StatusOK, policyToResponse(p))
}

func (s *RestServer) handleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req policyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := s.store.Policies().Get(r.Context(), id)
	if err == storage.ErrNotFound {
		s.errJSON(w, http.StatusNotFound, "policy not found")
		return
	}
	if err != nil {
		s.errJSON(w, http.StatusInternalServerError, "failed to get policy")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Expression != "" {
		existing.Expression = req.Expression
	}
	if req.Action != "" {
		existing.Action = req.Action
	}
	existing.Priority = req.Priority
	existing.Description = req.Description
	existing.Enabled = req.Enabled

	if err := s.store.Policies().Update(r.Context(), existing); err != nil {
		s.logger.Error("update policy failed", "err", err)
		s.errJSON(w, http.StatusInternalServerError, "failed to update policy")
		return
	}

	if s.policyEngine != nil {
		s.policyEngine.Reload()
	}
	s.json(w, http.StatusOK, policyToResponse(existing))
}

func (s *RestServer) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.Policies().Delete(r.Context(), id); err == storage.ErrNotFound {
		s.errJSON(w, http.StatusNotFound, "policy not found")
		return
	} else if err != nil {
		s.errJSON(w, http.StatusInternalServerError, "failed to delete policy")
		return
	}

	if s.policyEngine != nil {
		s.policyEngine.Reload()
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *RestServer) handleReloadPolicies(w http.ResponseWriter, r *http.Request) {
	if s.policyEngine != nil {
		s.policyEngine.Reload()
	}
	s.json(w, http.StatusOK, map[string]string{"status": "reloaded"})
}
